package main

// https://github.com/auth0/go-jwt-middleware/blob/master/examples/http-example/main.go
// https://github.com/twitchtv/twirp/issues/91

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/rubiojr/bonos/internal/config"
	"github.com/rubiojr/bonos/internal/db"
	mw "github.com/rubiojr/bonos/middleware"
	pb "github.com/rubiojr/bonos/proto/api/v1"
)

var dbPath string
var serverAddr string

func main() {
	flag.Parse()

	err := config.Validate()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading configuration")
	}

	db, err := db.New(dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading database")
	}

	bs := &BonosServer{db: db.Handler}

	twirpHandler := pb.NewBonosServer(bs)
	mux := gin.New()
	mux.GET("/ping", bs.pingHandler)
	mux.POST("/login", bs.loginHandler)
	twirpGrp := mux.Group("/twirp", mw.JWTMiddleware)
	twirpGrp.POST("/*w", gin.WrapH(twirpHandler))

	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatal().Err(err).Msg("error starting the service")
	}
}

func init() {
	flag.StringVar(&dbPath, "db", "bonos.db", "database path")
	flag.StringVar(&serverAddr, "addr", ":65139", "server address")
}
