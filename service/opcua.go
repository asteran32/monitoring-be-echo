package service

import (
	"app/model"
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
)

var (
	configfile = "plcConfig.json"
)

type opcMessage struct {
	Event string  `json:"event"`
	Data  opcData `json:"data"`
}

type opcData struct {
	Time   string      `json:"time"`
	NodeID string      `json:"nodeid"`
	Value  interface{} `json:"value"`
}

// Set opc ua server options
func (t *ThreadSafeWriter) SetConfigurationAndRun(config model.OpcUAServer) {
	interval := opcua.DefaultSubscriptionInterval.String() //100ms

	subInterval, err := time.ParseDuration(interval)
	if err != nil {
		return
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-signalCh
		println()
		cancel()
	}()

	endpoints, err := opcua.GetEndpoints(config.Endpoint)
	if err != nil {
		return
	}

	ep := opcua.SelectEndpoint(endpoints, config.Policy, ua.MessageSecurityModeFromString(config.Mode))
	if ep == nil {
		return
	}

	opts := []opcua.Option{
		opcua.SecurityPolicy(config.Policy),
		opcua.SecurityModeString(config.Mode),
		opcua.CertificateFile(config.Cert),
		opcua.PrivateKeyFile(config.Key),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c := opcua.NewClient(config.Endpoint, opts...)
	if err := c.Connect(ctx); err != nil {
		return
	}

	defer c.Close()

	m, err := monitor.NewNodeMonitor(c)
	if err != nil {
		return
	}

	m.SetErrorHandler(func(_ *opcua.Client, sub *monitor.Subscription, err error) {
		log.Printf("error: sub=%d err=%s", sub.SubscriptionID(), err.Error())
	})

	// start channel-based subscription
	go t.RunChannelBasedSub(ctx, m, subInterval, 0, config.NodeID...)

	<-ctx.Done()
}

func (t *ThreadSafeWriter) RunChannelBasedSub(ctx context.Context, m *monitor.NodeMonitor, interval, lag time.Duration, nodes ...string) {
	ch := make(chan *monitor.DataChangeMessage, 16)
	sub, err := m.ChanSubscribe(ctx, &opcua.SubscriptionParameters{Interval: interval}, ch, nodes...)
	if err != nil {
		return
	}

	defer func() {
		Cleanup(sub)
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-ch:
			if msg.Error != nil {
				log.Printf("[channel ] sub=%d error=%s", sub.SubscriptionID(), msg.Error)
				return

			} else {
				opcData := opcData{
					Time:   msg.SourceTimestamp.UTC().Format(time.RFC3339),
					NodeID: msg.NodeID.String(),
					Value:  msg.Value.Value(),
				}

				// log.Printf("[channel ] sub=%d data=%v", sub.SubscriptionID(), opcData) // for check

				if err := t.Conn.WriteJSON(opcMessage{Event: "event", Data: opcData}); err != nil {
					return
				}
			}
			time.Sleep(lag)
		}
	}
}

func Cleanup(sub *monitor.Subscription) {
	log.Printf("stats: sub=%d delivered=%d dropped=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe()
}
