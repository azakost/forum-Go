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
	password TEXT NOT NULL,
	fullname TEXT NOT NULL,
	language TEXT NOT NULL DEFAULT 'en',
	UNIQUE (username, email) );

CREATE TABLE posts (
	postId INTEGER PRIMARY KEY AUTOINCREMENT,
	posted DATETIME DEFAULT CURRENT_TIMESTAMP,
	userId INTEGER NOT NULL,
	title TEXT NOT NULL,
	text TEXT NOT NULL,
	categories TEXT NOT NULL );
	
CREATE TABLE reactions (
	reactionId INTEGER PRIMARY KEY AUTOINCREMENT,
	reacted DATETIME DEFAULT CURRENT_TIMESTAMP,
	postId INTEGER NOT NULL,
	userId INTEGER NOT NULL,
	reaction INTEGER NOT NULL );

CREATE TABLE replies (
	replyId INTEGER PRIMARY KEY AUTOINCREMENT,
	replied DATETIME DEFAULT CURRENT_TIMESTAMP,
	postId INTEGER NOT NULL,
	userId INTEGER NOT NULL,
	reply TEXT NOT NULL );	


INSERT INTO users(email, username, password, fullname) 
	values(
		"azakost@gmail.com",
		"azakost",
		"$2a$04$kitMig4Sfj/Id0C85pysxu3MQbFMC0qXDn5j4RA8ZoI8P9GMcE8Vm",
		"Azamat Alimbayev"
	);
	

	`
