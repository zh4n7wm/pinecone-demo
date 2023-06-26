package main

import (
	"context"
	"crypto/ed25519"
	"demo/pkg/router"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	prouter "github.com/matrix-org/pinecone/router"
	"github.com/matrix-org/pinecone/types"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	listentcp := flag.String("listen", ":0", "address to listen for TCP connections")
	connect := flag.String("connect", "", "peers to connect to")
	pubKeyStr := flag.String("pubkey", "", "peer public key")
	flag.Parse()

	_, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", 0)
	quit := make(chan bool)
	r := router.CreateDefaultRouter(logger, sk, router.RouterConfig{HopLimiting: true}, quit)
	connMgr := r.NewConnectionManager(nil)

	if listentcp != nil && len(*listentcp) > 0 {
		listener := net.ListenConfig{}
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

				if _, err := r.Connect(
					conn,
					prouter.ConnectionURI(conn.RemoteAddr().String()),
					prouter.ConnectionPeerType(prouter.PeerTypeRemote),
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
		for _, item := range strings.Split(*connect, ",") {
			connMgr.AddPeer(strings.TrimSpace(item))
		}
	}

	if pubKeyStr != nil && len(*pubKeyStr) > 0 {
		pkBytes, err := hex.DecodeString(*pubKeyStr)
		if err != nil {
			log.Fatalf("decode hex public key failed: %s\n", err)
		}
		pk := types.PublicKey{}
		copy(pk[:], pkBytes)

		ctx := context.Background()
		hops, duration, err := r.Ping(ctx, pk)
		if err != nil {
			panic(err)
		}
		fmt.Printf("hops: %s duration: %s\n", hops, duration)
	}

	<-sigs
}
