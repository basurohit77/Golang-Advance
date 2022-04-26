package controller

import "net/http"
import "github.com/gin-gonic/gin"
import "golang-gin/entity"
import "golang-gin/service"

type VedioController interface {
	Save(*gin.Context) error
	FindAll() []entity.Vedio
	ShowAll(*gin.Context)
}

type vediocontroller struct {
	service service.VedioService
}

func New(s service.VedioService) VedioController {
	return &vediocontroller{
		service: s,
	}
}

func (c *vediocontroller) Save(ctx *gin.Context) error {
	var vedio entity.Vedio
	err := ctx.ShouldBindJSON(&vedio)
	if err != nil {
		return err
	}
	c.service.Save(vedio)
	return nil
}

func (c *vediocontroller) FindAll() []entity.Vedio {
	return c.service.FindAll()
}

func (c *vediocontroller) ShowAll(ctx *gin.Context) {
	videos := c.service.FindAll()
	data := gin.H{
		"title":  "Video Page",
		"videos": videos,
	}
	ctx.HTML(http.StatusOK, "index.html", data)
}
