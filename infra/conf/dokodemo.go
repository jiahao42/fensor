package conf

import (
	"github.com/golang/protobuf/proto"
	"v2ray.com/core/proxy/dokodemo"
)

type DokodemoConfig struct {
	Host           *Address     `json:"address"`
	PortValue      uint16       `json:"port"`
	RelayPortValue uint16       `json:"relayport"`
	NetworkList    *NetworkList `json:"network"`
	TimeoutValue   uint32       `json:"timeout"`
	Redirect       bool         `json:"followRedirect"`
	UserLevel      uint32       `json:"userLevel"`
}

func (v *DokodemoConfig) Build() (proto.Message, error) {
	config := new(dokodemo.Config)
	if v.Host != nil {
		config.Address = v.Host.Build()
	}
	config.Port = uint32(v.PortValue)
	config.RelayPort = uint32(v.RelayPortValue)
	//newDebugMsg("Conf: Dokodemo: " + StructString(v))
	//newDebugMsg("Conf: Dokodemo: " + StructString(config))
	config.Networks = v.NetworkList.Build()
	config.Timeout = v.TimeoutValue
	config.FollowRedirect = v.Redirect
	config.UserLevel = v.UserLevel
	return config, nil
}
