package main

import (
	"flag"
	"log"
	"net/http"

	pb "github.com/rubiojr/bonos/proto/api/v1"
	"github.com/rubiojr/kv"
)

func main() {
	flag.Parse()

	db, err := kv.New("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	bs := &BonosServer{db: db}
	if err := bs.load(); err != nil {
		log.Fatalf("error loading database: %s", err)
	}

	twirpHandler := pb.NewBonosServer(bs)
	mux := http.NewServeMux()
	mux.Handle(twirpHandler.PathPrefix(), twirpHandler)
	err = http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("error starting the service: %s", err)
	}
}

func init() {
	flag.StringVar(&dbPath, "db", "bonos.db", "database path")
	flag.StringVar(&serverAddr, "addr", ":65139", "server address")
}
