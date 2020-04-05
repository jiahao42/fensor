#!/bin/bash

cd `go env GOPATH`/src/github.com/jiahao42/fensor/main
env CGO_ENABLED=0 go build -o `go env GOPATH`/src/github.com/jiahao42/fensor/playground/v2ray -ldflags "-s -w"
cd `go env GOPATH`/src/github.com/jiahao42/fensor/infra/control/main
env CGO_ENABLED=0 go build -o `go env GOPATH`/src/github.com/jiahao42/fensor/playground/v2ctl -tags confonly -ldflags "-s -w"


