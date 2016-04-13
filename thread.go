package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (this *DB) selectThread(id int, rel Related) gin.H {
	var result ClassThread
	this.Map.SelectOne(&result, "SELECT * FROM `thread` WHERE `idThread` = ?;", id)
	var cnt int64 = 0
	if !result.Isdeleted {
		cnt, _ = this.Map.SelectInt("SELECT COUNT(*) FROM `post` WHERE `threadParentId` = ? AND `isDeleted` = false;", id)
	}
	thread := gin.H{"date": result.Date, "dislikes": result.Dislikes, "forum": result.Forum, "id": result.Id, "isClosed": result.Isclosed, "isDeleted": result.Isdeleted, "likes": result.Likes, "message": result.Message, "slug": result.Slug, "title": result.Title, "user": result.User, "points": result.Likes - result.Dislikes, "posts": cnt}
	if rel.User {
		thread["user"] = this.selectUser(result.User)
	}
	if rel.Forum {
		var rrel Related
		rrel.Forum = false
		rrel.User = false
		thread["forum"] = this.selectForum(result.Forum, rrel)
	}
	return thread
}

func (this *DB) threadClose(c *gin.Context) {
	var params struct {
		Id int `json:"thread"`
	}
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `thread` SET `isClosed` = true WHERE `idThread` = ?;", params.Id)
	c.JSON(200, gin.H{"code": 0, "response": params})
}

func (this *DB) threadCreate(c *gin.Context) {
	var thread ClassThread
	c.BindJSON(&thread)
	result, err := this.Map.Exec("INSERT INTO `thread` (dateU, forumParentId, isClosed, isDeleted, message, slug, title, userCreatedId) VALUES (?, ?, ?, ?, ?, ?, ?, ?);", thread.Date, thread.Forum, thread.Isclosed, thread.Isdeleted, thread.Message, thread.Slug, thread.Title, thread.User)
	if err != nil {
		log.Error(err)
		c.JSON(200, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Error(err)
		c.JSON(200, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "response": gin.H{"date": thread.Date, "forum": thread.Forum, "id": id, "isClosed": thread.Isclosed, "isDeleted": thread.Isdeleted, "message": thread.Message, "slug": thread.Slug, "title": thread.Title, "user": thread.User}})
}

func (this *DB) threadDetails(c *gin.Context) {
	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, true, false)
	if err {
		c.JSON(200, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	thread, _ := strconv.Atoi(c.Query("thread"))
	result := this.selectThread(thread, rel)
	c.JSON(200, gin.H{"code": 0, "response": result})
}

func (this *DB) threadList(c *gin.Context) {
	forumId := c.Query("forum")
	userId := c.Query("user")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since")
	var rrel Related
	rrel.Forum = false
	rrel.User = false
	rrel.Thread = false

	if forumId != "" {
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
			result[idx] = this.selectThread(iid, rrel)
		}
		c.JSON(200, gin.H{"code": 0, "response": result})
	} else if userId != "" {
		var ids []string
		query := "SELECT idThread FROM `thread` WHERE `userCreatedId` = ?"
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
			_, err := this.Map.Select(&ids, query, userId, Since)
			if err != nil {
				log.Error(err)
			}
		} else {
			_, err := this.Map.Select(&ids, query, userId)
			if err != nil {
				log.Error(err)
			}
		}

		result := make([]gin.H, len(ids))
		for idx, threadId := range ids {
			iid, _ := strconv.Atoi(threadId)
			result[idx] = this.selectThread(iid, rrel)
		}
		c.JSON(200, gin.H{"code": 0, "response": result})
	} else {
		c.JSON(200, gin.H{"code": 3, "response": "Bad request"})
	}
}

