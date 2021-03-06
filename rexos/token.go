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
// Checks if the tokens custom clains contains a license item with the given composite name.
// If no composite name is attached, the license items are not verified.
// Only the first composite name of the array is checked.
func ValidateToken(c *gin.Context, signingKey string, compositeName ...string) {

	tokenString := c.GetHeader("authorization")
	if tokenString == "" {
		// If not access token is found in header, try to get the interceptor token which can be
		// merged in by the composite service itself.
		tokenString = c.GetString(AuthorizationKey)
		if tokenString == "" {
			log.Error("Missing authentication token in header")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}

	split := strings.Split(tokenString, " ")
	if strings.ToLower(split[0]) != "bearer" {
		log.Error("Missing bearer keyword in token")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	token, err := jwt.ParseWithClaims(split[1], &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
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

		// no composite name to check custom claims (license items)
		if len(compositeName) == 0 {
			c.Next()
			return
		}
		for i := range claims.ComplexAuthorities.LicenseItems {
			if claims.ComplexAuthorities.LicenseItems[i].Key == compositeName[0] {
				log.Debugf("License item for %s found.\n", compositeName[0])
				c.Next()
				return
			}
		}
		// no license item for composite found
		log.WithFields(event.Fields{
			"UserID": claims.UserID,
		}).Errorf("No license item for %s found.\n", compositeName[0])
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.AbortWithStatus(http.StatusForbidden)
	return
}
