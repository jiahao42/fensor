# fensor

"fensor" means "f\*\*k censor(ship)", it's a tool for circumventing censorship.

This project is based on [v2ray](https://github.com/v2ray/v2ray-core), and the modifications are mainly guided by the paper [Incentivizing censorship measurements via circumvention](https://dl.acm.org/doi/abs/10.1145/3230543.3230568) from SIGCOMM'18.


## Modifications

### On mechanism - adaptive circumvention on the fly 

Every URL will be stored as a tuple `(URL, status)` in a Redis database, fensor will choose different proxy protocol based on the URL status.

| URL status| Protocol| 
| ------------- |:-------------:|
| DNS blocked| [Freedom](https://v2ray.com/en/configuration/protocols/freedom.html) |
| TCP conn. blocked/reset| [Shadowsocks](https://v2ray.com/en/configuration/protocols/shadowsocks.html)/[VMess](https://v2ray.com/en/configuration/protocols/vmess.html) |
| Wrong/Blank webpage returned| [Shadowsocks](https://v2ray.com/en/configuration/protocols/shadowsocks.html)/[VMess](https://v2ray.com/en/configuration/protocols/vmess.html) |

Specifically, each v2ray instance uses one of its proxy protocols (e.g., Freedom and VMess).

![Freedom](./doc/img/v2ray_freedom.png)
![VMess](./doc/img/v2ray_vmess.png)

Different from v2ray, fensor will choose between proxy protocols dynamically, based on the status of each URL stored in the database.

![fensor](./doc/img/fensor.png)

### On protocols 

<!--* freedom: add global DNS servers, i.e., when there is no valid response from the local DNS server, it shall turn to -->


## Development

### Playground

1. Make sure you have golang install on your computer, and your `GOPATH` is set properly.
2. Pull the code by `go get -u github.com/jiahao42/fensor`
3. Run `fensor/playground/build.sh`, and you shall see two executables: `v2ray` and `v2ctl` under `fensor/playground`. 
4. Run `fensor/playground/run_{protocol}.sh`, `v2ray` will run as both client and server separately on your computer with the default config file (e.g., `vmess_client.json` and `vmess_server.json`). Check the status of client and server by using `screen -r v2ray`.
5. Set your proxy properly (e.g. in your browser) and you are ready to go. Note that in Firefox, you have to check `Proxy DNS when using SOCKS v5`, as we need to do DNS lookup in fensor instead of browser. 

### Test

To test the whole project, run `go test ./...` under the root directory.
