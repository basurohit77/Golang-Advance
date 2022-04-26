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

	server.Static("/css", "./templates/css")
	server.LoadHTMLGlob("templates/*.html")

	apiRoutes := server.Group("/api")
	{
		apiRoutes.GET("/all", func(ctx *gin.Context) {
			ctx.JSON(200, vediocontroller.FindAll())
		})

		apiRoutes.POST("/save", func(ctx *gin.Context) {
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
	}

	// viewRoutes := server.Group("/view")
	// {
	// 	viewRoutes.GET("/videos", func(ctx *gin.Context) {
	// 		vediocontroller.ShowAll(ctx)
	// 	})
	// }
	viewRoutes := server.Group("/view")
	{
		viewRoutes.GET("/videos", vediocontroller.ShowAll)
	}

	server.Run(":8080")
}
