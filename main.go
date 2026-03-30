package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/txthinking/socks5"
)

var (
	proxyUsername = flag.String("u", "", "proxy username")
	proxyPassword = flag.String("p", "", "proxy password")
	proxyHost     = flag.String("h", "socks.pr.oxylabs.io:7777", "proxy host with port, i.e. host:port")
	remoteHost    = flag.String("t", "echo-udp.oxylabs.io", "target host")
	payload       = flag.String("body", "", "optional request body")
)

type SOCKS5PacketConn struct {
	net.Conn
	targetAddr net.Addr
}

func (s *SOCKS5PacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	n, err := s.Conn.Read(p)
	return n, s.targetAddr, err
}

func (s *SOCKS5PacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	return s.Conn.Write(p)
}

func (s *SOCKS5PacketConn) LocalAddr() net.Addr {
	return s.Conn.LocalAddr()
}

func (s *SOCKS5PacketConn) SetDeadline(t time.Time) error {
	return s.Conn.SetDeadline(t)
}

func (s *SOCKS5PacketConn) SetReadBuffer(bytes int) error {
	if udpConn, ok := s.Conn.(interface{ SetReadBuffer(int) error }); ok {
		return udpConn.SetReadBuffer(bytes)
	}
	return nil
}

func (s *SOCKS5PacketConn) SetWriteBuffer(bytes int) error {
	if udpConn, ok := s.Conn.(interface{ SetReadBuffer(int) error }); ok {
		return udpConn.SetReadBuffer(bytes)
	}
	return nil
}

func HTTP3viaProxy(proxyUsername, proxyPassword, proxyHost string, target string, payload string) error {
	socks5Client, err := socks5.NewClient(proxyHost, proxyUsername, proxyPassword, 5, 5)
	if err != nil {
		return err
	}
	closeChan := make(chan struct{})
	defer func() {
		closeChan <- struct{}{}
	}()
	client := http.Client{
		Transport: &http3.Transport{
			Dial: proxyDialer(closeChan, socks5Client),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		"https://"+target,
		bytes.NewReader([]byte(payload)),
	)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("Status: ", resp.StatusCode)
	fmt.Println("Protocol: ", resp.Proto)
	fmt.Printf("Response: \n%s", string(res))

	return nil
}

func proxyDialer(closeChan <-chan struct{}, socks5client *socks5.Client) func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
	return func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
		proxyConn, err := socks5client.Dial("udp", addr)
		if err != nil {
			return nil, err
		}
		go func() {
			<-closeChan
			if proxyConn != nil {
				err := proxyConn.Close()
				if err != nil {
					slog.Error("Failed to close proxy connection due to: ", "error", err)
				}
			}
		}()

		remoteAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return nil, err
		}
		socks5PacketConn := &SOCKS5PacketConn{
			Conn:       proxyConn,
			targetAddr: remoteAddr,
		}
		deadline := time.Now().Add(40 * time.Second)
		err = proxyConn.SetReadDeadline(deadline)
		if err != nil {
			return nil, err
		}
		err = proxyConn.SetWriteDeadline(deadline)
		if err != nil {
			return nil, err
		}

		earlyConn, err := quic.DialEarly(ctx, socks5PacketConn, remoteAddr, tlsCfg, cfg)
		if err != nil {
			return nil, err
		}

		return earlyConn, nil
	}
}

func main() {
	flag.Parse()
	start := time.Now()
	err := HTTP3viaProxy(*proxyUsername, *proxyPassword, *proxyHost, *remoteHost, *payload)
	if err != nil {
		slog.Error("failed to perform HTTP3 request via proxy", "error", err)
	}

	fmt.Printf("Request took: %d ms\n", time.Since(start).Milliseconds())
}