func (this *DB) threadListPosts(c *gin.Context) {
	threadId := c.Query("thread")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since")
	sort := c.Query("sort")
	var rrel Related
	rrel.Forum = false
	rrel.User = false
	rrel.Thread = false

	if sort == "" {
		threadIID, _ := strconv.Atoi(threadId)
		var ids []string
		query := "SELECT idPost FROM `post` WHERE `threadParentId` = ?"
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
			_, err := this.Map.Select(&ids, query, threadIID, Since)
			if err != nil {
				log.Error(err)
			}
		} else {
			_, err := this.Map.Select(&ids, query, threadIID)
			if err != nil {
				log.Error(err)
			}
		}

		result := make([]gin.H, len(ids))
		for idx, postId := range ids {
			iid, _ := strconv.Atoi(postId)
			result[idx] = this.selectPost(iid, rrel)
		}
		c.JSON(200, gin.H{"code": 0, "response": result})
	} else if sort == "tree" {
		threadIID, _ := strconv.Atoi(threadId)
		var ids []string
		query := "SELECT idPost FROM `post` WHERE `threadParentId` = ?"
		if Since != "" {
			query += " AND `dateU` >= ?"
		}
		if order != "asc" {
			order = "desc"
		}
		query += " ORDER BY levelnum " + order + ", `postPath` ASC"
		if limit != "" {
			query += " LIMIT " + string(limit)
		}
		query += ";"
		log.Notice(query)
		if Since != "" {
			_, err := this.Map.Select(&ids, query, threadIID, Since)
			if err != nil {
				log.Error(err)
			}
		} else {
			_, err := this.Map.Select(&ids, query, threadIID)
			if err != nil {
				log.Error(err)
			}
		}
		result := make([]gin.H, len(ids))
		for idx, postId := range ids {
			iid, _ := strconv.Atoi(postId)
			result[idx] = this.selectPost(iid, rrel)
		}
		c.JSON(200, gin.H{"code": 0, "response": result})
	} else if sort == "parent_tree" {
		threadIID, _ := strconv.Atoi(threadId)
		var ids []string
		query := "SELECT idPost FROM `post` WHERE `threadParentId` = ?"
		if order != "asc" {
			order = "desc"
		}
		if limit != "" {
			cnt, _ := strconv.Atoi(limit)
			if order == "asc" {
				query += " AND `levelnum` < " + strconv.Itoa(cnt+1)
			} else {
				maxidx, _ := this.Map.SelectInt("SELECT MAX(levelnum) FROM `post` WHERE `threadParentId` = ?", threadIID)
				query += " AND `levelnum` > " + strconv.Itoa(int(maxidx-int64(cnt)))
			}
		}
		if Since != "" {
			query += " AND `dateU` >= ?"
		}

		query += " ORDER BY levelnum " + order + ", `postPath` ASC"
		query += ";"
		log.Notice(query)
		if Since != "" {
			_, err := this.Map.Select(&ids, query, threadIID, Since)
			if err != nil {
				log.Error(err)
			}
		} else {
			_, err := this.Map.Select(&ids, query, threadIID)
			if err != nil {
				log.Error(err)
			}
		}
		result := make([]gin.H, len(ids))
		for idx, postId := range ids {
			iid, _ := strconv.Atoi(postId)
			result[idx] = this.selectPost(iid, rrel)
		}
		c.JSON(200, gin.H{"code": 0, "response": result})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": "OK"})
	}
}

func (this *DB) threadOpen(c *gin.Context) {
	var params struct {
		Id int `json:"thread"`
	}
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `thread` SET `isClosed` = false WHERE `idThread` = ?;", params.Id)
	c.JSON(200, gin.H{"code": 0, "response": params})
}

func (this *DB) threadRemove(c *gin.Context) {
	var params struct {
		Id int `json:"thread"`
	}
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `thread` SET `isDeleted` = true WHERE `idThread` = ?;", params.Id)
	this.Map.Exec("UPDATE `post` SET `isDeleted` = true WHERE `threadParentId` = ?;", params.Id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": params})
}

func (this *DB) threadRestore(c *gin.Context) {
	var params struct {
		Id int `json:"thread"`
	}
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `thread` SET `isDeleted` = false WHERE `idThread` = ?;", params.Id)
	this.Map.Exec("UPDATE `post` SET `isDeleted` = false WHERE `threadParentId` = ?;", params.Id)
	c.JSON(200, gin.H{"code": 0, "response": params})
}

func (this *DB) threadSubscribe(c *gin.Context) {
	var params ClassSubscribe
	c.BindJSON(&params)
	this.Map.Exec("INSERT INTO `subscribe` (`threadId`, `userId`) VALUES (?, ?);", params.Thread, params.User)
	c.JSON(200, gin.H{"code": 0, "response": params})
}

func (this *DB) threadUnsubscribe(c *gin.Context) {
	var params ClassSubscribe
	c.BindJSON(&params)
	this.Map.Exec("DELETE FROM `subscribe` WHERE `threadId` = ? AND `userId` = ?;", params.Thread, params.User)
	c.JSON(200, gin.H{"code": 0, "response": params})
}

func (this *DB) threadUpdate(c *gin.Context) {
	var params ClassUpdateThread
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `thread` SET `message` = ?, `slug` = ? WHERE `idThread` = ?;", params.Message, params.Slug, params.Thread)
	var rrel Related
	rrel.User = false
	rrel.Forum = false
	c.JSON(200, gin.H{"code": 0, "response": this.selectThread(params.Thread, rrel)})
}

func (this *DB) threadVote(c *gin.Context) {
	var params ClassVote
	c.BindJSON(&params)
	if params.Vote > 0 {
		this.Map.Exec("UPDATE `thread` SET `likes` = `likes` + 1 WHERE `idThread` = ?;", params.Thread)
	} else {
		this.Map.Exec("UPDATE `thread` SET `dislikes` = `dislikes` + 1 WHERE `idThread` = ?;", params.Thread)
	}
	var rrel Related
	rrel.User = false
	rrel.Forum = false
	c.JSON(200, gin.H{"code": 0, "response": this.selectThread(params.Thread, rrel)})
}
