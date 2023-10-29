package users

import (
	"database/sql"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const charintset = "0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// CreateUniqueUserSlug creates unique slug for user Eg:- axCbe690 (5 letter alphabets 3 letter numbers)
func CreateUniqueUserSlug(db *sql.DB) (string, error) {
	nonuniqueslug := true
	userslug := ""
	for nonuniqueslug {
		a := make([]byte, 5)
		for i := range a {
			a[i] = charset[seededRand.Intn(len(charset))]
		}
		b := make([]byte, 3)
		for i := range b {
			b[i] = charintset[seededRand.Intn(len(charintset))]
		}
		userslug = string(a) + string(b)
		var count sql.NullInt64
		err := db.QueryRow(IsUniqueSlug, userslug).Scan(&count)
		if err != nil {
			return "", err
		}
		if count.Int64 == 0 {
			nonuniqueslug = false
		}
	}

	return userslug, nil

}

// GenerateJWTAccessToken creates access token
func GenerateJWTAccessToken(userId int64, userslug, secretKey string, tokenLifeTime time.Duration) (string, error) {
	var mySigningKey = []byte(secretKey)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["userslug"] = userslug
	claims["userid"] = userId
	claims["exp"] = time.Now().Add(tokenLifeTime).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		return "", err
	}
	return tokenString, nil
}
