package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (this *DB) selectPost(id int, rel Related) gin.H {
	var result ClassPost
	err := this.Map.SelectOne(&result, "SELECT * FROM `post` WHERE `idPost` = ?;", id)
	if err != nil {
		log.Warning(err)
		return nil
	}
	post := gin.H{"date": result.Date, "dislikes": result.Dislikes, "forum": result.Forum, "id": result.Id, "isApproved": result.Isapproved, "isDeleted": result.Isdeleted, "isEdited": result.Isedited, "isHighlighted": result.Ishighlighted, "isSpam": result.Isspam, "likes": result.Likes, "message": result.Message, "parent": result.Parent, "points": result.Likes - result.Dislikes, "thread": result.Thread, "user": result.User}
	var rrel Related
	rrel.Forum = false
	rrel.Thread = false
	rrel.User = false

	if rel.User {
		post["user"] = this.selectUser(result.User)
	}
	if rel.Thread {
		post["thread"] = this.selectThread(result.Thread, rrel)
	}
	if rel.Forum {
		post["forum"] = this.selectForum(result.Forum, rrel)
	}
	if post["parent"] == -666 {
		post["parent"] = nil
	}
	return post
}

func (this *DB) postCreate(c *gin.Context) {
	var post ClassPost
	post.Parent = -666
	c.BindJSON(&post)
	var id int64
	if post.Parent == -666 {
		cnt, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `post` WHERE `level` = 0 AND `threadParentId` = ?;", post.Thread)
		cnt += 1
		path := strconv.Itoa(int(cnt)) + "#"
		if cnt < 10 {
			path = "000000000" + path
		} else if cnt < 100 {
			path = "00000000" + path
		} else if cnt < 1000 {
			path = "0000000" + path
		} else if cnt < 10000 {
			path = "000000" + path
		} else if cnt < 100000 {
			path = "00000" + path
		} else if cnt < 1000000 {
			path = "0000" + path
		} else if cnt < 10000000 {
			path = "000" + path
		} else {
			path = "00" + path
		}
		result, _ := this.Map.Exec("INSERT INTO `post` (dateU, forumParentId, isApproved, isDeleted, isEdited, isHighlighted, isSpam, message, postParentId, threadParentId, userCreatedId, levelnum, postPath) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", post.Date, post.Forum, post.Isapproved, post.Isdeleted, post.Isedited, post.Ishighlighted, post.Isspam, post.Message, -666, post.Thread, post.User, cnt, path)
		id, _ = result.LastInsertId()
	} else {
		lnum, _ := this.Map.SelectInt("SELECT level + 1 FROM `post` WHERE `idPost` = ?;", post.Parent)
		levelnum, _ := this.Map.SelectInt("SELECT levelnum FROM `post` WHERE `idPost` = ?;", post.Parent)
		cnt, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `post` WHERE `level` = ? AND `postParentId` = ?;", lnum, post.Parent)
		ppath, _ := this.Map.SelectStr("SELECT `postPath` FROM `post` WHERE `idPost` = ?;", post.Parent)
		cnt += 1
		npath := strconv.Itoa(int(cnt)) + "#"
		if cnt < 10 {
			npath = "000000000" + npath
		} else if cnt < 100 {
			npath = "00000000" + npath
		} else if cnt < 1000 {
			npath = "0000000" + npath
		} else if cnt < 10000 {
			npath = "000000" + npath
		} else if cnt < 100000 {
			npath = "00000" + npath
		} else if cnt < 1000000 {
			npath = "0000" + npath
		} else if cnt < 10000000 {
			npath = "000" + npath
		} else {
			npath = "00" + npath
		}
		path := ppath + npath
		result, _ := this.Map.Exec("INSERT INTO `post` (dateU, forumParentId, isApproved, isDeleted, isEdited, isHighlighted, isSpam, message, postParentId, threadParentId, userCreatedId, level, levelnum, postPath) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", post.Date, post.Forum, post.Isapproved, post.Isdeleted, post.Isedited, post.Ishighlighted, post.Isspam, post.Message, post.Parent, post.Thread, post.User, lnum, levelnum, path)
		id, _ = result.LastInsertId()
	}

	if post.Parent == -666 {
		c.JSON(200, gin.H{"code": 0, "response": gin.H{"date": post.Date, "forum": post.Forum, "id": id, "isApproved": post.Isapproved, "isDeleted": post.Isdeleted, "isEdited": post.Isedited, "isHighlighted": post.Ishighlighted, "isSpam": post.Isspam, "message": post.Message, "parent": nil, "thread": post.Thread, "user": post.User}})
	} else {
		c.JSON(200, gin.H{"code": 0, "response": gin.H{"date": post.Date, "forum": post.Forum, "id": id, "isApproved": post.Isapproved, "isDeleted": post.Isdeleted, "isEdited": post.Isedited, "isHighlighted": post.Ishighlighted, "isSpam": post.Isspam, "message": post.Message, "parent": post.Parent, "thread": post.Thread, "user": post.User}})
	}
}

func (this *DB) postDetails(c *gin.Context) {
	post, _ := strconv.Atoi(c.Query("post"))
	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, true, true)
	if err {
		c.JSON(200, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	result := this.selectPost(post, rel)
	if result != nil {
		c.JSON(200, gin.H{"code": 0, "response": result})
	} else {
		c.JSON(200, gin.H{"code": 1, "response": "Post not found"})
	}
}

func (this *DB) postList(c *gin.Context) {
	forumId := c.Query("forum")
	threadId := c.Query("thread")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since")
	var rrel Related
	rrel.Forum = false
	rrel.User = false
	rrel.Thread = false

	if forumId != "" {
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
			result[idx] = this.selectPost(iid, rrel)
		}
		c.JSON(200, gin.H{"code": 0, "response": result})
	} else if threadId != "" {
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
	} else {
		c.JSON(200, gin.H{"code": 3, "response": "Bad request"})
	}
}

func (this *DB) postRemove(c *gin.Context) {
	var params struct {
		Id int `json:"post"`
	}
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `post` SET `isDeleted` = true WHERE `idPost` = ?;", params.Id)
	c.JSON(200, gin.H{"code": 0, "response": params})
}

func (this *DB) postRestore(c *gin.Context) {
	var params struct {
		Id int `json:"post"`
	}
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `post` SET `isDeleted` = false WHERE `idPost` = ?;", params.Id)
	c.JSON(200, gin.H{"code": 0, "response": params})
}

func (this *DB) postUpdate(c *gin.Context) {
	var params ClassUpdatePost
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `post` SET `message` = ? WHERE `idPost` = ?;", params.Message, params.Post)
	var rrel Related
	rrel.User = false
	rrel.Forum = false
	rrel.Thread = false
	c.JSON(200, gin.H{"code": 0, "response": this.selectPost(params.Post, rrel)})
}

func (this *DB) postVote(c *gin.Context) {
	var params ClassVotePost
	c.BindJSON(&params)
	if params.Vote > 0 {
		this.Map.Exec("UPDATE `post` SET `likes` = `likes` + 1 WHERE `idPost` = ?;", params.Post)
	} else {
		this.Map.Exec("UPDATE `post` SET `dislikes` = `dislikes` + 1 WHERE `idPost` = ?;", params.Post)
	}
	var rrel Related
	rrel.User = false
	rrel.Forum = false
	rrel.Thread = false
	c.JSON(200, gin.H{"code": 0, "response": this.selectThread(params.Post, rrel)})
}
