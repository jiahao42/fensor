// +build !confonly

package dokodemo

//go:generate errorgen

import (
	"context"
	"sync/atomic"
	"time"

	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/db"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/session"
	"v2ray.com/core/common/signal"
	"v2ray.com/core/common/task"
	"v2ray.com/core/features/policy"
	"v2ray.com/core/features/routing"
	"v2ray.com/core/transport/internet"
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		d := new(DokodemoDoor)
		err := core.RequireFeatures(ctx, func(pm policy.Manager) error {
			return d.Init(config.(*Config), pm)
		})
		return d, err
	}))
}

type DokodemoDoor struct {
	policyManager policy.Manager
	config        *Config
	address       net.Address
	port          net.Port
	relayport     net.Port
	pool          *db.Pool
	useRelay      bool
	relayInitStep int
	targetAddr    string
}

// Init initializes the DokodemoDoor instance with necessary parameters.
func (d *DokodemoDoor) Init(config *Config, pm policy.Manager) error {
	if (config.NetworkList == nil || len(config.NetworkList.Network) == 0) && len(config.Networks) == 0 {
		return newError("no network specified")
	}
	d.config = config
	d.address = config.GetPredefinedAddress()
	//newDebugMsg("Predefined address " + d.address.String())
	d.port = net.Port(config.Port)
	d.relayport = net.Port(config.RelayPort)
	d.policyManager = pm
	d.pool = db.New()
	d.pool.Start("tcp", "localhost", "6379")
	d.useRelay = false
	d.relayInitStep = 1

	//newDebugMsg("DokodemoDoor: " + StructString(d.port))
	newDebugMsg("DokodemoDoor: Port " + StructString(config.Port) + ", " + StructString(config.RelayPort))

	return nil
}

// Network implements proxy.Inbound.
func (d *DokodemoDoor) Network() []net.Network {
	if len(d.config.Networks) > 0 {
		return d.config.Networks
	}

	return d.config.NetworkList.Network
}

func (d *DokodemoDoor) policy() policy.Session {
	config := d.config
	p := d.policyManager.ForLevel(config.UserLevel)
	if config.Timeout > 0 && config.UserLevel == 0 {
		p.Timeouts.ConnectionIdle = time.Duration(config.Timeout) * time.Second
	}
	return p
}

type hasHandshakeAddress interface {
	HandshakeAddress() net.Address
}

