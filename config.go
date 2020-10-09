package main

import "time"

const (
	salt         = "writedrunkeditsober"
	secret       = "bGl2ZWZhc3RkaWV5b3VuZw=="
	tokenLife    = time.Hour
	tokenRefresh = tokenLife / 2
)

const initialQuery = `
CREATE TABLE users (
	userId INTEGER PRIMARY KEY AUTOINCREMENT,
	role TEXT NOT NULL DEFAULT 'user',
	registered DATETIME DEFAULT CURRENT_TIMESTAMP,
	username TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	password TEXT,
	fullname TEXT,
	language TEXT,
	UNIQUE (username, email) );

INSERT INTO users(email, username, password, fullname) 
	values(
		"azakost@gmail.com",
		"azakost",
		"123456",
		"Azamat Alimbayev"
	);
	
CREATE TABLE posts (
	postId INTEGER PRIMARY KEY AUTOINCREMENT,
	posted DATETIME DEFAULT CURRENT_TIMESTAMP,
	userId INTEGER,
	title TEXT,
	text TEXT );
	
	
	`
