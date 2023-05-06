package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DhruvikDonga/wordsbattle/cmd/server/modules/users"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	rs "github.com/DhruvikDonga/wordsbattle/pkg"
)

// RegisterHandler registers a new user
func (app *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {

	var in *users.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&in)

	if err != nil {
		res := fmt.Sprintf("failed to parse request string Err: %v" + err.Error())
		rs.RespondWithError(w, http.StatusInternalServerError, res)
		return
	}

	// check email is unique
	var count sql.NullInt64
	err = app.db.QueryRowContext(r.Context(), users.IsUniqueEmail, in.Email).Scan(&count)
	if err != nil {
		res := fmt.Sprintf("failed to run query string Err: %v" + err.Error())
		rs.RespondWithError(w, http.StatusInternalServerError, res)
		return
	}
	if count.Int64 > 0 {
		res := fmt.Sprintf("Email address already exisist")
		rs.RespondWithError(w, http.StatusPreconditionFailed, res)
		return
	}
	// create userslug
	userslug, err := users.CreateUniqueUserSlug(app.db)
	if err != nil {
		res := fmt.Sprintf("failed operation creating slug %v", err)
		rs.RespondWithError(w, http.StatusInternalServerError, res)
		return
	}
	log.Println(userslug)

	// create hash password
	bytes, err := bcrypt.GenerateFromPassword([]byte(in.Password), 14)
	in.Password = string(bytes)

	_, err = app.db.ExecContext(r.Context(), users.CreateUserQuery, in.UserName, userslug, in.Email, in.Password, time.Now())
	if err != nil {
		res := fmt.Sprintf("failed to enter new user error in processing query Error{errorcode :- %v , errormessage :- %v}", err.(*pq.Error).Code, err)
		rs.RespondWithError(w, http.StatusInternalServerError, res)
		return
	}
	res := &users.RegisterResponse{
		UserSlug: userslug,
		UserName: in.UserName,
	}
	rs.RespondwithJSON(w, http.StatusCreated, res)
}

// LoginHandler login a user
func (app *App) LoginHandler(w http.ResponseWriter, r *http.Request) {

	var in *users.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&in)

	if err != nil {
		res := fmt.Sprintf("failed to parse request string Err: %v" + err.Error())
		rs.RespondWithError(w, http.StatusInternalServerError, res)
		return
	}

	// check email is present
	var user_slug, user_password sql.NullString
	var user_id sql.NullInt64
	err = app.db.QueryRowContext(r.Context(), users.GetUserQuery, in.Email).Scan(&user_id, &user_slug, &user_password)
	if err != nil {
		res := fmt.Sprintf("failed to run query string Err: %v" + err.Error())
		rs.RespondWithError(w, http.StatusInternalServerError, res)
		return
	}
	if !user_slug.Valid {
		res := fmt.Sprintf("Email address or password not correct")
		rs.RespondWithError(w, http.StatusNotFound, res)
		return
	}

	// compare hash password
	err = bcrypt.CompareHashAndPassword([]byte(user_password.String), []byte(in.Password))
	if err != nil {
		res := fmt.Sprintf("Email address or password not correct")
		rs.RespondWithError(w, http.StatusNotFound, res)
		return
	}

	//generate token
	validToken, err := users.GenerateJWTAccessToken(user_id.Int64, user_slug.String, app.config.JwtSecretKey, app.config.AccessTokenDuration)
	if err != nil {
		res := fmt.Sprintf("failed to generate token string Err: %v" + err.Error())
		rs.RespondWithError(w, http.StatusInternalServerError, res)
		return
	}
	res := &users.LoginRespone{
		UserSlug:    user_slug.String,
		AccessToken: validToken,
	}
	rs.RespondwithJSON(w, http.StatusCreated, res)
}
