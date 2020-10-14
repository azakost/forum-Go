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
	status INTEGER NOT NULL DEFAULT 123,
	UNIQUE (username, email) );

INSERT INTO users(email, username, password, fullname) 
	values(
		'azakost@gmail.com',
		'azakost',
		'$2a$04$kitMig4Sfj/Id0C85pysxu3MQbFMC0qXDn5j4RA8ZoI8P9GMcE8Vm',
		'Azamat Alimbayev'
	);

INSERT INTO users(email, username, password, fullname) 
	values(
		'geradot@gmail.com',
		'udot',
		'$2a$04$kitMig4Sfj/Id0C85pysxu3MQbFMC0qXDn5j4RA8ZoI8P9GMcE8Vm',
		'German Udotov'
	);

CREATE TABLE categories (
	categoryId INTEGER PRIMARY KEY AUTOINCREMENT,
	created DATETIME DEFAULT CURRENT_TIMESTAMP,
	name TEXT NOT NULL,
	description TEXT NOT NULL );

INSERT INTO categories(name, description)
	values(
		'golang',
		'This category is for gophers!'
	);

INSERT INTO categories(name, description)
	values(
		'js',
		'JavaScript is a mother of all web devs!'
	);

CREATE TABLE posts (
	postId INTEGER PRIMARY KEY AUTOINCREMENT,
	posted DATETIME DEFAULT CURRENT_TIMESTAMP,
	userId INTEGER NOT NULL,
	title TEXT NOT NULL,
	text TEXT NOT NULL,
	categories TEXT NOT NULL,
	status INTEGER DEFAULT 1 );

INSERT INTO posts(userId, title, text, categories) 
	values(
		'1',
		'The best title ever!',
		'This is a very long text written for testing purposes!',
		'"1","2"'
	);

INSERT INTO posts(userId, title, text, categories) 
	values(
		'1',
		'Hello!',
		'sddsd',
		'"2"'
	);
		

CREATE TABLE reactions (
	reactionId INTEGER PRIMARY KEY AUTOINCREMENT,
	reacted DATETIME DEFAULT CURRENT_TIMESTAMP,
	postId INTEGER NOT NULL,
	userId INTEGER NOT NULL,
	reaction INTEGER );


INSERT INTO reactions(postId, userId, reaction) 
	values('1',	'1', '1');

INSERT INTO reactions(postId, userId, reaction) 
	values('1',	'2', '0');

INSERT INTO reactions(postId, userId, reaction) 
	values('2',	'2', '0');

CREATE TABLE replies (
	replyId INTEGER PRIMARY KEY AUTOINCREMENT,
	replied DATETIME DEFAULT CURRENT_TIMESTAMP,
	postId INTEGER NOT NULL,
	userId INTEGER NOT NULL,
	reply TEXT NOT NULL );	

INSERT INTO replies(postId, userId, reply) 
	values('1',	'2', 'Sad! Not for udots!');

INSERT INTO replies(postId, userId, reply) 
	values('1',	'1', 'Best content!');


`
