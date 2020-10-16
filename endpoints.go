package main

import (
	"net/http"
	"strconv"
	"strings"
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
	structError := structFromDB(&creds, query, nil, login.Username)

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
		w.WriteHeader(400)
		returnJSON(validity, w)
	}
}

func viewposts(w http.ResponseWriter, r *http.Request) {

	// Get request query params
	cat := "%\"" + reqQuery("cat", r) + "\"%"
	userID := reqQuery("userID", r)
	search := "%" + reqQuery("search", r) + "%"
	status := reqQuery("status", r)
	logged := fromCtx("userID", r)

	// Pagination params
	pageSize := 10
	page, atoiError := strconv.Atoi(r.FormValue("page"))
	if atoiError != nil || page <= 0 {
		page = 1
	}
	offset := page*pageSize - pageSize

	// Select RAW slice data from DB
	var postDB []struct {
		PostID     int64
		Posted     int64
		AuthorID   int64
		Username   string
		Title      string
		Text       string
		Likes      int64
		Dislikes   int64
		Reaction   string
		Categories []interface{}
	}
	query := `
	SELECT 
		p.postId,
		CAST(strftime('%s', p.posted) AS INT),
		p.userId,
		(SELECT username FROM users u WHERE u.userId = p.userId),
		p.title, 
		p.text,
		(SELECT COUNT(*) FROM reactions r WHERE r.postId = p.postId AND reaction = 'like'),
		(SELECT COUNT(*) FROM reactions r WHERE r.postId = p.postId AND reaction = 'dislike'),
		COALESCE((SELECT reaction FROM reactions r WHERE r.postId = p.postId AND r.userId = $1), "idle"),
		p.categories
	FROM posts p WHERE 
	p.status > '0' 
	AND p.categories LIKE $2 
	AND p.userId LIKE $3 
	AND p.title LIKE $4 
	AND p.status LIKE $5 LIMIT $6 OFFSET $7`

	sliceFromDB(&postDB, query, getCats, logged, cat, userID, search, status, pageSize, offset)

	if len(postDB) == 0 {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	returnJSON(postDB, w)

}

func readpost(w http.ResponseWriter, r *http.Request) {

	userID := fromCtx("userID", r)
	var postDB struct {
		PostID     int64
		Posted     int64
		Username   string
		Title      string
		Text       string
		Likes      int64
		Dislikes   int64
		Reaction   string
		Categories []interface{}
	}

	query := `
	SELECT 
		p.postId,
		CAST(strftime('%s', p.posted) AS INT),
		(SELECT username FROM users u WHERE u.userId = p.userId),
		p.title, 
		p.text,
		(SELECT COUNT(*) FROM reactions r WHERE r.postId = p.postId AND reaction = 'like'),
		(SELECT COUNT(*) FROM reactions r WHERE r.postId = p.postId AND reaction = 'dislike'),
		COALESCE((SELECT reaction FROM reactions r WHERE r.postId = p.postId AND r.userId = $1), "idle"),
		p.categories
	FROM posts p WHERE 
	p.postId = $2
	`

	post := r.FormValue("postID")
	if post == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	structError := structFromDB(&postDB, query, getCats, userID, post)
	if structError != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	returnJSON(postDB, w)

}

func readcomments(w http.ResponseWriter, r *http.Request) {
	userID := fromCtx("userID", r)
	var comments []struct {
		CommentID int64
		Commented int64
		AuthorID  int64
		Username  string
		Comment   string
		Like      int64
		Dislike   int64
		Reaction  string
	}
	query := `
	SELECT 
		commentId,
		CAST(strftime('%s', commented) AS INT),
		c.userId,
		(SELECT username FROM users u WHERE u.userId = c.userId),
		comment,
		(SELECT COUNT(*) FROM comreact r WHERE r.commentId = c.commentId AND reaction = 'like'),
		(SELECT COUNT(*) FROM comreact r WHERE r.commentId = c.commentId AND reaction = 'dislike'),
		COALESCE((SELECT reaction FROM comreact r WHERE r.commentId = c.commentId AND r.userId = $1), "idle")
	FROM comments c
	WHERE c.postId = $2
	`
	sliceFromDB(&comments, query, nil, userID, r.FormValue("postID"))
	returnJSON(comments, w)
}

func writecomment(w http.ResponseWriter, r *http.Request) {
	var comment struct {
		PostID  int64  `json:"postID"`
		Comment string `json:"comment"`
	}
	structBody(r, &comment)
	var validity report
	validity.regcheck("too short comment", strings.TrimSpace(comment.Comment), `^.{2,}$`)
	query := `INSERT INTO comments(postId, userId, comment) VALUES ((SELECT postId FROM posts WHERE postId = $1), $2, $3)`
	if len(validity) == 0 {
		execQuery(query, comment.PostID, fromCtx("userID", r), comment.Comment)
	} else {
		w.WriteHeader(400)
		returnJSON(validity, w)
	}
}

func reaction(w http.ResponseWriter, r *http.Request) {
	var reaction struct {
		PostID    int64  `json:"postID"`
		CommentID int64  `json:"commentID"`
		Reaction  string `json:"reaction"`
	}
	structBody(r, &reaction)

	var query string
	var id int64
	if reaction.PostID > 0 && reaction.CommentID == 0 && (reaction.Reaction == "like" || reaction.Reaction == "dislike") {
		id = reaction.PostID
		query = `INSERT INTO reactions(postId, userId, reaction) VALUES ((SELECT postId FROM posts WHERE postId = $1), $2, $3)`
	} else if reaction.PostID == 0 && reaction.CommentID > 0 && (reaction.Reaction == "like" || reaction.Reaction == "dislike") {
		id = reaction.CommentID
		query = `INSERT INTO comreact(commentId, userId, reaction) VALUES ((SELECT commentId FROM comments WHERE commentId = $1), $2, $3)`
	} else {
		w.WriteHeader(400)
		return
	}

	execQuery(query, id, fromCtx("userID", r), reaction.Reaction)
}
