package auth

import (
    "time"
    "os"
    "net/http"

    "github.com/golang-jwt/jwt"
    "github.com/joho/godotenv"

    "github.com/enzdor/goblog/models"
)

func NewToken(length int) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
	ExpiresAt: time.Now().AddDate(0, 0, length).Unix(),
	Issuer: "https://goblog.com",
    })

    err := godotenv.Load()
    if err != nil {
	return "", err
    }
    key := os.Getenv("JWTKEY")

    tkn, err := token.SignedString([]byte(key))
    if err != nil {
	return "", err
    }

    return tkn, nil
}

func Authorizer(key string) func (w http.ResponseWriter, r *http.Request) error {
    return func (w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("auth")
	if err != nil {
	    return err
	}

	token, err := jwt.ParseWithClaims(cookie.Value, 
	    &jwt.StandardClaims{}, 
	    func(t *jwt.Token) (interface{}, error){
		return []byte(key), nil
	    })
	if err != nil {
	    return err
	}

	tkn, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
	    return &models.AuthError{Message: "Token claims do not match standard claims"} 
	}

	if err := tkn.Valid(); err != nil {
	    return err
	}

	return nil
    }
}









