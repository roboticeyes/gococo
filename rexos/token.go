package rexos

import (
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/roboticeyes/gococo/event"
)

const (
	// AuthorizationKey is the key for getting the HTTP header authorization
	AuthorizationKey = "authorization"

	// KeyUserID is used as an identifier
	KeyUserID = "UserID"
)

// CustomClaims is our custom metadata of the JWT
type CustomClaims struct {
	ComplexAuthorities struct {
		MaxStorage struct {
			Value int `json:"value"`
		} `json:"max_storage"`
		LicenseItems []struct {
			Key          string `json:"key"`
			ValueBoolean bool   `json:"valueBoolean,omitempty"`
			ValueLong    int    `json:"valueLong,omitempty"`
		} `json:"license_items"`
	} `json:"complex_authorities"`
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// ValidateToken checks the token of a given context
func ValidateToken(c *gin.Context) {

	tokenString := c.GetHeader("authorization")
	if tokenString == "" {
		log.Error("Missing authentication token in header")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	split := strings.Split(tokenString, " ")
	if strings.ToLower(split[0]) != "bearer" {
		log.Error("Missing bearer keyword in token")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	token, err := jwt.ParseWithClaims(split[1], &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(GlobalVault.JwtSigningKey), nil
	})
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Validate the token and return the custom claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		c.Set(KeyUserID, claims.UserID)
		t := time.Unix(claims.StandardClaims.ExpiresAt, 0)
		log.WithFields(event.Fields{
			"UserID": claims.UserID,
		}).Debugf("Token is valid. Expires in %v\n", t.Sub(time.Now()))
		// for _, licenseItem := range claims.ComplexAuthorities.LicenseItems {
		// 	log.Printf("License item: %s\n", licenseItem.Key)
		// }
	} else {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.Next()
}
