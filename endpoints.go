package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
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

	// TODO - PAGINATION!!!!!!!!!!!!!!!

	var postsDB []struct {
		PostID     int64
		Posted     time.Time
		Username   string
		Title      string
		Text       string
		Likes      int64
		Dislikes   int64
		Reaction   string
		Categories string
	}
	cat := "%\"" + reqQuery("cat", r) + "\"%"
	userID := reqQuery("userID", r)
	search := "%" + reqQuery("search", r) + "%"
	status := reqQuery("status", r)

	query := `SELECT 
		p.postId,
		p.posted,
		u.username, 
		p.title, 
		p.text,
		COUNT(CASE WHEN r.reaction = 1 THEN 1 END),
		COUNT(CASE WHEN r.reaction = 0 THEN 1 END),
		(CASE WHEN p.userID = $1 THEN (CASE WHEN r.reaction = 1 THEN 'like' ELSE 'dislike' END) ELSE (CASE WHEN $1 = 0 THEN 'forbidden' ELSE 'idle' END) END),
		categories
	FROM posts AS p 
	LEFT JOIN reactions r ON r.postId = p.postId
	INNER JOIN users u ON u.userId = p.userId
	WHERE p.status > '0' 
	AND p.categories LIKE $2 
	AND p.userId LIKE $3 
	AND p.title LIKE $4 
	AND p.status LIKE $5 `

	sliceFromDB(&postsDB, query, fromCtx("userID", r), cat, userID, search, status)

	postsView := make([]struct {
		PostID     int64
		Posted     int64
		Username   string
		Title      string
		Text       string
		Likes      int64
		Dislikes   int64
		Reaction   string
		Categories interface{}
	}, len(postsDB))

	for i, p := range postsDB {
		postsView[i].PostID = p.PostID
		postsView[i].Posted = p.Posted.Unix()
		postsView[i].Username = p.Username
		postsView[i].Title = p.Title
		postsView[i].Text = p.Text
		postsView[i].Likes = p.Likes
		postsView[i].Dislikes = p.Dislikes
		postsView[i].Reaction = p.Reaction
		postsView[i].Categories = getCatNames(p.Categories)
	}
	returnJSON(postsView, w)
}

func readpost(w http.ResponseWriter, r *http.Request) {

	userID := fromCtx("userID", r)
	fmt.Println(userID)

	var postDB struct {
		PostID     int64
		Posted     time.Time
		Username   string
		Title      string
		Text       string
		Likes      int64
		Dislikes   int64
		Reaction   string
		Categories string
	}

	query := `SELECT 
		p.postId,
		p.posted,
		u.username, 
		p.title, 
		p.text,
		COUNT(CASE WHEN r.reaction = 1 THEN 1 END),
		COUNT(CASE WHEN r.reaction = 0 THEN 1 END),
		(CASE WHEN p.userID = $1 THEN (CASE WHEN r.reaction = 1 THEN 'like' ELSE 'dislike' END) ELSE (CASE WHEN $1 = 0 THEN 'forbidden' ELSE 'idle' END) END),
		categories
	FROM posts AS p
	LEFT JOIN reactions r ON r.postId = p.postId
	INNER JOIN users u ON u.userId = p.userId
	WHERE p.postId = $2`

	post := r.FormValue("postID")
	if post == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	structError := structFromDB(&postDB, query, userID, post)
	if structError != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	var postView struct {
		PostID     int64
		Posted     int64
		Username   string
		Title      string
		Text       string
		Likes      int64
		Dislikes   int64
		Reaction   string
		Categories interface{}
	}

	postView.PostID = postDB.PostID
	postView.Posted = postDB.Posted.Unix()
	postView.Username = postDB.Username
	postView.Title = postDB.Title
	postView.Likes = postDB.Likes
	postView.Dislikes = postDB.Dislikes
	postView.Reaction = postDB.Reaction
	postView.Categories = getCatNames(postDB.Categories)
	returnJSON(postView, w)
}
