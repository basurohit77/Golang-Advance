package entity

type Person struct {
	FirstName string `json:"firstname" binding:"required"`
	LastName  string `json:"lastname" binding:"required"`
	Age       int8   `json:"age" binding:"gte=1,lte=120"`
	Email     string `json:"email" binding:"required,email"`
}

type Vedio struct {
	Title       string `json:"title" binding:"min=2,max=200"`
	Description string `json:"description" binding:"max=200"`
	URL         string `json:"url" binding:"required,url"`
	Author      Person `json:"author" binding:"required"`
}
