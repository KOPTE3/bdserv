package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/op/go-logging"
	"gopkg.in/gorp.v1"
)

var log = logging.MustGetLogger("main")
var format = logging.MustStringFormatter(`%{color} %{shortfunc} ▶ %{level:.5s} %{id:03x}%{color:reset} %{message}`)

type Configuration struct {
	SERVPORT string
	DBUSER   string
	DBPASS   string
	DBNAME   string
	DBPORT   string
}

func loadConfig() Configuration {
	log.Info("Загружаем конфиг")
	file, err := os.Open("config.json")
	if err != nil {
		log.Critical(err)
		panic(err)
	}
	decoder := json.NewDecoder(file)
	config := Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Critical(err)
		panic(err)
	}
	log.Notice("Конфиг загружен!")
	return config
}

func openConnection(config *Configuration) *DB {
	var connectionString = config.DBUSER + ":" + config.DBPASS + "@tcp(localhost:" + config.DBPORT + ")/" + config.DBNAME
	log.Info("connect to " + connectionString)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Critical(err)
		panic(err)
	}
	log.Notice("Подключено!")
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Encoding: "utf8", Engine: "InnoDB"}}
	return &DB{Map: dbmap}
}

func main() {
	logging.SetFormatter(format)
	config := loadConfig()
	actions := openConnection(&config)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	common := router.Group("/db/api/")
	{
		common.POST("clear/", actions.commonClear)
		common.GET("status/", actions.commonStatus)
	}

	forum := router.Group("/db/api/forum/")
	{
		forum.POST("create/", actions.forumCreate)
		forum.GET("details/", actions.forumDetails)
		forum.GET("listPosts/", actions.forumListPosts)
		forum.GET("listThreads/", actions.forumListThreads)
		forum.GET("listUsers/", actions.forumListUsers)
	}

	thread := router.Group("/db/api/thread/")
	{
		thread.POST("close/", actions.threadClose)
		thread.POST("create/", actions.threadCreate)
		thread.GET("details/", actions.threadDetails)
		thread.GET("list/", actions.threadList)
		thread.GET("listPosts/", actions.threadListPosts)
		thread.POST("open/", actions.threadOpen)
		thread.POST("remove/", actions.threadRemove)
		thread.POST("restore/", actions.threadRestore)
		thread.POST("subscribe/", actions.threadSubscribe)
		thread.POST("unsubscribe/", actions.threadUnsubscribe)
		thread.POST("update/", actions.threadUpdate)
		thread.POST("vote/", actions.threadVote)
	}

	post := router.Group("/db/api/post/")
	{
		post.POST("create/", actions.postCreate)
		post.GET("details/", actions.postDetails)
		post.GET("list/", actions.postList)
		post.POST("remove/", actions.postRemove)
		post.POST("restore/", actions.postRestore)
		post.POST("update/", actions.postUpdate)
		post.POST("vote/", actions.postVote)
	}

	user := router.Group("/db/api/user/")
	{
		user.POST("create/", actions.userCreate)
		user.GET("details/", actions.userDetails)
		user.POST("follow/", actions.userFollow)
		user.GET("listFollowers/", actions.userListFollowers)
		user.GET("listFollowing/", actions.userListFollowing)
		user.GET("listPosts/", actions.userListPosts)
		user.POST("unfollow/", actions.userUnfollow)
		user.POST("updateProfile/", actions.userUpdateProfile)
	}

	runError := router.Run(":" + config.SERVPORT)
	if runError != nil {
		log.Fatal(runError)
		panic(runError)
	}
}

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

func (this *DB) commonClear(c *gin.Context) {
	this.Map.Exec("DELETE FROM `follow`;")
	this.Map.Exec("DELETE FROM `subscribe`;")
	this.Map.Exec("DELETE FROM `post`;")
	this.Map.Exec("DELETE FROM `thread`;")
	this.Map.Exec("DELETE FROM `forum`;")
	this.Map.Exec("DELETE FROM `user`;")
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": "OK"})
}

func (this *DB) commonStatus(c *gin.Context) {
	userCount, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `user`")
	postCount, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `post`")
	forumCount, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `forum`")
	threadCount, _ := this.Map.SelectInt("SELECT COUNT(*) FROM `thread`")
	response := gin.H{"code": 0, "response": gin.H{"user": userCount, "thread": threadCount, "forum": forumCount, "post": postCount}}
	c.JSON(http.StatusOK, response)
}

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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": forum})
}

func (this *DB) forumDetails(c *gin.Context) {
	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, false, false)
	if err {
		c.JSON(http.StatusOK, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	result := this.selectForum(c.Query("forum"), rel)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
}

func (this *DB) forumListPosts(c *gin.Context) {
	forumId := c.Query("forum")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since")

	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, true, true)
	if err {
		c.JSON(http.StatusOK, gin.H{"code": 3, "response": "Bad request"})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
}

func (this *DB) forumListThreads(c *gin.Context) {
	forumId := c.Query("forum")
	limit := c.Query("limit")
	order := c.Query("order")
	Since := c.Query("since")

	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, true, false)
	if err {
		c.JSON(http.StatusOK, gin.H{"code": 3, "response": "Bad request"})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
}

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
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": gin.H{"date": post.Date, "forum": post.Forum, "id": id, "isApproved": post.Isapproved, "isDeleted": post.Isdeleted, "isEdited": post.Isedited, "isHighlighted": post.Ishighlighted, "isSpam": post.Isspam, "message": post.Message, "parent": nil, "thread": post.Thread, "user": post.User}})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": gin.H{"date": post.Date, "forum": post.Forum, "id": id, "isApproved": post.Isapproved, "isDeleted": post.Isdeleted, "isEdited": post.Isedited, "isHighlighted": post.Ishighlighted, "isSpam": post.Isspam, "message": post.Message, "parent": post.Parent, "thread": post.Thread, "user": post.User}})
	}
}

