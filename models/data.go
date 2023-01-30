package models

import (
    "github.com/enzdor/goblog/sqlc"
)

type ErrorData struct {
	Status int
	Message string
}

type IndexData struct {
	Articles []sqlc.Article
}

type ManageData struct {
	Articles []sqlc.Article
}

type ArticleData struct {
	Article sqlc.Article
}

type PostData struct {
	Title string
	Content string
	Errors [2]FormError
}

type EditData struct {
	Id int
	Title string
	Content string
	Errors [2]FormError
}

type DeleteData struct {
	Article sqlc.Article
}

type LoginData struct {
	Name string
	Password string
	Errors [2]FormError
}








