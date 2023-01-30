package controllers

import (
	"time"
	"strings"
	"context"
	"strconv"
	"net/http"
	"fmt"
	"log"

	"github.com/enzdor/goblog/models"
	"github.com/enzdor/goblog/utils"
	"github.com/enzdor/goblog/sqlc"
	"github.com/enzdor/goblog/auth"
)


func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := utils.Serve("index")
	ps := strings.Split(r.URL.Path, "/")

	if len(ps) > 1 {
	    if ps[1] != "" {
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
		return
	    }
	}

	articles, err := h.q.GetArticles(context.Background())
	if err != nil {
	    log.Fatal(err)
	    /*
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusInternalServerError), http.StatusSeeOther)
	    return
	    */
	}

	data := models.IndexData{
	    Articles: articles,
	}

	tmpl.ExecuteTemplate(w, "layout", data)
}

func (h *Handler) ServeManage(w http.ResponseWriter, r *http.Request) {
	tmpl := utils.Serve("manage")
	ps := strings.Split(r.URL.Path, "/")

	err := auth.Authorizer(h.key)(w, r) 
	if err != nil {
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusUnauthorized), http.StatusSeeOther)
	    return
	}

	if len(ps) > 2 {
	    if ps[2] != "" {
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
		return
	    }
	}

	articles, err := h.q.GetArticles(context.Background())
	if err != nil {
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusInternalServerError), http.StatusSeeOther)
	    return
	}

	data := models.IndexData{
	    Articles: articles,
	}

	tmpl.ExecuteTemplate(w, "layout", data)
}

func (h *Handler) ServeArticle(w http.ResponseWriter, r *http.Request) {
	vs, err := utils.GetPathValues(strings.Split(r.URL.Path, "/"))
	if err != nil {
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
	    return
	}
	id := vs.Id

	article, err := utils.GetArticleData(h.q, int32(id))
	if err != nil {
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusInternalServerError), http.StatusSeeOther)
	    return
	}

	tmpl, err := utils.ServeArticleTemplate(article.Content)
	if err != nil {
	    fmt.Println(err)
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusInternalServerError), http.StatusSeeOther)
	    return
	}
	
	data := models.ArticleData{
	    Article: article,
	}

	tmpl.ExecuteTemplate(w, "layout", data)
	return
}

func (h *Handler) ServePost(w http.ResponseWriter, r *http.Request) {
	tmpl := utils.Serve("post")
	method := r.Method

	err := auth.Authorizer(h.key)(w, r) 
	if err != nil {
	    fmt.Println(err)
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusUnauthorized), http.StatusSeeOther)
	    return
	}

	switch method {
	case "GET":
	    data := models.PostData{
		Title: "",
		Content: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: false, Message: "", Field: "content"},
		},
	    }
	    tmpl.ExecuteTemplate(w, "layout", data)
	    return

	case "POST":
	    errors, err := utils.ValidateArticle(r.FormValue("title"), r.FormValue("content"))
	    if err != nil {
		data := models.PostData{
		    Title: r.FormValue("title"),
		    Content: r.FormValue("content"),
		    Errors: errors,
		}
		tmpl.ExecuteTemplate(w, "layout", data)
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
		return
	    }

	    if _, err := h.q.CreateArticle(context.Background(), sqlc.CreateArticleParams{
		Title: r.FormValue("title"),
		Content: r.FormValue("content"),
		Date: utils.FormatDate(time.Now()),
	    }); err != nil {
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
		return
	    }

	    http.Redirect(w, r, "/manage", http.StatusSeeOther)
	    return
	}
}

func (h *Handler) ServeEdit(w http.ResponseWriter, r *http.Request) {
	tmpl := utils.Serve("edit")
	method := r.Method

	err := auth.Authorizer(h.key)(w, r) 
	if err != nil {
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusUnauthorized), http.StatusSeeOther)
	    return
	}

	vs, err := utils.GetPathValues(strings.Split(r.URL.Path, "/"))
	if err != nil {
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
	    return
	}
	id := vs.Id

	switch method {
	case "GET":

	    article, err := utils.GetArticleData(h.q, int32(id))
	    if err != nil {
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusInternalServerError), http.StatusSeeOther)
		return
	    }

	    data := models.EditData{
		Id: id,
		Title: article.Title,
		Content: article.Content,
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: false, Message: "", Field: "content"},
		},
	    }

	    tmpl.ExecuteTemplate(w, "layout", data)
	    return

	case "POST":

	    errors, err := utils.ValidateArticle(r.FormValue("title"), r.FormValue("content"))
	    if err != nil {
		data := models.EditData{
		    Id: id,
		    Title: r.FormValue("title"),
		    Content: r.FormValue("content"),
		    Errors: errors,
		}
		tmpl.ExecuteTemplate(w, "layout", data)
		return
	    }

	    if _, err := h.q.EditArticle(context.Background(), sqlc.EditArticleParams{
		Title: r.FormValue("title"),
		Content: r.FormValue("content"),
		ArticleID: int32(id),
	    }); err != nil {
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
		return
	    }

	    http.Redirect(w, r, "/manage", http.StatusSeeOther)
	    return
	}
}

