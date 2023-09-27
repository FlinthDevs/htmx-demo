package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func DatabaseInit() *gorm.DB {
	dsn := "htmx:htmx@tcp(db:3306)/htmx?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{}) // change the database provider if necessary

	if err != nil {
		panic("Failed to connect to database!")
	}

	database.AutoMigrate(&Post{}) // register Post model

	return database
}

func main() {
	router := gin.Default()

	// Init head urls.
	headUrls := map[string]string{
		"htmx":        "https://unpkg.com/htmx.org@1.9.5",
		"hyperscript": "https://unpkg.com/hyperscript.org@0.9.11",
		"chota":       "https://unpkg.com/chota@latest",
	}

	// Handle locally hosted libraries.
	offline := os.Getenv("OFFLINE")
	if offline == "1" {
		headUrls["chota"] = "/static/css/chota.css"
		headUrls["htmx"] = "/static/js/htmx.js"
		headUrls["hyperscript"] = "/static/js/hyperscript.js"
	}

	// Wait for DB container to be up and running...
	time.Sleep(time.Second)
	db := DatabaseInit()

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/**/*")
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Get index.
	router.GET("/", func(c *gin.Context) {
		var posts []Post
		db.Order("created_at desc").Limit(5).Find(&posts)

		boostedHeader := c.Request.Header["Hx-Boosted"]
		headlessMode := len(boostedHeader) > 0 && boostedHeader[0] == "true"
		c.HTML(http.StatusOK, "index.gohtml", gin.H{"Title": "Main blog", "CurrentPage": "home", "Posts": posts, "NextPage": 2, "headUrls": headUrls, "HeadlessMode": headlessMode})
		// c.HTML(http.StatusOK, "index.gohtml", gin.H{"Title": "Main blog", "CurrentPage": "home", "Posts": posts, "NextPage": 2, "headUrls": headUrls})
	})

	// Get About page.
	router.GET("/about", func(c *gin.Context) {
		boostedHeader := c.Request.Header["Hx-Boosted"]
		headlessMode := len(boostedHeader) > 0 && boostedHeader[0] == "true"
		c.HTML(http.StatusOK, "about.gohtml", gin.H{"Title": "About", "CurrentPage": "about", "headUrls": headUrls, "HeadlessMode": headlessMode})
		// c.HTML(http.StatusOK, "about.gohtml", gin.H{"Title": "About", "CurrentPage": "about", "headUrls": headUrls})
	})

	// Get paginated posts.
	router.GET("/posts/page/:pageNumber", func(c *gin.Context) {
		var posts []Post

		pageNumber, err := strconv.Atoi(c.Param("pageNumber"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		db.Order("created_at desc").Offset(5 * (pageNumber - 1)).Limit(5).Find(&posts)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// Make it less snappy.
		time.Sleep(time.Second * 2)

		// Set nextPage to -1 if there are no more posts.
		nextPage := pageNumber + 1

		if len(posts) < 5 {
			nextPage = -1
		}

		c.HTML(http.StatusOK, "postsGet.gohtml", gin.H{"Posts": posts, "NextPage": nextPage})
	})

	// Get a specific post.
	router.GET("/posts/:id", func(c *gin.Context) {
		var post Post
		err := db.First(&post, c.Param("id")).Error

		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.HTML(http.StatusOK, "postGet.gohtml", gin.H{"Post": post})
	})

	// Get all posts.
	router.GET("/posts", func(c *gin.Context) {
		var posts []Post
		db.Order("created_at desc").Find(&posts)
		c.HTML(http.StatusOK, "postsGet.gohtml", gin.H{"Posts": posts})
	})

	// Creates post.
	router.POST("/posts", func(c *gin.Context) {
		var post Post
		err := c.MustBindWith(&post, binding.FormPost)

		if err != nil {
			fmt.Println(err)
			return
		}

		db.Create(&post)
		c.HTML(http.StatusOK, "postGet.gohtml", gin.H{"Post": post})
	})

	// Get update post form
	router.GET("/posts/:id/edit", func(c *gin.Context) {
		var post Post
		err := db.First(&post, c.Param("id")).Error

		if err != nil {
			fmt.Println(err)
			return
		}

		c.HTML(http.StatusOK, "postEdit.gohtml", gin.H{"Post": post})
	})

	// Updates  a specific post.
	router.PATCH("/posts/:id", func(c *gin.Context) {
		var post Post

		err := db.First(&post, c.Param("id")).Error

		if err != nil {
			fmt.Println(err)
			return
		}

		lagValue, err := strconv.Atoi(c.PostForm("lagValue"))
		if err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(time.Second * time.Duration(lagValue))

		// Update db post from data.
		db.Model(&post).Updates(Post{Title: c.PostForm("title"), Content: c.PostForm("content")})
		c.HTML(http.StatusOK, "postGet.gohtml", gin.H{"Post": post})
	})

	// Delete a specific post.
	router.DELETE("/posts/:id", func(c *gin.Context) {
		var post Post
		db.First(&post, c.Param("id"))
		db.Delete(&post)
		c.Status(http.StatusOK)
	})

	// Get "Create post" modal.
	router.GET("/modal", func(c *gin.Context) {
		c.HTML(http.StatusOK, "modal.gohtml", gin.H{})
	})

	router.Run(":8081")
}
