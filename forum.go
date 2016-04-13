package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (this *DB) selectForum(shortName string, rel Related) gin.H {
	var result ClassForum
	this.Map.SelectOne(&result, "SELECT * FROM `forum` WHERE `shortname` = ?", shortName)

	forum := gin.H{"id": result.Id, "name": result.Name, "short_name": result.Shortname, "user": result.User}
	if rel.User {
		forum["user"] = this.selectUser(result.User)
	}
	return forum
}

func (this *DB) forumCreate(c *gin.Context) {
	var result ClassForum
	c.BindJSON(&result)
	this.Map.Exec("INSERT INTO `forum` (name, shortname, userCreatedId, idForum) VALUES (?, ?, ?, ?);", result.Name, result.Shortname, result.User, result.Shortname)
	rel := Related{}
	rel.User = false
	forum := this.selectForum(result.Shortname, rel)
	c.JSON(200, gin.H{"code": 0, "response": forum})
}

func (this *DB) forumDetails(c *gin.Context) {
	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, false, false)
	if err {
		c.JSON(200, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	result := this.selectForum(c.Query("forum"), rel)
	c.JSON(200, gin.H{"code": 0, "response": result})
}

func (this *DB) forumListPosts(c *gin.Context) {
	forumId := c.Query("forum")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since")

	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, true, true)
	if err {
		c.JSON(200, gin.H{"code": 3, "response": "Bad request"})
		return
	}

	var ids []string
	query := "SELECT idPost FROM `post` WHERE `forumParentId` = ?"
	if Since != "" {
		query += " AND `dateU` >= ?"
	}
	if order != "asc" {
		order = "desc"
	}
	query += " ORDER BY `dateU` " + order
	if limit != "" {
		query += " LIMIT " + string(limit)
	}
	query += ";"
	if Since != "" {
		_, err := this.Map.Select(&ids, query, forumId, Since)
		if err != nil {
			log.Error(err)
		}
	} else {
		_, err := this.Map.Select(&ids, query, forumId)
		if err != nil {
			log.Error(err)
		}
	}

	result := make([]gin.H, len(ids))
	for idx, postId := range ids {
		iid, _ := strconv.Atoi(postId)
		result[idx] = this.selectPost(iid, rel)
	}
	c.JSON(200, gin.H{"code": 0, "response": result})
}

func (this *DB) forumListThreads(c *gin.Context) {
	forumId := c.Query("forum")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since")

	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, true, false)
	if err {
		c.JSON(200, gin.H{"code": 3, "response": "Bad request"})
		return
	}

	var ids []string
	query := "SELECT idThread FROM `thread` WHERE `forumParentId` = ?"
	if Since != "" {
		query += " AND `dateU` >= ?"
	}
	if order != "asc" {
		order = "desc"
	}
	query += " ORDER BY `dateU` " + order
	if limit != "" {
		query += " LIMIT " + string(limit)
	}
	query += ";"
	if Since != "" {
		_, err := this.Map.Select(&ids, query, forumId, Since)
		if err != nil {
			log.Error(err)
		}
	} else {
		_, err := this.Map.Select(&ids, query, forumId)
		if err != nil {
			log.Error(err)
		}
	}

	result := make([]gin.H, len(ids))
	for idx, threadId := range ids {
		iid, _ := strconv.Atoi(threadId)
		result[idx] = this.selectThread(iid, rel)
	}
	c.JSON(200, gin.H{"code": 0, "response": result})
}

func (this *DB) forumListUsers(c *gin.Context) {
	forumId := c.Query("forum")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since_id")

	var emails []string
	query := "SELECT DISTINCT userCreatedId FROM `post` JOIN `user` ON userCreatedId = idUser WHERE `forumParentId` = ?"
	if Since != "" {
		query += " AND userCreatedId >= ?"
	}
	if order != "asc" {
		order = "desc"
	}
	query += " ORDER BY `user`.`name` " + order
	if limit != "" {
		query += " LIMIT " + string(limit)
	}
	query += ";"
	if Since != "" {
		_, err := this.Map.Select(&emails, query, forumId, Since)
		if err != nil {
			log.Error(err)
		}
	} else {
		_, err := this.Map.Select(&emails, query, forumId)
		if err != nil {
			log.Error(err)
		}
	}
	result := make([]gin.H, len(emails))
	for idx, email := range emails {
		result[idx] = this.selectUser(email)
	}
	c.JSON(200, gin.H{"code": 0, "response": result})
}
