package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
	"github.com/rubiojr/bonos/internal/config"
	"github.com/rubiojr/bonos/internal/constants"
	ijwt "github.com/rubiojr/bonos/internal/jwt"
	pb "github.com/rubiojr/bonos/proto/api/v1"
	"github.com/rubiojr/kv"
	kve "github.com/rubiojr/kv/errors"
	"golang.org/x/crypto/bcrypt"
)

type BonosServer struct {
	m  sync.Mutex
	db kv.Database
}

type User struct {
	Username string
	Password []byte
}

func (b *BonosServer) Remaining(ctx context.Context, req *pb.EmptyReq) (*pb.RemainingResp, error) {

	bono, err := b.LoadBono(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.RemainingResp{Remaining: bono.Remaining}, nil
}

func (b *BonosServer) Details(ctx context.Context, req *pb.EmptyReq) (*pb.DetailsResp, error) {

	bono, err := b.LoadBono(ctx)
	if err != nil {
		return nil, err
	}

	if bono.Remaining == 0 {
		return nil, errors.New("no current bono available, use /New to create one")
	}

	return &pb.DetailsResp{Remaining: bono.Remaining, CreatedAt: bono.CreatedAt}, nil
}

func (b *BonosServer) New(ctx context.Context, req *pb.NewReq) (*pb.NewResp, error) {
	b.m.Lock()
	defer b.m.Unlock()

	username := ctx.Value("username").(string)

	bono, err := b.LoadBono(ctx)
	if err == nil && bono.Remaining > 0 {
		return nil, fmt.Errorf("there's a bono already in use for %s", username)
	}

	if err != nil {
		if err == kve.ErrKeyNotFound {
			bono = &Bono{Username: username}
		} else {
			return nil, err
		}
	}

	amount := uint32(10)
	if req.Amount != 0 {
		amount = req.Amount
	}

	bono.Remaining = amount
	bono.CreatedAt = time.Now().Format(time.RFC3339)
	return &pb.NewResp{Remaining: bono.Remaining, CreatedAt: bono.CreatedAt}, b.serialize(bono)
}

func (b *BonosServer) Use(ctx context.Context, req *pb.EmptyReq) (*pb.UseResp, error) {
	b.m.Lock()
	defer b.m.Unlock()

	bono, err := b.LoadBono(ctx)
	if err != nil {
		return nil, err
	}

	if err := bono.Use(); err != nil {
		return nil, err
	}

	if err := b.serialize(bono); err != nil {
		return nil, fmt.Errorf("error saving bono to the database: %w")
	}

	return &pb.UseResp{Remaining: bono.Remaining}, nil
}

func (b *BonosServer) Stamps(ctx context.Context, req *pb.EmptyReq) (*pb.StampsResp, error) {
	bono, err := b.LoadBono(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.StampsResp{Stamps: bono.Stamps}, nil
}

func (b *BonosServer) serialize(bono *Bono) error {
	buf, err := json.Marshal(bono)
	if err != nil {
		return fmt.Errorf("error marshalling json: %w", err)
	}

	fmt.Printf("saving bono for /bono/%s\n", bono.Username)
	return b.db.Set(b.bonoKey(bono.Username), buf, nil)
}

func (b *BonosServer) loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid username or password",
		})
		return
	}

	user, err := b.getUser(username)
	var userCreated bool

	if err != nil {
		if err == kve.ErrKeyNotFound {
			user, err = b.createUser(username, password)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "user creation failed",
				})
				return
			}
			userCreated = true
		}
	}

	err = b.checkUserAuth(user, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "auth failed",
		})
		return
	}

	token, err := b.jwtToken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	status := gin.H{
		"message": "authenticated", "jwt": token,
	}

	if userCreated {
		status["message"] = fmt.Sprintf("user %s created", username)
	}

	c.JSON(http.StatusOK, status)
}

func (b *BonosServer) checkUserAuth(u *User, password string) error {
	hashFromDatabase := u.Password

	// Comparing the password with the hash
	if err := bcrypt.CompareHashAndPassword(hashFromDatabase, []byte(password)); err != nil {
		return err
	}

	return nil
}

func (b *BonosServer) createUser(user, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := User{Username: user, Password: hash}
	uj, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	return &u, b.db.Set(b.userKey(user), uj, nil)
}

func (b *BonosServer) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func (b *BonosServer) jwtToken(username string) (string, error) {
	claims := &ijwt.Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(constants.JWTExpiry * time.Minute).Unix(),
			Audience:  constants.Audience,
			Issuer:    constants.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(config.HMACSecret()))
}

func (b *BonosServer) LoadBono(ctx context.Context) (*Bono, error) {
	username := ctx.Value("username").(string)
	log.Info().Msgf("loading bono for %s", username)

	buf, err := b.db.Get(b.bonoKey(username))
	if err != nil {
		return nil, err
	}

	var bono Bono
	err = json.Unmarshal(buf, &bono)
	if err != nil {
		return nil, fmt.Errorf("error marshalling json: %w", err)
	}
	return &bono, nil
}

func (b *BonosServer) getUser(username string) (*User, error) {
	buf, err := b.db.Get(b.userKey(username))
	if err != nil {
		return nil, err
	}

	var u User
	err = json.Unmarshal(buf, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (b *BonosServer) bonoKey(username string) string {
	return fmt.Sprintf("/bono/%s", username)
}

func (b *BonosServer) userKey(username string) string {
	return fmt.Sprintf("/user/%s", username)
}
