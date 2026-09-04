package main

import (
	"flag"
	"log"
	"net/http"

	"gramlane/internal/appenv"
	"gramlane/internal/server"
)

func main() {
	addr := flag.String("addr", appenv.Listen(), "listen address")
	flag.Parse()
	s, err := server.New(*addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Gramlane kasdomain %s data=%s public=%q", *addr, appenv.DataDir(), appenv.PublicBase())
	log.Fatal(http.ListenAndServe(*addr, s.Handler()))
}
