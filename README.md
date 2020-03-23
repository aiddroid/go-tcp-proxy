### TCP-Proxy
A simple proxy that allow you serve TCP proxy from a port to another.

### features
- TCP proxy
- IP white list
- daemon mode
- logging

#### usage
```
Start TCP proxy from a port to another.

Usage:
  tcp-proxy [command]

Available Commands:
  help        Help about any command
  start       Start TCP proxy

Flags:
  -c, --config string    config file (default is $HOME/.tcp-proxy.yaml)
  -h, --help             help for tcp-proxy
  -l, --logfile string   log file path (default is STDOUT)

```
Example: Proxy MYSQL at port 80

`sudo ./tcp-proxy start -f 80 -t 3306`
