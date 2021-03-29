package rexos

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"

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

// type LicenseItemsValidator interface {
// 	LicenseItemsValid(*CustomClaims, interface{}) bool
// }

type LicenseItemsValidator func(*CustomClaims, interface{}) bool

// getKey verifies the given key. If the key is a simple signing key the key is returned
// as byte array ([]byte). If the key is a rsa/ dsa/ ecdsa key, the string is first pem decoded,
// then parsed and returned as public key in its particular type (e.g. *rsa.PublicKey)
func getKey(alg string, signingKey string, signingPublicKey []byte) interface{} {
	if alg == "HS256" {
		return []byte(signingKey)
	} else if alg == "RS256" {
		pub, err := x509.ParsePKIXPublicKey(signingPublicKey)

		if err != nil {
			log.Error("Get key for token validation. Cannot parse public key. " + err.Error())
			return nil
		}

		return pub.(*rsa.PublicKey)
	}
	log.Error("Get key for token validation. Not suppoorted token signature algorithm. " + alg)
	return nil
}

// ValidateToken checks the token of a given context
// Checks if the tokens custom clains contains a license item with the given composite name.
func ValidateToken(c *gin.Context, signingKey string, signingPublicKey []byte, licenseItemsValid LicenseItemsValidator, validationItems interface{}) {

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

	// parse Header
	parts := strings.Split(split[1], ".")
	token := &jwt.Token{Raw: split[1]}
	var headerBytes []byte
	var err error
	if headerBytes, err = jwt.DecodeSegment(parts[0]); err != nil {
		log.WithFields(event.Fields{
			"error": err.Error(),
		}).Error("Error decode header segment")
		return
	}
	if err = json.Unmarshal(headerBytes, &token.Header); err != nil {
		log.Error("Error unmarshal header bytes")
		return
	}
	alg := token.Header["alg"].(string)

	token, err = jwt.ParseWithClaims(split[1], &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getKey(alg, signingKey, signingPublicKey), nil
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

		if licenseItemsValid(claims, validationItems) {
			c.Next()
			return
		}

		log.WithFields(event.Fields{
			"UserID": claims.UserID,
		}).Error("No valid license items found.")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.AbortWithStatus(http.StatusForbidden)
	return
}

// ReadPEMFile reads a pem file and returns the public key
func ReadPEMFile(privateKeyReader io.Reader, size int64) ([]byte, error) {

	pembytes := make([]byte, size)
	buffer := bufio.NewReader(privateKeyReader)
	_, err := buffer.Read(pembytes)
	if err != nil {
		log.WithFields(event.Fields{
			"error": err.Error(),
		}).Error("Error reading PEM file.")
		return []byte{}, err
	}
	data, _ := pem.Decode([]byte(pembytes))

	return data.Bytes, nil
}

// ClaimsContainCompositeName checks if claims contains the given license item name
// If the given license items name is empty, the license items are not verified
func ClaimsContainCompositeName(claims *CustomClaims, item interface{}) bool {
	itemName := item.(string)
	if itemName == "" {
		return true
	}
	for i := range claims.ComplexAuthorities.LicenseItems {
		if claims.ComplexAuthorities.LicenseItems[i].Key == itemName {
			return true
		}
	}
	return false
}
