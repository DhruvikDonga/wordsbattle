package handler

import (
	"fmt"
	"net/http"

	rs "github.com/DhruvikDonga/wordsbattle/pkg"
	"github.com/golang-jwt/jwt"
)

func (app *App) IsAuthorized(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] == nil {
			res := fmt.Sprintf("No Token Found")
			rs.RespondWithError(w, http.StatusBadRequest, res)
			return
		}

		var mySigningKey = []byte(app.config.JwtSecretKey)

		token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error in parsing")
			}
			return mySigningKey, nil
		})

		if err != nil {
			res := fmt.Sprintf("Your Token has been expired")
			rs.RespondWithError(w, http.StatusBadRequest, res)
			return
		}

		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// if claims["role"] == "admin" {

			// 	r.Header.Set("Role", "admin")
			// 	handler.ServeHTTP(w, r)
			// 	return

			// } else if claims["role"] == "user" {

			// 	r.Header.Set("Role", "user")
			// 	handler.ServeHTTP(w, r)
			// 	return
			// }
			handler.ServeHTTP(w, r)
			return
		}

		res := fmt.Sprintf("Not Authorized")
		rs.RespondWithError(w, http.StatusBadRequest, res)

	}
}