// Process implements proxy.Inbound.
func (d *DokodemoDoor) Process(ctx context.Context, network net.Network, conn internet.Connection, dispatcher routing.Dispatcher) error {
	newError("processing connection from: ", conn.RemoteAddr()).AtDebug().WriteToLog(session.ExportIDToError(ctx))
	dest := net.Destination{
		Network: network,
		Address: d.address,
		Port:    d.port,
	}
	relayDest := net.Destination{
		Network: network,
		Address: d.address,
		Port:    d.relayport,
	}

	destinationOverridden := false
	if d.config.FollowRedirect {
		if outbound := session.OutboundFromContext(ctx); outbound != nil && outbound.Target.IsValid() {
			dest = outbound.Target
			destinationOverridden = true
		} else if handshake, ok := conn.(hasHandshakeAddress); ok {
			addr := handshake.HandshakeAddress()
			if addr != nil {
				dest.Address = addr
				destinationOverridden = true
			}
		}
	}
	if !dest.IsValid() || dest.Address == nil {
		return newError("unable to get destination")
	}

	if inbound := session.InboundFromContext(ctx); inbound != nil {
		inbound.User = &protocol.MemoryUser{
			Level: d.config.UserLevel,
		}
	}

	plcy := d.policy()
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, plcy.Timeouts.ConnectionIdle)

	ctx = policy.ContextWithBufferPolicy(ctx, plcy.Buffer)
	link, err := dispatcher.Dispatch(ctx, dest)
	relayLink, err := dispatcher.Dispatch(ctx, relayDest)
	if relayLink != nil {
	}

	if err != nil {
		return newError("failed to dispatch request").Base(err)
	}

	requestCount := int32(1)
	requestDone := func() error {
		defer func() {
			if atomic.AddInt32(&requestCount, -1) == 0 {
				timer.SetTimeout(plcy.Timeouts.DownlinkOnly)
			}
		}()

		var reader buf.Reader
		if dest.Network == net.Network_UDP {
			reader = buf.NewPacketReader(conn)
		} else {
			reader = buf.NewReader(conn)
		}
		if !d.useRelay {
			d.targetAddr, err = buf.SmartCopy(reader, link.Writer, d.pool, buf.UpdateActivity(timer))
			//newDebugMsg("Dokodemo: SmartCopy return buffer " + d.targetAddr)
			if err != nil && err.Error() == "USE_RELAY" {
				d.useRelay = true
			}
		}
		if d.useRelay {
			if d.relayInitStep == 1 {
				// Should have a func to create SOCKS conn, and send the unfinished request
				_, err = buf.RelayCopy(reader, relayLink.Writer, d.relayInitStep, d.targetAddr, buf.UpdateActivity(timer))
				if err != nil {
					return newError("failed to transport request").Base(err)
				}
				d.relayInitStep++
			} else if d.relayInitStep == 3 {
				buf.RelayCopy(reader, relayLink.Writer, d.relayInitStep, d.targetAddr, buf.UpdateActivity(timer))
				d.relayInitStep++
			} else {
				// operate as normal
				err = buf.Copy(reader, link.Writer, buf.UpdateActivity(timer))
			}

		}
		return nil
	}

	tproxyRequest := func() error {
		return nil
	}

	var writer buf.Writer
	if network == net.Network_TCP {
		writer = buf.NewWriter(conn)
	} else {
		//if we are in TPROXY mode, use linux's udp forging functionality
		if !destinationOverridden {
			writer = &buf.SequentialWriter{Writer: conn}
		} else {
			sockopt := &internet.SocketConfig{
				Tproxy: internet.SocketConfig_TProxy,
			}
			if dest.Address.Family().IsIP() {
				sockopt.BindAddress = dest.Address.IP()
				sockopt.BindPort = uint32(dest.Port)
			}
			tConn, err := internet.DialSystem(ctx, net.DestinationFromAddr(conn.RemoteAddr()), sockopt)
			if err != nil {
				return err
			}
			defer tConn.Close()

			writer = &buf.SequentialWriter{Writer: tConn}
			tReader := buf.NewPacketReader(tConn)
			requestCount++
			tproxyRequest = func() error {
				defer func() {
					if atomic.AddInt32(&requestCount, -1) == 0 {
						timer.SetTimeout(plcy.Timeouts.DownlinkOnly)
					}
				}()
				newDebugMsg("Dokodemo: TPROXY mode")
				if err := buf.Copy(tReader, link.Writer, buf.UpdateActivity(timer)); err != nil {
					return newError("failed to transport request (TPROXY conn)").Base(err)
				}
				return nil
			}
		}
	}

	responseDone := func() error {
		defer timer.SetTimeout(plcy.Timeouts.UplinkOnly)
		//newDebugMsg("Dokodemo: responseDone started")

		// Write to the forwarded address
		//buffer, err := buf.SmartCopy(link.Reader, writer, d.pool, buf.UpdateActivity(timer))
		// TODO: through the response, we may able to distinguish the blank pages
		if !d.useRelay {
			// commuicate as normal
			err = buf.Copy(link.Reader, writer, buf.UpdateActivity(timer))
		}
		if d.useRelay {
			if d.relayInitStep == 2 {
				buf, err := buf.RelayCopy(link.Reader, writer, d.relayInitStep, d.targetAddr, buf.UpdateActivity(timer))
				if err != nil {
					return newError("failed to transport request").Base(err)
				}
				//newDebugMsg("Dokodemo: responseDone step 2 buf " + buf)
				if buf == "\x05\x00" {
					//newDebugMsg("Dokodemo: responseDone goto step 3")
					d.relayInitStep++
				}
			} else if d.relayInitStep == 4 {
				// operate as normal
				err = buf.Copy(link.Reader, writer, buf.UpdateActivity(timer))
			}
		}
		return nil
	}

	if err := task.Run(ctx, task.OnSuccess(requestDone, task.Close(link.Writer)), responseDone, tproxyRequest); err != nil {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		common.Interrupt(relayLink.Reader)
		common.Interrupt(relayLink.Writer)
		return newError("connection ends").Base(err)
	}

	return nil
}
