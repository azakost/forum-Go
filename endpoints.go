package main

import (
	"log"
	"net/http"
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
	var creds struct {
		Password string
		UserID   int64
	}

	query := `SELECT password, userId FROM users WHERE username = $1`
	structError := structFromDB(&creds, query, login.Username)

	// If no such user in DB
	if structError != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	// Check passwors
	if !cryptIsValid(creds.Password, login.Password) {
		http.Error(w, http.StatusText(403), 403)
		return
	}

	// Set new JWT if password correct
	setJWT(creds.UserID, w)

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
		Status     int64   `json:"status"`
		Categories []int64 `json:"categories"`
	}
	structBody(r, &post)

	// Create validation report
	var validity report
	validity.regcheck("wrong title", post.Title, `^.{3,140}$`)
	validity.logcheck("wrong status", post.Status > 2 || post.Status < 0)
	cats := processCategories(&validity, post.Categories)

	if len(validity) == 0 {
		userID := fromCtx("userID", r)
		query := `UPDATE posts SET 
			title = $1, 
			text = $2, 
			categories = $3,
			status = $4 
		WHERE postId = $5 AND userId = $6`
		execQuery(query, post.Title, post.Text, cats, post.Status, post.PostID, userID)
	} else {
		log.Println("User Error: Post content is not valid!")
		w.WriteHeader(400)
		returnJSON(validity, w)
	}
}

func viewposts(w http.ResponseWriter, r *http.Request) {
	var posts []struct {
		PostID     int64
		Username   string
		Title      string
		Text       string
		Categories string
	}
	cat := "%\"" + reqQuery("cat", r) + "\"%"
	userID := reqQuery("userID", r)
	search := "%" + reqQuery("search", r) + "%"
	status := reqQuery("status", r)

	query := `SELECT postId, u.username, title, text, categories 
	FROM posts AS p 
	INNER JOIN users AS u ON u.userId = p.userId
	WHERE p.status > '0' 
	AND p.categories LIKE $1 
	AND p.userId LIKE $2 
	AND p.title LIKE $3 
	AND p.status LIKE $4 `

	sliceFromDB(&posts, query, cat, userID, search, status)

	new := make([]struct {
		PostID     int64
		Username   string
		Title      string
		Text       string
		Categories []string
	}, len(posts))

	fn := func(c rune) bool {
		return !unicode.IsNumber(c)
	}
	var categ struct {
		Name string
	}
	catQuery := `SELECT name FROM categories WHERE categoryId = ?`

	for i, p := range posts {
		new[i].PostID = p.PostID
		new[i].Username = p.Username
		new[i].Title = p.Title
		new[i].Text = p.Text
		cats := strings.FieldsFunc(p.Categories, fn)
		for _, c := range cats {
			structError := structFromDB(&categ, catQuery, c)
			err(structError)
			new[i].Categories = append(new[i].Categories, categ.Name)
		}
	}

	returnJSON(new, w)
}

func test(w http.ResponseWriter, r *http.Request) {

	var user []struct {
		UserID   int64
		Username string
	}

	query := `SELECT userId, username FROM users WHERE userId = 5`
	sliceFromDB(&user, query)
	returnJSON(user, w)

}
