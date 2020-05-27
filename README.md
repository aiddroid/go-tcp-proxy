### TCP-Proxy
A simple proxy that allow you serve TCP proxy from a port to another.

### features
- TCP port proxy
- IP white list
- logging

#### usage
```
Start TCP proxy from a port to another.

Usage:
  tcp-proxy [command]

Available Commands:
  start       Start TCP proxy
  help        Help about

Flags:
  -c, --config string    config file (default is $HOME/.tcp-proxy.yaml)
  -l, --logfile string   log file path (default is STDOUT)
  -h, --help             help for tcp-proxy

```
Example: Proxy MYSQL at port 80:
`sudo ./tcp-proxy start -f 80 -t 3306 -w whiteip.json -l proxy.log`

Auth current IP to white IP list:
`curl -XPOST http://your-server-ip:port/auth/RANDOM-AUTH-URI`