func (this *DB) postDetails(c *gin.Context) {
	post, _ := strconv.Atoi(c.Query("post"))
	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, true, true)
	if err {
		c.JSON(http.StatusOK, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	result := this.selectPost(post, rel)
	if result != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 1, "response": "Post not found"})
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 3, "response": "Bad request"})
	}
}

func (this *DB) postRemove(c *gin.Context) {
	var params struct {
		Id int `json:"post"`
	}
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `post` SET `isDeleted` = true WHERE `idPost` = ?;", params.Id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": params})
}

func (this *DB) postRestore(c *gin.Context) {
	var params struct {
		Id int `json:"post"`
	}
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `post` SET `isDeleted` = false WHERE `idPost` = ?;", params.Id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": params})
}

func (this *DB) postUpdate(c *gin.Context) {
	var params ClassUpdatePost
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `post` SET `message` = ? WHERE `idPost` = ?;", params.Message, params.Post)
	var rrel Related
	rrel.User = false
	rrel.Forum = false
	rrel.Thread = false
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": this.selectPost(params.Post, rrel)})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": this.selectThread(params.Post, rrel)})
}

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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": params})
}

func (this *DB) threadCreate(c *gin.Context) {
	var thread ClassThread
	c.BindJSON(&thread)
	result, err := this.Map.Exec("INSERT INTO `thread` (dateU, forumParentId, isClosed, isDeleted, message, slug, title, userCreatedId) VALUES (?, ?, ?, ?, ?, ?, ?, ?);", thread.Date, thread.Forum, thread.Isclosed, thread.Isdeleted, thread.Message, thread.Slug, thread.Title, thread.User)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusOK, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusOK, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": gin.H{"date": thread.Date, "forum": thread.Forum, "id": id, "isClosed": thread.Isclosed, "isDeleted": thread.Isdeleted, "message": thread.Message, "slug": thread.Slug, "title": thread.Title, "user": thread.User}})
}

func (this *DB) threadDetails(c *gin.Context) {
	entities := c.Request.URL.Query()["related"]
	rel, err := convertRelated(entities, true, true, false)
	if err {
		c.JSON(http.StatusOK, gin.H{"code": 3, "response": "Bad request"})
		return
	}
	thread, _ := strconv.Atoi(c.Query("thread"))
	result := this.selectThread(thread, rel)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 3, "response": "Bad request"})
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": params})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": params})
}

func (this *DB) threadSubscribe(c *gin.Context) {
	var params ClassSubscribe
	c.BindJSON(&params)
	this.Map.Exec("INSERT INTO `subscribe` (`threadId`, `userId`) VALUES (?, ?);", params.Thread, params.User)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": params})
}

func (this *DB) threadUnsubscribe(c *gin.Context) {
	var params ClassSubscribe
	c.BindJSON(&params)
	this.Map.Exec("DELETE FROM `subscribe` WHERE `threadId` = ? AND `userId` = ?;", params.Thread, params.User)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": params})
}

func (this *DB) threadUpdate(c *gin.Context) {
	var params ClassUpdateThread
	c.BindJSON(&params)
	this.Map.Exec("UPDATE `thread` SET `message` = ?, `slug` = ? WHERE `idThread` = ?;", params.Message, params.Slug, params.Thread)
	var rrel Related
	rrel.User = false
	rrel.Forum = false
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": this.selectThread(params.Thread, rrel)})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": this.selectThread(params.Thread, rrel)})
}

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
		c.JSON(http.StatusOK, gin.H{"code": 5, "response": "User already exists"})
	} else {
		userDetails := this.selectUser(result.Email)
		c.JSON(http.StatusOK, gin.H{"code": 0, "response": userDetails})
	}
}

func (this *DB) userDetails(c *gin.Context) {
	userDetails := this.selectUser(c.Query("user"))
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": userDetails})
}

func (this *DB) userFollow(c *gin.Context) {
	var data ClassFollow
	c.BindJSON(&data)
	this.Map.Exec("INSERT INTO `follow` (ktoId, kogoId) VALUES (?, ?);", data.Kto, data.Kogo)

	userDetails := this.selectUser(data.Kto)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": userDetails})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
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
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": result})
}

func (this *DB) userUnfollow(c *gin.Context) {
	var data ClassFollow
	c.BindJSON(&data)
	this.Map.Exec("DELETE FROM `follow` WHERE ktoId = ? AND kogoId = ?;", data.Kto, data.Kogo)
	userDetails := this.selectUser(data.Kto)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": userDetails})
}

func (this *DB) userUpdateProfile(c *gin.Context) {
	var data ClassUpdateUser
	c.BindJSON(&data)
	this.Map.Exec("UPDATE `user` SET about = ?, name = ? WHERE idUser = ?;", data.About, data.Name, data.UserEmail)
	userDetails := this.selectUser(data.UserEmail)
	c.JSON(http.StatusOK, gin.H{"code": 0, "response": userDetails})
}
