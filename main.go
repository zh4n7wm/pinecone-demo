package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/matrix-org/pinecone/connections"
	"github.com/matrix-org/pinecone/multicast"
	"github.com/matrix-org/pinecone/router"
	"github.com/matrix-org/pinecone/router/events"
	"github.com/matrix-org/pinecone/types"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	listentcp := flag.String("listen", ":0", "address to listen for TCP connections")
	connect := flag.String("connect", "", "peers to connect to")
	flag.Parse()

	_, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", 0)
	// pineconeRouter := router.NewRouter(logger, sk, router.RouterOptionBlackhole(true))
	pineconeRouter := router.NewRouter(logger, sk)
	pineconeRouter.EnableHopLimiting()
	pineconeRouter.EnableWakeupBroadcasts()
	pineconeEvents := make(chan events.Event)
	pineconeRouter.Subscribe(pineconeEvents)

	pineconeMulticast := multicast.NewMulticast(logger, pineconeRouter)
	pineconeMulticast.Start()

	connMgr := connections.NewConnectionManager(pineconeRouter, nil)

	listener := net.ListenConfig{}
	if listentcp != nil && *listentcp != "" {
		go func() {
			listener, err := listener.Listen(context.Background(), "tcp", *listentcp)
			if err != nil {
				panic(err)
			}
			fmt.Printf("listening on: %s", listener.Addr())

			for {
				conn, err := listener.Accept()
				if err != nil {
					panic(err)
				}

				if _, err := pineconeRouter.Connect(
					conn,
					router.ConnectionURI(conn.RemoteAddr().String()),
					router.ConnectionPeerType(router.PeerTypeRemote),
				); err != nil {
					fmt.Println("Inbound TCP connection", conn.RemoteAddr(), "error:", err)
					conn.Close()
				} else {
					fmt.Println("Inbound TCP connection", conn.RemoteAddr(), "is connected")
				}
			}
		}()
	}

	if connect != nil && *connect != "" {
		for _, uri := range strings.Split(*connect, ",") {
			// NOTE:
			// - 一个网络中可以通过 `Router identity` 连接
			// - 不再一个网络中可以通过 websocket、tcp address 连接，比如：`-connect host:port`
			connMgr.AddPeer(strings.TrimSpace(uri))
		}
	}

	/*
		go func(ch <-chan events.Event) { // events
			for {
				select {
				case event := <-ch:
					switch e := event.(type) {
					case events.PeerAdded:
						fmt.Printf("add peer event: %+v\n", e)
					case events.PeerRemoved:
						fmt.Printf("remove peer event: %+v\n", e)
					case events.BroadcastReceived:
						fmt.Printf("broadcast event: %+v\n", e)
					default:
						fmt.Printf("unknown event: %+v\n", e)
					}
				}
			}
		}(pineconeEvents)
	*/

	/*
		go func() { // send
			for {
				time.Sleep(2 * time.Second)
				peers := pineconeRouter.Peers()
				fmt.Printf("peer: %+v\n", peers)

				for _, p := range peers {
					if pineconeRouter.PublicKey().String() == p.PublicKey {
						continue
					}

					msg := []byte(fmt.Sprintf("ping from %s", pineconeRouter.PublicKey()))

					pkBytes, err := hex.DecodeString(p.PublicKey)
					if err != nil {
						fmt.Printf("decode hex public key failed: %s\n", err)
						continue
					}
					pk := types.PublicKey{}
					copy(pk[:], pkBytes)
					fmt.Printf("send %q to %s\n", msg, pk)

					// send msg
					if _, err := pineconeRouter.WriteTo(
						[]byte(msg), pk,
					); err != nil {
						fmt.Printf("write to %s failed: %s\n", pk, err)
						continue
					}
					time.Sleep(5 * time.Second)
				}
			}
		}()
	*/

	go func() { // read msg
		for {
			var msg [types.MaxPayloadSize]byte
			_, remoteAddr, err := pineconeRouter.ReadFrom(msg[:])
			if err != nil {
				fmt.Printf("got msg failed: %s\n", err)
				continue
			}
			content := bytes.TrimRight(msg[:], "\x00")
			fmt.Printf("got msg <%s> from <%s>\n", content, remoteAddr)
		}
	}()

	<-sigs
}
