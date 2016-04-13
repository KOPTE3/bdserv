package main

import (
	"database/sql"
	"encoding/json"
	"os"

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
