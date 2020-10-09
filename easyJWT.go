package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func encrypt(data string) string {
	crypt, bycryptError := bcrypt.GenerateFromPassword([]byte(data+salt), bcrypt.MinCost)
	err(bycryptError)
	return string(crypt)
}

func cryptIsValid(crypt, data string) bool {
	e := bcrypt.CompareHashAndPassword([]byte(crypt), []byte(data+salt))
	return e == nil
}

var sessions = make(map[string]string)

type jwt struct {
	Username string    `json:"username"`
	Expire   time.Time `json:"expire"`
	Token    string    `json:"token"`
}

func setJWT(username string, w http.ResponseWriter) {
	exp := time.Now().Add(tokenLife)
	token := encrypt(username + exp.String())

	var jTok jwt

	jTok.Username = username
	jTok.Expire = exp
	jTok.Token = token

	data, marshalError := json.Marshal(jTok)
	err(marshalError)

	// Encrypt JWT
	block, cipherError := aes.NewCipher([]byte(secret))
	err(cipherError)
	gcm, gcmError := cipher.NewGCM(block)
	err(gcmError)
	nonce := make([]byte, gcm.NonceSize())
	_, readError := io.ReadFull(rand.Reader, nonce)
	err(readError)
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	val := base64.StdEncoding.EncodeToString([]byte(ciphertext))

	// Write JWT to map
	sessions[username] = token

	// Add cookie
	addCookie(w, "jwt", val, exp)
}

func validateJWT(w http.ResponseWriter, r *http.Request) bool {

	// Read Cookie
	var jTok jwt
	credsError := getCreds(&jTok, r)
	if credsError != nil {
		// addCookie(w, "jwt", "", time.Unix(0, 0))
		return false
	}

	// Check if email exists in sessions map
	val, ok := sessions[jTok.Username]

	if !ok {
		// addCookie(w, "jwt", "", time.Unix(0, 0))
		return false
	}

	// Filter 1 - Stright token compare and fail if not equal

	if val != jTok.Token {
		// delete(sessions, jTok.Username)
		// addCookie(w, "jwt", "", time.Unix(0, 0))
		return false
	}

	// Filter 2 - If token is expired
	if time.Now().After(jTok.Expire) {
		delete(sessions, jTok.Username)
		// addCookie(w, "jwt", "", time.Unix(0, 0))
		return false
	}

	// Filter 3 - if token is more than 1 hour old - refresh it
	if jTok.Expire.Before(time.Now().Add(tokenRefresh)) {
		setJWT(jTok.Username, w)
	}
	return true
}

func getCreds(model *jwt, r *http.Request) error {
	// Read Cookie
	cookie, cookieError := r.Cookie("jwt")
	if cookieError != nil {
		return cookieError
	}

	// Decode Cookie value from base64
	data, base64Error := base64.StdEncoding.DecodeString(cookie.Value)
	if base64Error != nil {
		return base64Error
	}

	// Dectypt JWT
	block, aesError := aes.NewCipher([]byte(secret))
	err(aesError)
	gcm, ciperError := cipher.NewGCM(block)
	err(ciperError)
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	decrypted, gcmError := gcm.Open(nil, nonce, ciphertext, nil)
	err(gcmError)

	unmarshalError := json.Unmarshal(decrypted, &model)
	if unmarshalError != nil {
		return unmarshalError
	}
	return nil
}

func addCookie(w http.ResponseWriter, name, value string, exp time.Time) {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  exp,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
}
