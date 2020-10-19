package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {

	// Create Database if not exist and execute pre-written initial query from "config.go"
	createDB(initialQuery)

	// Get port from dedicated server environment, if running locally assign port to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Starting handler to run our front-end application
	http.Handle("/", http.FileServer(http.Dir("./front")))

	// Our API endpoints
	endpoint("/api/register", register)
	endpoint("/api/login", login)
	endpoint("/api/logout", logout, "check JWT")

	// Get all posts or sertain post by ID (also include all post filtering)
	endpoint("/api/posts", posts)

	// Write or update post
	endpoint("/api/writepost", writepost, "check JWT")

	// Get all comments by post ID
	endpoint("/api/comments", comments)

	// Wrie comment or update comment
	endpoint("/api/writecomment", writecomment, "check JWT")

	// Like-Dislike on post or comment
	endpoint("/api/reaction", reaction, "check JWT")

	// ADMIN FEATURES
	endpoint("/api/categories", categories)
	endpoint("/api/updcategory", updcategory, "check JWT")
	endpoint("/api/deletecategory", deletecategory, "check JWT")
	endpoint("/api/users", users, "check JWT")

	// Listen server
	log.Println("Running http://localhost:" + port)
	e := http.ListenAndServe(":"+port, nil)
	log.Println(e)

}

type ctxKey string
type ctxData struct {
	ID   int64
	Role string
}

// Middleware handler
func endpoint(path string, page func(w http.ResponseWriter, r *http.Request), secure ...interface{}) {

	fn := func(w http.ResponseWriter, r *http.Request) {

		// Set headers for CORS enabling
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		isValid, id, role := validateJWT(w, r)

		// JWT validation handler
		if len(secure) > 0 && !isValid {
			http.Error(w, http.StatusText(403), 403)
			return
		}

		// Save userID and User Role to context
		var data ctxData
		data.ID = id
		data.Role = role

		var key ctxKey = "user"
		ctx := context.WithValue(r.Context(), key, data)
		req := r.WithContext(ctx)

		// Error handler
		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case *json.SyntaxError, *json.UnmarshalTypeError:
					http.Error(w, http.StatusText(400), 400)
				default:
					log.Printf("Server Error: %+v", err)
					http.Error(w, http.StatusText(500), 500)
				}
			}
		}()

		http.HandlerFunc(page).ServeHTTP(w, req)
	}
	http.Handle(path, http.HandlerFunc(fn))
}
