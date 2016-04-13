package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (this *DB) selectUser(email string) gin.H {
	var result ClassUser
	var kto, kogo []string
	var subscriptions []int
	err := this.Map.SelectOne(&result, "SELECT * FROM user WHERE idUser = ?", email)
	if err != nil {
		log.Warning(err)
	}
	this.Map.Select(&kto, "SELECT ktoId FROM follow WHERE kogoId = ?", email)
	this.Map.Select(&kogo, "SELECT kogoId FROM follow WHERE ktoId = ?", email)
	this.Map.Select(&subscriptions, "SELECT threadId FROM subscribe WHERE userId = ?", email)
	if result.Isanonymous {
		return gin.H{"email": result.Email, "followers": kto, "following": kogo, "id": result.Email, "isAnonymous": result.Isanonymous, "subscriptions": subscriptions, "name": nil, "about": nil, "username": nil}
	} else {
		return gin.H{"email": result.Email, "followers": kto, "following": kogo, "id": result.Email, "isAnonymous": result.Isanonymous, "subscriptions": subscriptions, "name": result.Name, "about": result.About, "username": result.Username}
	}
}

func (this *DB) userCreate(c *gin.Context) {
	result := ClassUser{}
	eror := c.BindJSON(&result)
	if eror != nil {
		log.Warning(eror)
	}
	var err error
	_, err = this.Map.Exec("INSERT INTO user (idUser, about, isAnonymous, name, username) VALUES (?, ?, ?, ?, ?)", result.Email, result.About, result.Isanonymous, result.Name, result.Username)
	if err != nil {
		log.Warning(err)
		c.JSON(200, gin.H{"code": 5, "response": "User already exists"})
	} else {
		userDetails := this.selectUser(result.Email)
		c.JSON(200, gin.H{"code": 0, "response": userDetails})
	}
}

func (this *DB) userDetails(c *gin.Context) {
	userDetails := this.selectUser(c.Query("user"))
	c.JSON(200, gin.H{"code": 0, "response": userDetails})
}

func (this *DB) userFollow(c *gin.Context) {
	var data ClassFollow
	c.BindJSON(&data)
	this.Map.Exec("INSERT INTO `follow` (ktoId, kogoId) VALUES (?, ?);", data.Kto, data.Kogo)

	userDetails := this.selectUser(data.Kto)
	c.JSON(200, gin.H{"code": 0, "response": userDetails})
}

func (this *DB) userListFollowers(c *gin.Context) {
	userEmail := c.Query("user")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since_id")

	var emails []string
	query := "SELECT ktoId FROM `follow` JOIN `user` ON `follow`.`ktoId` = `user`.`idUser` WHERE kogoId = ?"
	if Since != "" {
		query += " AND idUser >= ?"
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
		_, err := this.Map.Select(&emails, query, userEmail, Since)
		if err != nil {
			log.Error(err)
		}
	} else {
		_, err := this.Map.Select(&emails, query, userEmail)
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

func (this *DB) userListFollowing(c *gin.Context) {
	userEmail := c.Query("user")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since_id")

	var emails []string
	query := "SELECT kogoId FROM `follow` JOIN `user` ON `follow`.`kogoId` = `user`.`idUser` WHERE ktoId = ?"
	if Since != "" {
		query += " AND idUser >= ?"
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
		_, err := this.Map.Select(&emails, query, userEmail, Since)
		if err != nil {
			log.Error(err)
		}
	} else {
		_, err := this.Map.Select(&emails, query, userEmail)
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

func (this *DB) userListPosts(c *gin.Context) {
	userEmail := c.Query("user")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since_id")

	var ids []string
	query := "SELECT idPost FROM `post` WHERE `userCreatedId` = ?"
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
		_, err := this.Map.Select(&ids, query, userEmail, Since)
		if err != nil {
			log.Error(err)
		}
	} else {
		_, err := this.Map.Select(&ids, query, userEmail)
		if err != nil {
			log.Error(err)
		}
	}

	result := make([]gin.H, len(ids))
	var rrel Related
	rrel.Forum = false
	rrel.User = false
	rrel.Thread = false
	for idx, postId := range ids {
		iid, _ := strconv.Atoi(postId)
		result[idx] = this.selectPost(iid, rrel)
	}
	c.JSON(200, gin.H{"code": 0, "response": result})
}

func (this *DB) userUnfollow(c *gin.Context) {
	var data ClassFollow
	c.BindJSON(&data)
	this.Map.Exec("DELETE FROM `follow` WHERE ktoId = ? AND kogoId = ?;", data.Kto, data.Kogo)
	userDetails := this.selectUser(data.Kto)
	c.JSON(200, gin.H{"code": 0, "response": userDetails})
}

func (this *DB) userUpdateProfile(c *gin.Context) {
	var data ClassUpdateUser
	c.BindJSON(&data)
	this.Map.Exec("UPDATE `user` SET about = ?, name = ? WHERE idUser = ?;", data.About, data.Name, data.UserEmail)
	userDetails := this.selectUser(data.UserEmail)
	c.JSON(200, gin.H{"code": 0, "response": userDetails})
}
