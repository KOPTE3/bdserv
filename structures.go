package main

import (
	"gopkg.in/gorp.v1"
)

func convertRelated(entities []string, user, forum, thread bool) (Related, bool) {
	rel := Related{}
	rel.User = false
	err := false
	for _, entity := range entities {
		if entity == "user" && user {
			rel.User = true
		} else if entity == "forum" && forum {
			rel.Forum = true
		} else if entity == "thread" && thread {
			rel.Thread = true
		} else {
			err = true
		}
	}
	return rel, err
}

type DB struct {
	Map *gorp.DbMap
}

type ClassForum struct {
	Id        string `json:"id" db:"idForum"`
	Name      string `json:"name" db:"name"`
	Shortname string `json:"short_name" db:"shortname"`
	User      string `json:"user" db:"userCreatedId"`
}

type ClassThread struct {
	Date      string `json:"date" db:"dateU"`
	Dislikes  int    `json:"dislikes" db:"dislikes"`
	Forum     string `json:"forum" db:"forumParentId"`
	Id        int    `json:"id" db:"idThread"`
	Isclosed  bool   `json:"isClosed" db:"isClosed"`
	Isdeleted bool   `json:"isDeleted" db:"isDeleted"`
	Likes     int    `json:"likes" db:"likes"`
	Message   string `json:"message" db:"message"`
	Points    int    `json:"points"`
	Posts     int    `json:"posts"`
	Slug      string `json:"slug" db:"slug"`
	Title     string `json:"title" db:"title"`
	User      string `json:"user" db:"userCreatedId"`
}

type ClassPost struct {
	Date          string  `json:"date" db:"dateU"`
	Dislikes      int     `json:"dislikes" db:"dislikes"`
	Forum         string  `json:"forum" db:"forumParentId"`
	Id            int     `json:"id" db:"idPost"`
	Isapproved    bool    `json:"isApproved" db:"isApproved"`
	Isdeleted     bool    `json:"isDeleted" db:"isDeleted"`
	Isedited      bool    `json:"isEdited" db:"isEdited"`
	Ishighlighted bool    `json:"isHighlighted" db:"isHighlighted"`
	Isspam        bool    `json:"isSpam" db:"isSpam"`
	Level         int     `db:"level"`
	Levelnum      *int    `db:"levelnum"`
	Likes         int     `json:"likes" db:"likes"`
	Message       string  `json:"message" db:"message"`
	Parent        int     `json:"parent" db:"postParentId"`
	Path          *string `db:"postPath"`
	Points        int     `json:"points" db:"points"`
	Thread        int     `json:"thread" db:"threadParentId"`
	User          string  `json:"user" db:"userCreatedId"`
}

type ClassUser struct {
	About       string `json:"about" db:"about"`
	Email       string `json:"email" binding:"required" db:"idUser"`
	Id          string `json:"id"`
	Isanonymous bool   `json:"isAnonymous" db:"isAnonymous"`
	Name        string `json:"name" db:"name"`
	Username    string `json:"username" db:"username"`
}

type ClassFollow struct {
	Kto  string `json:"follower"`
	Kogo string `json:"followee"`
}

type ClassSubscribe struct {
	User   string `json:"user"`
	Thread int    `json:"thread"`
}

type ClassVote struct {
	Vote   int `json:"vote"`
	Thread int `json:"thread"`
}

type ClassVotePost struct {
	Vote int `json:"vote"`
	Post int `json:"post"`
}

type ClassUpdateUser struct {
	About     string `json:"about"`
	UserEmail string `json:"user"`
	Name      string `json:"name"`
}

type ClassUpdateThread struct {
	Message string `json:"message"`
	Slug    string `json:"slug"`
	Thread  int    `json:"thread"`
}

type ClassUpdatePost struct {
	Message string `json:"message"`
	Post    int    `json:"post"`
}

type Related struct {
	User   bool
	Forum  bool
	Thread bool
}
