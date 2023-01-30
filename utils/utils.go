package utils

import (
    "strconv"
    "net/http"
    "context"
    "strings"
    "html/template"
    "path/filepath"
    "log"
    "fmt"
    "time"

    "golang.org/x/crypto/bcrypt"

    "github.com/enzdor/goblog/models"
    "github.com/enzdor/goblog/sqlc"
)

func Serve(page string) *template.Template {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", page + ".html")

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
	    log.Fatal(err)
	}

	return tmpl
}

func ValidateArticle(title string, content string) ([2]models.FormError, error) {
    errors := [2]models.FormError{
	{
	    Bool: false,
	    Message: "",
	    Field: "title",
	},
	{
	    Bool: false,
	    Message: "",
	    Field: "content",
	},
    }

    if strings.TrimSpace(title) == "" {
	errors[0] = models.FormError{
	    Bool: true,
	    Message: "This field is required",
	    Field: "title",
	}

    }

    if strings.TrimSpace(content) == "" {
	errors[1] = models.FormError{
	    Bool: true,
	    Message: "This field is required",
	    Field: "content",
	}

    }

    if errors[0].Bool || errors[1].Bool {
	err := &models.ValidateError{Message: "One of the fields has not passed the required validation rules."}
	return errors, err
    }

    return errors, nil
}

func ValidateLogin(name string, password string) ([2]models.FormError, error) {
    errors := [2]models.FormError{
	{
	    Bool: false,
	    Message: "",
	    Field: "name",
	},
	{
	    Bool: false,
	    Message: "",
	    Field: "password",
	},
    }

    if strings.TrimSpace(name) == "" {
	errors[0] = models.FormError{
	    Bool: true,
	    Message: "This field is required",
	    Field: "name",
	}

    }

    if strings.TrimSpace(password) == "" {
	errors[1] = models.FormError{
	    Bool: true,
	    Message: "This field is required",
	    Field: "password",
	}

    }

    if errors[0].Bool || errors[1].Bool {
	err := &models.ValidateError{Message: "One of the fields has not passed the required validation rules."}
	return errors, err
    }

    return errors, nil
}

func GetArticleData(queries *sqlc.Queries, id int32) (sqlc.Article, error){
    erra := sqlc.Article{
	Title: "",
	Content: "",
	Date: FormatDate(time.Now()),
    }

    article, err := queries.GetArticle(context.Background(), id)
    if err != nil {
	return erra, err
    }

    return article, nil
}

func CreateErrorData(status int) models.ErrorData{
    switch status {
    case http.StatusNotFound:
	return models.ErrorData{
	    Status: status,
	    Message: "Not found",
	}
    case http.StatusUnauthorized:
	return models.ErrorData{
	    Status: status,
	    Message: "Unauthorized",
	}
    default:
	return models.ErrorData{
	    Status: http.StatusInternalServerError,
	    Message: "Internal server error",
	}
    }
}

type PathInfo struct {
    Id int
}

func GetPathValues(ps []string) (PathInfo, error){
    r := PathInfo{
	Id: 0,
    }

    if len(ps) > 3 {
	if ps[3] != "" {
	    err := &models.PathError{Message: "Not found"}
	    return r, err
	}
    }

    id, err := strconv.Atoi(ps[2])
    if err != nil {
	err := &models.PathError{Message: "Not integer"}
	return r, err
    }
    r.Id = id

    return r, err
}

func ServeArticleTemplate(s string) (*template.Template, error) {
    var res []string
    ss := strings.Split(s, "\n")
    for _, a := range ss {
	a = strings.TrimSpace(a)
	if a != "" {
	    a = "<p class=\"par\">" + a + "</p>"
	}
	res = append(res, a)
    }

    r := ""

    for _, i := range res {
	r = r + i
    }

    r = "{{ define \"content\" }} \n" + r + "\n{{ end }}"

    lp := filepath.Join("templates", "layout.html")
    ap := filepath.Join("templates", "article.html")
    tmpl, err := template.New("layout").ParseFiles(ap, lp)
    if err != nil {
	return tmpl, err
    }

    tmpl.New("content").Parse(r)

    return tmpl, nil
}

func HashPassword(password string) (string, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
	return "", err
    }

    return string(hashedPassword), nil
}

func CheckPassword(hashedPassword string, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func FormatDate(t time.Time) string {
    y, m, d := t.Date()

    return fmt.Sprintf("%d-%d-%d", y, m, d)    
}









