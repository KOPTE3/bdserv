package main

import (
	"github.com/gin-gonic/gin"
)

func (this *DB) commonClear(c *gin.Context) {
	this.Map.Exec("DELETE FROM `follow`;")
	this.Map.Exec("DELETE FROM `subscribe`;")
	this.Map.Exec("DELETE FROM `post`;")
	this.Map.Exec("DELETE FROM `thread`;")
	this.Map.Exec("DELETE FROM `forum`;")
	this.Map.Exec("DELETE FROM `user`;")
	c.JSON(200, gin.H{"code": 0, "response": "OK"})
}

func (this *DB) commonStatus(c *gin.Context) {
	userCount, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `user`")
	postCount, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `post`")
	forumCount, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `forum`")
	threadCount, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `thread`")
	response := gin.H{"code": 0, "response": gin.H{"user": userCount, "thread": threadCount, "forum": forumCount, "post": postCount}}
	c.JSON(200, response)
}
