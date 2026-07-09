package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tylerfiekens/local-sqs/internal/server"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	port := flag.Int("port", 9324, "port to listen on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("local-sqs listening on %s", addr)

	if err := http.ListenAndServe(addr, server.New()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
