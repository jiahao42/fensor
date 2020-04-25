module v2ray.com/core

require (
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.3.5
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-cmp v0.2.0
	github.com/gorilla/websocket v1.4.1
	github.com/lucas-clemente/quic-go v0.15.2
	github.com/miekg/dns v1.1.4
	github.com/refraction-networking/utls v0.0.0-20190909200633-43c36d3c1f57
	go.starlark.net v0.0.0-20190919145610-979af19b165c
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd
	google.golang.org/grpc v1.24.0
	h12.io/socks v1.0.0
)

replace v2ray.com/core v4.19.1+incompatible => ./

go 1.14
