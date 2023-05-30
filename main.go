package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if conf.SqsQueue != "" {
		go listenOnQueue()
	}

	router := gin.Default()
	router.StaticFile("/favicon.ico", "./public/favicon.ico")
	router.StaticFile("/", "./public/index.html")
	router.StaticFile("/fonts/Roboto-Thin.ttf", "./public/fonts/Roboto-Thin.ttf")
	router.GET("/state", getState)
	router.POST("/state", postState)
	router.POST("/pause", pause)
	router.POST("/resume", resume)
	router.Run(":8080")
}