func (h *Handler) ServeDelete(w http.ResponseWriter, r *http.Request) {
	tmpl := utils.Serve("delete")
	method := r.Method

	err := auth.Authorizer(h.key)(w, r) 
	if err != nil {
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusUnauthorized), http.StatusSeeOther)
	    return
	}

	vs, err := utils.GetPathValues(strings.Split(r.URL.Path, "/"))
	if err != nil {
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
	    return
	}
	id := vs.Id

	switch method {
	case "GET":

	    article, err := utils.GetArticleData(h.q, int32(id))
	    if err != nil {
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusInternalServerError), http.StatusSeeOther)
		return
	    }

	    data := models.DeleteData{
		Article: article,
	    }

	    tmpl.ExecuteTemplate(w, "layout", data)
	    return

	case "POST":

	    if _, err := h.q.DeleteArticle(context.Background(), int32(id)); err != nil {
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
		return
	    }

	    http.Redirect(w, r, "/manage", http.StatusSeeOther)
	    return
	}
}

func (h *Handler) ServeLogin(w http.ResponseWriter, r *http.Request) {
	tmpl := utils.Serve("login")
	method := r.Method

	err := auth.Authorizer(h.key)(w, r) 
	if err == nil {
	    http.Redirect(w, r, "/manage", http.StatusSeeOther)
	    return
	}

	switch method {
	case "GET":
	    data := models.LoginData{
		Name: "",
		Password: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "name"},
		    {Bool: false, Message: "", Field: "password"},
		},
	    }
	    tmpl.ExecuteTemplate(w, "layout", data)
	    return

	case "POST":

	    if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusInternalServerError), http.StatusSeeOther)
		return
	    }

	    errors, err := utils.ValidateLogin(r.FormValue("name"), r.FormValue("password"))
	    if err != nil {
		data := models.LoginData{
		    Name: r.FormValue("name"),
		    Password: "",
		    Errors: errors,
		}
		tmpl.ExecuteTemplate(w, "layout", data)
		return
	    }

	    user, err := h.q.GetUserName(context.Background(), r.FormValue("name"))
	    if err != nil {
		data := models.LoginData{
		    Name: r.FormValue("name"),
		    Password: "",
		    Errors: [2]models.FormError{
			{
			    Bool: true,
			    Message: "User not found",
			    Field: "name",
			},
			{
			    Bool: false,
			    Message: "",
			    Field: "password",
			},
		    },
		}
		tmpl.ExecuteTemplate(w, "layout", data)
		return
	    }

	    if err := utils.CheckPassword(user.Password, r.FormValue("password"));
	    err != nil {
		data := models.LoginData{
		    Name: r.FormValue("name"),
		    Password: "",
		    Errors: [2]models.FormError{
			{
			    Bool: false,
			    Message: "",
			    Field: "name",
			},
			{
			    Bool: true,
			    Message: "Invalid credentials",
			    Field: "password",
			},
		    },
		}
		tmpl.ExecuteTemplate(w, "layout", data)
	    }

	    token, err := auth.NewToken(1)
	    if err != nil {
		http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusInternalServerError), http.StatusSeeOther)
		return
	    }

	    http.SetCookie(w, &http.Cookie{
		Name: "auth",
		Value: token,
		HttpOnly: true,
	    })

	    http.Redirect(w, r, "/manage", http.StatusSeeOther)
	}
}

func (h *Handler) ServeLogout(w http.ResponseWriter, r *http.Request) {
	token, err := auth.NewToken(-1)
	if err != nil {
	    http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusInternalServerError), http.StatusSeeOther)
	    return
	}

	http.SetCookie(w, &http.Cookie{
	    Name: "auth",
	    Value: token,
	    HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

func (h *Handler) ServeError(w http.ResponseWriter, r *http.Request) {
    tmpl := utils.Serve("error")

    vs, err := utils.GetPathValues(strings.Split(r.URL.Path, "/"))
    if err != nil {
	http.Redirect(w, r, "/error/" + strconv.Itoa(http.StatusNotFound), http.StatusSeeOther)
	return
    }
    status := vs.Id

    data := utils.CreateErrorData(status)

    tmpl.ExecuteTemplate(w, "layout", data)
}










