package main

import (
	"context"
	"demo1/rpcclient"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	rpcURL := flag.String("rpc-url", "http://127.0.0.1:10099", "RPC server base URL")
	timeout := flag.Duration("timeout", 5*time.Second, "request timeout")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client, err := rpcclient.New(*rpcURL, rpcclient.WithHTTPClient(&http.Client{Timeout: *timeout}))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Health(ctx); err != nil {
		log.Fatalf("rpc health check failed: %v", err)
	}

	pong, err := client.Ping(ctx, "hi")
	if err != nil {
		log.Fatalf("system.ping failed: %v", err)
	}
	fmt.Printf("ping result: ok=%v echo=%q\n", pong.OK, pong.Echo)

	methods, err := client.Methods(ctx)
	if err != nil {
		log.Fatalf("system.methods failed: %v", err)
	}
	fmt.Printf("methods: %v\n", methods)
}
