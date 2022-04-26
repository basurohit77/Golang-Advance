package main

import "os"
import "io"
import "github.com/gin-gonic/gin"
import "golang-gin/service"
import "golang-gin/controller"
import "golang-gin/middlewares"
import gindump "github.com/tpkeeper/gin-dump"

var (
	vedioservice    service.VedioService       = service.New()
	vediocontroller controller.VedioController = controller.New(vedioservice)
)

func setupLogOutput() {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}

func main() {
	setupLogOutput()
	//server := gin.Default()
	//OR
	//server := gin.New()
	//server.Use(gin.Recovery())
	//server.Use(gin.Logger())
	//or
	//server.Use(gin.Recovery(), gin.Logger())

	server := gin.New()
	server.Use(gin.Recovery(), middlewares.Logger(),
		middlewares.BasicAuth(), gindump.Dump())

	// server.GET("/all", func(ctx *gin.Context) {
	// 	ctx.JSON(200, gin.H{
	// 		"message": "HI",
	// 	})
	// })
	server.GET("/all", func(ctx *gin.Context) {
		ctx.JSON(200, vediocontroller.FindAll())
	})

	server.POST("/save", func(ctx *gin.Context) {
		ctx.JSON(200, vediocontroller.Save(ctx))

	})

	server.Run(":8080")
}
