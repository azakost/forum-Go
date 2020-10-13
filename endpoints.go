package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
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
	validity.regcheck("wrong username format", reg.Username, `^[a-zA-Z0-9_]{3,10}$`)
	validity.regcheck("wrong email format", reg.Email, `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	validity.regcheck("wrong fullname format", reg.Fullname, `^.{3,20}$`)
	validity.regcheck("short password", reg.Password, `^.{6,}$`)

	// Encrypt password for safe storage
	pass := encrypt(reg.Password)

	// Making conditional insert - if data is valid insert will be commited
	query := `INSERT INTO users(email, username, fullname, password) values( $1, $2, $3, $4)`
	insertError := conditionalInsert(len(validity) > 0, query, reg.Email, reg.Username, reg.Fullname, pass)

	// If conditional insert returned error - insert will rollback
	if insertError != nil {

		// Validate Email and Username for case of uniqness in DB
		validity.errcheck("email already exist", insertError, "users.email")
		validity.errcheck("user already exist", insertError, "users.user")

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

	// Get encrypted password from DB and user ID
	var creds []struct {
		Password string
		UserID   int64
	}

	query := `SELECT password, userId FROM users WHERE username = $1`
	structFromDB(&creds, query, login.Username)

	// If no such user in DB
	if len(creds) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	// If more than one username found - it would be a developer error
	if len(creds) > 1 {
		log.Println("Server Error: Check DB for username dublication!")
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// Check passwors
	if !cryptIsValid(creds[0].Password, login.Password) {
		http.Error(w, http.StatusText(403), 403)
		return
	}

	// Set new JWT if password correct
	setJWT(creds[0].UserID, w)

}

func logout(w http.ResponseWriter, r *http.Request) {
	userID := fromCtx("userID", r)
	delete(sessions, userID)
	addCookie(w, "jwt", "", time.Unix(0, 0))
}

func addpost(w http.ResponseWriter, r *http.Request) {

	// Struct request body
	var post struct {
		Title      string  `json:"title"`
		Text       string  `json:"text"`
		Categories []int64 `json:"categories"`
	}
	structBody(r, &post)

	// Create validation report
	var validity report
	validity.regcheck("wrong title", post.Title, `^.{3,140}$`)
	cats := processCategories(&validity, post.Categories)

	if len(validity) == 0 {
		userID := fromCtx("userID", r)
		query := `INSERT INTO posts(userId, title, text, categories) values($1, $2, $3, $4)`
		execQuery(query, userID, post.Title, post.Text, cats)
	} else {
		log.Println("User Error: Post content is not valid!")
		w.WriteHeader(400)
		returnJSON(validity, w)
	}
}

func updpost(w http.ResponseWriter, r *http.Request) {

	// Struct request body
	var post struct {
		PostID     int64   `json:"postID"`
		Title      string  `json:"title"`
		Text       string  `json:"text"`
		Categories []int64 `json:"categories"`
	}
	structBody(r, &post)

	// Create validation report
	var validity report
	validity.regcheck("wrong title", post.Title, `^.{3,140}$`)
	cats := processCategories(&validity, post.Categories)

	if len(validity) == 0 {
		userID := fromCtx("userID", r)
		query := `UPDATE posts SET 
			title = $1, 
			text = $2, 
			categories = $3 
		WHERE postId = $4 AND userId = $5`
		execQuery(query, post.Title, post.Text, cats, post.PostID, userID)
	} else {
		log.Println("User Error: Post content is not valid!")
		w.WriteHeader(400)
		returnJSON(validity, w)
	}
}

func viewposts(w http.ResponseWriter, r *http.Request) {
	var posts []struct {
		PostID     int64
		Title      string
		Text       string
		Categories string
	}
	query := `SELECT postId, title, text, categories FROM posts`
	structFromDB(&posts, query)

	new := make([]struct {
		PostID     int64
		Title      string
		Text       string
		Categories []int64
	}, len(posts))

	fn := func(c rune) bool {
		return !unicode.IsNumber(c)
	}

	for i, p := range posts {
		new[i].PostID = p.PostID
		new[i].Title = p.Title
		new[i].Text = p.Text
		cats := strings.FieldsFunc(p.Categories, fn)
		for _, c := range cats {
			id, _ := strconv.ParseInt(c, 10, 64)
			new[i].Categories = append(new[i].Categories, id)
		}
	}

	returnJSON(new, w)
}
