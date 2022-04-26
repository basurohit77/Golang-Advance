package main

import "github.com/gin-gonic/gin"
import "golang-gin/service"
import "golang-gin/controller"

var (
	vedioservice    service.VedioService       = service.New()
	vediocontroller controller.VedioController = controller.New(vedioservice)
)

func main() {
	server := gin.Default()

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
