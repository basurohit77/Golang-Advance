package service

import "golang-gin/entity"

//import "fmt"

type VedioService interface {
	Save(entity.Vedio) entity.Vedio
	FindAll() []entity.Vedio
}

type vedioservice struct {
	vedios []entity.Vedio
}

func New() VedioService {
	return &vedioservice{}
}

func (vs *vedioservice) Save(v entity.Vedio) entity.Vedio {
	var dv entity.Vedio
	dv.URL = "Url Not Found"
	if v.URL != "" {
		vs.vedios = append(vs.vedios, v)
		//return v
	}
	return dv
}

func (vs *vedioservice) FindAll() []entity.Vedio {
	return vs.vedios
}
