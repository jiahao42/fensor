# fensor

"fensor" means "f\*\*k censor(ship)", it's a tool for circumventing censorship.

This project is based on [v2ray](https://github.com/v2ray/v2ray-core), and the modifications are mainly guided by the paper [Incentivizing censorship measurements via circumvention](https://dl.acm.org/doi/abs/10.1145/3230543.3230568) from SIGCOMM'18.


## Modifications

### On mechanism - adaptive circumvention on the fly 

Every URL will be stored as a tuple `(URL, status)` in a Redis database, fensor will choose different proxy protocol based on the URL status.

| URL status| Protocol| 
| ------------- |:-------------:|
| DNS blocked| [Freedom](https://v2ray.com/en/configuration/protocols/freedom.html) |
| TCP conn. blocked/reset| [Shadowsocks](https://v2ray.com/en/configuration/protocols/shadowsocks.html)/[Vmess](https://v2ray.com/en/configuration/protocols/vmess.html) |
| Wrong/Blank webpage returned| [Shadowsocks](https://v2ray.com/en/configuration/protocols/shadowsocks.html)/[Vmess](https://v2ray.com/en/configuration/protocols/vmess.html) |

### On protocols 

<!--* freedom: add global DNS servers, i.e., when there is no valid response from the local DNS server, it shall turn to -->


## Development

### Playground

1. Make sure you have golang install on your computer, and your `GOPATH` is set properly.
2. Pull the code by `go get -u github.com/jiahao42/fensor`
3. Run `fensor/playground/build.sh`, and you shall see two executables: `v2ray` and `v2ctl` under `fensor/playground`. 
4. Run `fensor/playground/run_{protocol}.sh`, `v2ray` will run as both client and server separately on your computer with the default config file (e.g., `vmess_inbound.json` and `vmess_outbound.json`). Check the status of client and server by using `screen -r v2ray`.
5. Set your proxy properly (e.g. in your browser) and you are ready to go.

### Test

To test the whole project, run `go test ./...` under the root directory.
