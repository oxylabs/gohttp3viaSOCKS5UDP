# HTTP3 via SOCKS5 UDP Example

An example on how to do an HTTP3 request via a SOCKS5 UDP proxy.

Usage
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

Full params example:
```
# If you want to build from source (need Go 1.25+)
go build -o http3viasocks5udp .
# Do a request:
./http3viasocks5udp -h socks.pr.oxylabs.io:7777 -u myproxyusername -p myproxypassword -t echo-udp.oxylabs.io -body "Hello HTTP3"
```