# HTTP3 via SOCKS5 UDP Example

An example on how to do an HTTP3 request via a SOCKS5 UDP proxy.

HTTP/3 uses UDP instead of TCP at the transport layer, 
which requires SOCKS5 UDP support, a feature not widely available 
in standard HTTP client libraries. This tool bridges that gap by:

- Implementing SOCKS5 UDP association
- Handling QUIC/HTTP3 connections over the proxy
- Providing a simple CLI interface


## Usage
```
  -body string
        optional request body
  -h string
        proxy host with port, i.e. host:port (default "socks.pr.oxylabs.io:7777")
  -p string
        proxy password
  -t string
        target host (default "echo-udp.oxylabs.io")
  -u string
        proxy username
```

i.e.
```bash
./http3viasocks5udp -h <proxy-host:port> -u <username> -p <password> -t <target-host>
```

real params example:
```bash
./http3viasocks5udp -h socks.pr.oxylabs.io:7777 -u myproxyusername -p myproxypassword -t echo-udp.oxylabs.io -body "Hello HTTP3"
```

## Installation

Clone and build the tool:

```bash
git clone https://github.com/oxylabs/gohttp3viaSOCKS5UDP.git
cd gohttp3viaSOCKS5UDP
go build -o http3viasocks5udp .
```
Or download a pre-built binary from the [releases page](https://github.com/oxylabs/gohttp3viaSOCKS5UDP/releases/latest).

## Integration

For programmatic usage, refer to `main.go` for implementation details on:

- Setting up SOCKS5 UDP dialers
- Configuring HTTP/3 transport
- Handling authentication