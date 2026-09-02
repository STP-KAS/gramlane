package main

import (
	"flag"
	"log"
	"net/http"

	"gramlane/internal/server"
)

func main() {
	addr := flag.String("addr", ":8081", "listen address")
	flag.Parse()
	s, err := server.New(*addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Gramlane dApp http://localhost%s — L1 work credits, no L2, not a dollar", *addr)
	log.Fatal(http.ListenAndServe(*addr, s.Handler()))
}
