package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
	"github.com/rubiojr/bonos/internal/config"
	myjwt "github.com/rubiojr/bonos/internal/jwt"
)

var hmacSecret = []byte(config.HMACSecret())

func JWTMiddleware(c *gin.Context) {
	authHead := c.Request.Header["Authorization"]
	if len(authHead) == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
		return
	}

	var tknStr string
	if t := strings.Split(authHead[0], "Bearer "); len(t) > 1 {
		tknStr = t[1]
	}

	claims := &myjwt.Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})

	if err != nil {
		log.Error().Err(err).Msg("error validating JWT")
		c.AbortWithStatusJSON(
			http.StatusUnauthorized, gin.H{"error": "invalid JWT token"},
		)
		return
	}

	if !tkn.Valid {
		log.Error().Err(err).Msg("JWT token not valid")
		c.AbortWithStatusJSON(
			http.StatusUnauthorized, gin.H{"error": "invalid JWT token"},
		)
		return
	}

	newc := context.WithValue(c.Request.Context(), "username", claims.Username)
	c.Request = c.Request.WithContext(newc)

	log.Info().Msgf("JWT token valid for user %s", claims.Username)
	c.Next()
}
