package main

import (
	"log"
	"net/http"
)

func register(w http.ResponseWriter, r *http.Request) {

	// Struct request body
	var reg struct {
		Username string `json:"username"`
		Fullname string `json:"fullname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	structBody(r, &reg)

	// Create validation report
	var validity report

	// Regex validation of creds
	validity.regcheck("username", reg.Username, `^[a-zA-Z0-9_]{3,10}$`)
	validity.regcheck("email", reg.Email, `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	validity.regcheck("fullname", reg.Fullname, `^.{3,20}$`)
	validity.regcheck("password", reg.Password, `^.{6,}$`)

	// Encrypt password for safe storage
	pass := encrypt(reg.Password)

	// Making conditional insert - if data is valid insert will be commited
	query := `INSERT INTO users(email, username, fullname, password) values( $1, $2, $3, $4)`
	insertError := conditionalInsert(len(validity) > 0, query, reg.Email, reg.Username, reg.Fullname, pass)

	// If conditional insert returned error - insert will rollback
	if insertError != nil {

		// Validate Email and Username for case of uniqness in DB
		validity.errcheck("hasEmail", insertError, "users.email")
		validity.errcheck("hasUser", insertError, "users.user")

		// Return Error with validation report
		w.WriteHeader(400)
		returnJSON(validity, w)
		return
	}
}

func login(w http.ResponseWriter, r *http.Request) {

	// Struct request body
	var login struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	structBody(r, &login)

	// Get encrypted password from DB
	query := `SELECT password as "password" FROM users WHERE username = $1`
	result := JSONfromDB(query, login.Username)

	// If no such user in DB
	if len(result) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	// If more than one username found - it would be a developer error
	if len(result) > 1 {
		log.Println("Server Error: Check DB for username dublication!")
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// Check passwors
	pass := result[0].(map[string]interface{})["password"].(string)
	if !cryptIsValid(pass, login.Password) {
		http.Error(w, http.StatusText(403), 403)
		return
	}

	// Set new JWT if password correct
	setJWT(login.Username, w)

}

func logout(w http.ResponseWriter, r *http.Request) {

	query := `SELECT 
		title AS "title" 
	FROM posts`

	js := JSONfromDB(query)

	returnJSON(js, w)

}
