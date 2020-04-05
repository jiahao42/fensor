"fensor" means "f\*\*k censor(ship)"

This project is based on [v2ray](https://github.com/v2ray/v2ray-core), and the modifications are mainly guided by the paper [Incentivizing censorship measurements via circumvention](https://dl.acm.org/doi/abs/10.1145/3230543.3230568) from SIGCOMM'18.


### Playground for development

1. Make sure you have golang install on your computer, and your `GOPATH` is set properly.
2. Pull the code by `go get -u github.com/jiahao42/fensor`
3. Run `fensor/playground/build.sh`, and you shall see two executables: `v2ray` and `v2ctl` under `fensor/playground`. 
4. Run `fensor/playground/run.sh`, `v2ray` will run as both client and server on your computer with the default config file (`inbound.json` and `outbound.json`). Check the status of client and server by using `screen -r v2ray`.
5. Set your proxy properly (e.g. in your browser) and you are ready to go.

