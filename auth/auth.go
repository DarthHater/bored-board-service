package auth

import (
	"errors"
	"net/http"
	"fmt"
	"io/ioutil"
	"crypto/rsa"
	"time"
	"strings"

	"github.com/DarthHater/bored-board-service/database"
	"github.com/DarthHater/bored-board-service/model"
	"github.com/DarthHater/bored-board-service/constants"
	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// IAuth defines an interface for auth related functionality.
type IAuth interface {
	ReadAndSetKeys()
	UserIsLoggedIn() gin.HandlerFunc
	UserIsInRole(d database.IDatabase, roles []constants.Role) gin.HandlerFunc
	GetTokenKey(c *gin.Context, keyName string) (interface{}, error)
	CreateToken(user model.User) (string, error)
}

type Auth struct {
}

var (
	signKey   *rsa.PrivateKey
	verifyKey *rsa.PublicKey
)

// ReadAndSetKeys will read public and private RSA keys and create a key for signing JWTs.
func (a *Auth) ReadAndSetKeys() {
	a.setUpViper()

	privKeyPath := viper.GetString("PRIVATE_KEY_PATH")
	pubKeyPath := viper.GetString("PUBLIC_KEY_PATH")

	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.WithFields(log.Fields{
			"privateKeyPath": privKeyPath,
		}).Fatal(err)
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}

	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.WithFields(log.Fields{
			"publiceKeyPath": pubKeyPath,
		}).Fatal(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}
}

// UserIsLoggedIn will read the JWT in the request header and verify that it is legitimate.
func (a *Auth) UserIsLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {

		token, err := a.getToken(c)

		if err != nil {
			c.Abort()
			return
		}

		if token.Valid {
			// save token in context for use in other middleware
			c.Set("token", token)
		} else {
			origin := c.GetHeader("Origin")
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					c.JSON(http.StatusForbidden, gin.H{"err": "token is malformed"})
					log.Error(fmt.Sprintf("Attempt to access resources with malformed token from origin: %s", origin))
				} else if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
					c.JSON(http.StatusForbidden, gin.H{"err": "token is expired"})
				} else {
					c.JSON(http.StatusForbidden, gin.H{"err": "error reading token"})
					log.Error(fmt.Sprintf("Error accessing resources with token from origin %s", origin))
				}
			} else {
				c.JSON(http.StatusForbidden, gin.H{"err": "error reading token"})
				log.Error(fmt.Sprintf("Error accessing resources with token from origin %s", origin))
			}
			c.Abort()
		}
	}
}

// UserIsInRole accepts a list of roles and determines whether a user's role in a JWT is in that list.
func (a *Auth) UserIsInRole(d database.IDatabase, roles []constants.Role) gin.HandlerFunc {
	return func(c *gin.Context) {

		var userRole constants.Role
		value, err := a.GetTokenKey(c, constants.UserRole)

		if err != nil {
			panic(err)
		}

		userRole = constants.Role(value.(int))

		for _, role := range roles {
			if userRole == role {
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"err": "User doesn't have access"})
		c.Abort()
	}
}

func (a *Auth) GetTokenKey(c *gin.Context, keyName string) (interface{}, error) {
	var token interface{}
	var ok bool

	if token, ok = c.Get("token"); !ok {
		c.JSON(http.StatusForbidden, gin.H{"err": "Error accessing token"})
		c.Abort()
		return nil, nil
	}

	if claims, ok := token.(jwt.Token).Claims.(jwt.MapClaims); ok {
		return claims[keyName], nil
	} else {
		c.Abort()
		return nil, nil
	}
}

// CreateToken generates a signed JWT string based on user info.
func (a *Auth) CreateToken(user model.User) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := make(jwt.MapClaims)
	claims[constants.Expires] = time.Now().Add(time.Hour * 24 * 7).Unix() // expires in one week
	claims[constants.IssuedAt] = time.Now().Unix()
	claims[constants.UserName] = user.Username
	claims[constants.UserID] = user.ID
	claims[constants.UserRole] = user.UserRole
	token.Claims = claims

	return token.SignedString(signKey)
}

func (a *Auth) getToken(c *gin.Context) (*jwt.Token, error) {
	tokenString := c.GetHeader("Authorization")

	if tokenString == "" {
		origin := c.GetHeader("Origin")
		c.JSON(http.StatusForbidden, gin.H{"err": "token is required"})
		log.Error(fmt.Sprintf("Attempt to access resources without token from origin: %s", origin))
		return nil, errors.New("Error")
	}

	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	return token, err
}

func (a *Auth) setUpViper() {
	viper.SetDefault("PRIVATE_KEY_PATH", "/var/bored-board-service/.keys/app.rsa")
	viper.SetDefault("PUBLIC_KEY_PATH", "/var/bored-board-service/.keys/app.rsa.pub")
	viper.BindEnv("PRIVATE_KEY_PATH")
	viper.BindEnv("PUBLIC_KEY_PATH")
}
