module github.com/jiahao42/fensor

require (
	github.com/golang/mock v1.2.0
	github.com/golang/protobuf v1.3.5
	github.com/google/go-cmp v0.2.0
	github.com/gorilla/websocket v1.4.1
	github.com/miekg/dns v1.1.4
	github.com/refraction-networking/utls v0.0.0-20190909200633-43c36d3c1f57
	go.starlark.net v0.0.0-20190919145610-979af19b165c
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190311183353-d8887717615a
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a
	google.golang.org/genproto v0.0.0-20180831171423-11092d34479b // indirect
	google.golang.org/grpc v1.24.0
	h12.io/socks v1.0.0
)

replace (
	v2ray.com/core v4.19.1+incompatible => .
)

go 1.14
