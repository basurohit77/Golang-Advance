package main

import "net/http"
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

	// server.POST("/save", func(ctx *gin.Context) {
	// 	ctx.JSON(200, vediocontroller.Save(ctx))

	// })
	server.POST("/save", func(ctx *gin.Context) {
		err := vediocontroller.Save(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		} else {
			ctx.JSON(200, gin.H{
				"message": "valid data",
			})
		}
	})

	server.Run(":8080")
}
