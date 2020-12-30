package util

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"log"
	"os"
	"time"
)

//secret key
var secretKey = []byte("abcd1234!@#$")

// ArithmeticCustomClaims custom declaration
type CustomClaims struct {
	Uid   string `json:"uid"`
	Email string `json:"Email"`
	jwt.StandardClaims
}

// jwtKeyFunc returns the key
func jwtKeyFunc(token *jwt.Token) (interface{}, error) {
	return secretKey, nil
}

// Sign generates a token
func ParseToken(authToken string) (map[string]interface{}, error) {
	// sample token string taken from the New example
	tokenString := authToken

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secretKey), nil
	})
	if err == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			log.Printf("claims : %v", claims)
			return claims, nil
		}
	} else {
		log.Printf("jwt error : %v", err)
		return nil, err
	}
	return nil, nil
}
func GenerateToken(uid, email string) (string, error) {

	// For the convenience of the demonstration, set the expiration after two minutes
	expAt := time.Now().Add(time.Minute * 10).Unix()

	// Create a statement
	claims := CustomClaims{
		Uid:   uid,
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expAt,
			Issuer:    "system",
		},
	}

	// Create a token, specify the encryption algorithm is HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate token
	return token.SignedString(secretKey)
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func CreateToken(userid string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 10).Unix()
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	var err error
	//Creating Access Token
	os.Setenv("ACCESS_SECRET", string(secretKey)) //this should be in an env file
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}
	//Creating Refresh Token
	os.Setenv("REFRESH_SECRET", string(secretKey)) //this should be in an env file
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["r_exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

//func CreateAuth(userid uint64, td *TokenDetails) error {
//	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
//	rt := time.Unix(td.RtExpires, 0)
//	now := time.Now()
//
//	errAccess := client.Set(td.AccessUuid, strconv.Itoa(int(userid)), at.Sub(now)).Err()
//	if errAccess != nil {
//		return errAccess
//	}
//	errRefresh := client.Set(td.RefreshUuid, strconv.Itoa(int(userid)), rt.Sub(now)).Err()
//	if errRefresh != nil {
//		return errRefresh
//	}
//	return nil
//}
