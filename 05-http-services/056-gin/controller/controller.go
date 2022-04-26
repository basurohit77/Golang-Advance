package controller

import "github.com/gin-gonic/gin"
import "golang-gin/entity"
import "golang-gin/service"

type VedioController interface {
	Save(*gin.Context) entity.Vedio
	FindAll() []entity.Vedio
}

type vediocontroller struct {
	service service.VedioService
}

func New(s service.VedioService) VedioController {
	return &vediocontroller{
		service: s,
	}
}

func (c *vediocontroller) Save(ctx *gin.Context) entity.Vedio {
	var vedio entity.Vedio
	ctx.BindJSON(&vedio)
	c.service.Save(vedio)
	return vedio
}

func (c *vediocontroller) FindAll() []entity.Vedio {
	return c.service.FindAll()
}
