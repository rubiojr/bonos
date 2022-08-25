package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"flag"

	pb "github.com/rubiojr/bonos/proto/api/v1"
	"github.com/rubiojr/kv"
	kve "github.com/rubiojr/kv/errors"
)

var dbPath string
var serverAddr string

type BonosServer struct {
	m    sync.Mutex
	bono *Bono
	db   kv.Database
}

type Bono struct {
	CreatedAt string   `json:"created_at"`
	Remaining uint32   `json:"remaining"`
	Stamps    []string `json:"stamps"`
}

func (b *Bono) Use() error {
	if b.Remaining <= 0 {
		return errors.New("no more passes left")
	}

	b.Remaining -= 1
	b.Stamps = append(b.Stamps, time.Now().Format(time.RFC3339))

	return nil
}

func (b *Bono) IsValid() bool {
	return b.Remaining > 0
}

func (s *BonosServer) Remaining(ctx context.Context, req *pb.EmptyReq) (*pb.RemainingResp, error) {
	return &pb.RemainingResp{Remaining: s.bono.Remaining}, nil
}

func (s *BonosServer) Details(ctx context.Context, req *pb.EmptyReq) (*pb.DetailsResp, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.bono.Remaining == 0 {
		return nil, errors.New("no current bono available")
	}

	return &pb.DetailsResp{Remaining: s.bono.Remaining, CreatedAt: s.bono.CreatedAt}, nil
}

func (s *BonosServer) New(ctx context.Context, req *pb.NewReq) (*pb.NewResp, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.bono.Remaining > 0 {
		return nil, errors.New("there's a bono already in use")
	}

	s.bono.Remaining = req.Amount
	s.bono.CreatedAt = time.Now().Format(time.RFC3339)
	return &pb.NewResp{Remaining: s.bono.Remaining, CreatedAt: s.bono.CreatedAt}, s.serialize()
}

func (s *BonosServer) Use(ctx context.Context, req *pb.EmptyReq) (*pb.UseResp, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if err := s.bono.Use(); err != nil {
		return nil, err
	}

	if err := s.serialize(); err != nil {
		return nil, fmt.Errorf("error saving bono to the database: %w")
	}

	return &pb.UseResp{Remaining: s.bono.Remaining}, nil
}

func (s *BonosServer) Stamps(ctx context.Context, req *pb.EmptyReq) (*pb.StampsResp, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return &pb.StampsResp{Stamps: s.bono.Stamps}, nil
}

func (b *BonosServer) serialize() error {
	buf, err := json.Marshal(b.bono)
	if err != nil {
		return fmt.Errorf("error marshalling json: %w", err)
	}

	return b.db.Set("bono", buf, nil)
}

func (b *BonosServer) load() error {
	var bono Bono

	buf, err := b.db.Get("bono")
	if err != nil && err == kve.ErrKeyNotFound {
		b.bono = &Bono{}
		return nil
	}

	if err != nil {
		return err
	}

	err = json.Unmarshal(buf, &bono)
	if err != nil {
		return fmt.Errorf("error marshalling json: %w", err)
	}
	b.bono = &bono

	return nil
}

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
