module v2ray.com/core

require (
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.3.5
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-cmp v0.4.0
	github.com/gorilla/websocket v1.4.2
	github.com/jiahao42/fensor v0.0.0-20200406200920-45f705057bfd
	github.com/miekg/dns v1.1.4
	github.com/refraction-networking/utls v0.0.0-20190909200633-43c36d3c1f57
	go.starlark.net v0.0.0-20200330013621-be5394c419b6
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20200406155108-e3b113bbe6a4
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20200406120821-33397c535dc2 // indirect
	google.golang.org/grpc v1.28.1
	h12.io/socks v1.0.0
)

replace v2ray.com/core v4.19.1+incompatible => ./

go 1.14
