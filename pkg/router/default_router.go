package router

import (
	"crypto/ed25519"
	"log"

	"github.com/matrix-org/pinecone/router"
)

type RouterConfig struct {
	HopLimiting bool
}

func CreateDefaultRouter(
	log *log.Logger,
	sk ed25519.PrivateKey,
	options RouterConfig,
	quit <-chan bool,
) SimRouter {
	rtr := &DefaultRouter{
		rtr: router.NewRouter(log, sk),
	}
	rtr.rtr.InjectPacketFilter(rtr.PingFilter)

	if options.HopLimiting {
		rtr.EnableHopLimiting()
	}
	rtr.EnableWakeupBroadcasts()
	go rtr.OverlayReadHandler(quit)

	return rtr
}
