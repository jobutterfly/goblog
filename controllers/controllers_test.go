package controllers

/* I could optimize these test, declaring repeated functions
and things of the sort, but I will leave it for another day
or project, to improve my testing skills.
*/

import (
	"testing"
	"bytes"
	"strings"
	"io"
	"io/ioutil"
	"os"
	"context"
	"strconv"
	"time"
	"html/template"
	"net/http"
	"net/http/httptest"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/enzdor/goblog/models"
	"github.com/enzdor/goblog/utils"
	"github.com/enzdor/goblog/sqlc"
	"github.com/enzdor/goblog/auth"
	"github.com/joho/godotenv"
)

type getCase struct {
	name	string
	req 	*http.Request
	w	*httptest.ResponseRecorder
	te	*template.Template
	te_data	any
}

func testGetCases(t *testing.T, testCases []getCase, serveFunc func(w http.ResponseWriter, r *http.Request )) {
    for _, tc := range testCases {
	t.Run(tc.name, func(t *testing.T){
	    ts , err := stringTemplate(tc.te, tc.te_data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    serveFunc(tc.w, tc.req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    responseBody, err := ioutil.ReadAll(res.Body) 
	    if err != nil {
		t.Errorf("Expected no error, got %v", err)
	    }

	    if string(responseBody) != ts {
		t.Errorf("Expected data to be equal to ts: %s", string(responseBody))
	    }
	})
    }

}


var Th *Handler

func stringTemplate(tmpl *template.Template, data any) (string, error){
    var buff bytes.Buffer 
    if err := tmpl.ExecuteTemplate(&buff, "layout", data); err != nil {
	return "", err
    }

    return buff.String(), nil
}

func start() error{
    if err := godotenv.Load("../.env"); err != nil {
	return err
    }
    user := os.Getenv("DBUSER")
    pass := os.Getenv("DBPASS")
    name := os.Getenv("TESTDBNAME")
    key := os.Getenv("JWTKEY")

    db := NewDB(user, pass, name)

    if _, err := db.Query("DELETE FROM articles; "); err != nil {
	return err
    }
    if _, err := db.Query("DELETE FROM users; "); err != nil {
	return err
    }

    Th = NewHandler(db, key)

    return nil
}

func testRedirect(path string, h func(w http.ResponseWriter, r *http.Request)) error{
    redirectTestCase := struct {
	req 	*http.Request
	w	*httptest.ResponseRecorder
    } {
	req: httptest.NewRequest(http.MethodGet, path, nil),
	w: httptest.NewRecorder(), 
    }

    Th.ServeIndex(redirectTestCase.w, redirectTestCase.req)
    res := redirectTestCase.w.Result()
    defer res.Body.Close()

    url, err := res.Location()
    if err != nil {
	return err
    }

    if url.Path != "/error/404" {
	return &models.PathError{Message: "expected path to be /error/404 but got" + url.Path}
    }

    return nil
}

func cleanString (s string) string{
    s = strings.ReplaceAll(s, "\n", "")
    s = strings.ReplaceAll(s, "\t", "")
    s = strings.ReplaceAll(s, " ", "")

    return s
}

func TestServeIndex(t *testing.T){
    if err := start(); err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    createThreads := []sqlc.CreateArticleParams{
	{
	    Title: "This is the first title",
	    Content: "This is the first comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
	},
	{
	    Title: "This is the second title",
	    Content: "This is the second comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
	},
	{
	    Title: "This is the third title",
	    Content: "This is the third comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
	},
	{
	    Title: "This is the fourth title",
	    Content: "This is the fourth comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
	},
    }

    // populating db with threads that are going to be queried

    for _, tt := range createThreads {
	_, err := Th.q.CreateArticle(context.Background(), tt)
	if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	}
    }

    articles, err := Th.q.GetArticles(context.Background())
    if err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    testCases := []struct {
	name	string
	req 	*http.Request
	w	*httptest.ResponseRecorder
	te	*template.Template
	te_data	any
    }{
	{
	    name: "index",
	    req: httptest.NewRequest(http.MethodGet, "/", nil),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("index"),
	    te_data: models.IndexData{
		Articles: articles,
	    },
	},
    } 

    for _, tc := range testCases {
	t.Run(tc.name, func(t *testing.T){
	    ts , err := stringTemplate(tc.te, tc.te_data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    Th.ServeIndex(tc.w, tc.req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    responseBody, err := ioutil.ReadAll(res.Body) 
	    if err != nil {
		t.Errorf("Expected no error, got %v", err)
	    }

	    if string(responseBody) != ts {
		t.Errorf("Expected data to be equal to ts: %s", string(responseBody))
	    }
	})
    }

    t.Run("redirect", func(t *testing.T){
	if err := testRedirect("/akldfjk", Th.ServeIndex); err != nil {
	    t.Errorf("expected no error and got %v", err)
	}
    })


}

func TestServeBoard(t *testing.T) {
    if err := start(); err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    createThreads := []sqlc.CreateArticleParams{
	{
	    Title: "This is the first title",
	    Content: "This is the first comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
	},
	{
	    Title: "This is the second title",
	    Content: "This is the second comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
	},
	{
	    Title: "This is the third title",
	    Content: "This is the third comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
	},
	{
	    Title: "This is the fourth title",
	    Content: "This is the fourth comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
	},
    }

    for _, tt := range createThreads {
	_, err := Th.q.CreateArticle(context.Background(), tt)
	if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	}
    }
    
    articles, err := Th.q.GetArticles(context.Background())
    if err != nil {
	t.Errorf("expected no error, got %v", err)
    }


    testCases := []struct {
	name 	string
	id	int
	w	*httptest.ResponseRecorder
    } {
	{
	    name: "article",
	    id: int(articles[0].ArticleID),
	    w: httptest.NewRecorder(), 
	},
    }

    for _, tc := range testCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodGet, "/article/" + strconv.Itoa(tc.id), nil)
	    art, err := Th.q.GetArticle(context.Background(), int32(tc.id))
	    if err != nil {
		t.Errorf("Expected no error, got %v", err)
	    }

	    tmpl, err := utils.ServeArticleTemplate(art.Content)
	    if err != nil {
		t.Errorf("expected no error, got %v", err)
	    }

	    data := models.ArticleData{
		Article: art,
	    }

	    ts, err := stringTemplate(tmpl, data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    Th.ServeArticle(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    responseBody, err := ioutil.ReadAll(res.Body) 
	    if err != nil {
		t.Errorf("Expected no error, got %v", err)
	    }

	    if string(responseBody) != ts {
		t.Errorf("Expected data to be equal to ts: %v", ts)
	    }

	})
    }

    t.Run("redirect", func(t *testing.T){
	if err := testRedirect("/article/fdjladlfkd", Th.ServeArticle); err != nil {
	    t.Errorf("expected no error and got %v", err)
	}
    })
}


func TestServeError(t *testing.T) {
    if err := start(); err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    testCases := []struct{
	name 	string
	status	int
	w	*httptest.ResponseRecorder
	te	*template.Template
    } {
	{
	    name: "not found",
	    status: http.StatusNotFound,
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("error"),
	},
	{
	    name: "internal server error",
	    status: http.StatusInternalServerError,
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("error"),
	},
	{
	    name: "other status",
	    status: http.StatusForbidden,
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("error"),
	},
    }

    for _, tc := range testCases {
	t.Run(tc.name, func(t *testing.T){
	    data := utils.CreateErrorData(tc.status)

	    req := httptest.NewRequest(http.MethodGet, "/error/" + strconv.Itoa(tc.status), nil)

	    ts , err := stringTemplate(tc.te, data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    Th.ServeError(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    responseBody, err := ioutil.ReadAll(res.Body) 
	    if err != nil {
		t.Errorf("Expected no error, got %v", err)
	    }

	    if string(responseBody) != ts {
		t.Errorf("Expected data to be equal to ts: %s", string(responseBody))
	    }
	})
    }
}


func TestServePost(t *testing.T) {
    if err := start(); err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    getTestCases := []struct{
	name 	string
	w	*httptest.ResponseRecorder
	te	*template.Template
	data	models.PostData
    } {
	{
	    name: "get with no errors",
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("post"),
	    data: models.PostData{
		Title: "",
		Content: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: false, Message: "", Field: "content"},
		},
	    },
	},
    }

    for _, tc := range getTestCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodGet, "/post", nil)

	    token, err := auth.NewToken(1)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }
	    req.AddCookie(&http.Cookie{
		Name: "auth",
		Value: token,
		HttpOnly: true,
	    })

	    ts, err := stringTemplate(tc.te, tc.data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    Th.ServePost(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    responseBody, err := ioutil.ReadAll(res.Body) 
	    if err != nil {
		t.Errorf("Expected no error, got %v", err)
	    }

	    if string(responseBody) != ts {
		t.Errorf("expected response to be equal to template string \n%s", string(responseBody))
	    }
	})
    }
//--data-raw 'title=a+new+post&comment=this+is+the+comment+for+the+new+post'

    postTestCases := []struct{
	name 	string
	resPath	string
	body	io.Reader
	w	*httptest.ResponseRecorder
	te	*template.Template
	data	models.PostData
    } {
	{
	    name: "post with no errors",
	    resPath: "/manage",
	    body: bytes.NewReader([]byte("title=a+new+post+with+a+very+interesting+title&content=this+comment+is+too+good+for+you+to+understand")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("post"),
	    data: models.PostData{
		Title: "",
		Content: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: false, Message: "", Field: "content"},
		},
	    },
	},
	{
	    name: "post with errors",
	    resPath: "/post",
	    body: bytes.NewReader([]byte("title=a+new+post+with+a+very+interesting+title&content=")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("post"),
	    data: models.PostData{
		Title: "a new post with a very interesting title",
		Content: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: true, Message: "This field is required", Field: "content"},
		},
	    },
	},
    }
    
    for _, tc := range postTestCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodPost, "/post", tc.body)
	    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	    token, err := auth.NewToken(1)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    req.AddCookie(&http.Cookie{
		Name: "auth",
		Value: token,
		HttpOnly: true,
	    })

	    Th.ServePost(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    ts, err := stringTemplate(tc.te, tc.data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    if tc.name == "post with errors" {
		responseBody := tc.w.Body.String() 

		if cleanString(responseBody) != cleanString(ts) {
		    t.Errorf("expected response to be equal to template string %s \n %s", responseBody, string(ts))
		}
	    } else {
		url, err := res.Location()
		if err != nil {
		    t.Errorf("Expected no errors, got %v", err)
		}

		if url.Path != "/manage" {
		    t.Errorf("expected path to be /manage but got " + url.Path)
		}
	    }

	})
    }
}

func TestServeLogin(t *testing.T) {
    if err := start(); err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    pass := "Ilovepizza"
    name := "John1982"
    hashpass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
    if err != nil {
	t.Errorf("Expected no errors, got %v", err)
    }

    createUser := sqlc.CreateUserParams{
	Name: name,
	Password: string(hashpass),
    }

    if _, err := Th.q.CreateUser(context.Background(), createUser); err != nil {
	t.Errorf("Expected no errors, got %v", err)
    }

    getTestCases := []struct{
	name 	string
	w	*httptest.ResponseRecorder
	te	*template.Template
	data	models.LoginData
    } {
	{
	    name: "get with no errors",
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("login"),
	    data: models.LoginData{
		Name: "",
		Password: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "name"},
		    {Bool: false, Message: "", Field: "password"},
		},
	    },
	},
    }

    for _, tc := range getTestCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodGet, "/login", nil)

	    ts, err := stringTemplate(tc.te, tc.data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    Th.ServeLogin(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    responseBody, err := ioutil.ReadAll(res.Body) 
	    if err != nil {
		t.Errorf("Expected no error, got %v", err)
	    }

	    if string(responseBody) != ts {
		t.Errorf("expected response to be equal to template string \n%s", string(responseBody))
	    }
	})
    }
//--data-raw 'title=a+new+post&comment=this+is+the+comment+for+the+new+post'

    postTestCases := []struct{
	name 	string
	resPath	string
	withE	bool
	body	io.Reader
	w	*httptest.ResponseRecorder
	te	*template.Template
	data	models.LoginData
    } {
	{
	    name: "post with no errors",
	    resPath: "/manage",
	    withE: false,
	    body: bytes.NewReader([]byte("name=" + name + "&password=" + pass)),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("login"),
	    data: models.LoginData{
		Name: "",
		Password: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "name"},
		    {Bool: false, Message: "", Field: "password"},
		},
	    },
	},
	{
	    name: "post with errors, missing field",
	    resPath: "/login",
	    withE: true,
	    body: bytes.NewReader([]byte("name=" + name + "&password=")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("login"),
	    data: models.LoginData{
		Name: name,
		Password: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "name"},
		    {Bool: true, Message: "This field is required", Field: "password"},
		},
	    },
	},
	{
	    name: "post with errors, wrong password",
	    resPath: "/login",
	    withE: true,
	    body: bytes.NewReader([]byte("name=" + name + "&password=joemamma")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("login"),
	    data: models.LoginData{
		Name: name,
		Password: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "name"},
		    {Bool: true, Message: "Invalid credentials", Field: "password"},
		},
	    },
	},
    }
    
    for _, tc := range postTestCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodPost, "/login", tc.body)
	    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	    Th.ServeLogin(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    ts, err := stringTemplate(tc.te, tc.data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    if tc.withE {
		responseBody := tc.w.Body.String()

		if cleanString(string(responseBody)) != cleanString(ts) {
		    t.Errorf("expected response to be equal to ts")
		}
	    } else {
		url, err := res.Location()
		if err != nil {
		    t.Errorf("Expected no error, got %v", err)
		}

		if url.Path != "/manage" {
		    t.Errorf("expected path to be /manage, got %s", url.Path)
		}
	    }

	})
    }
}


func TestServeEdit(t *testing.T) {
    if err := start(); err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    createArticle := sqlc.CreateArticleParams{
	    Title: "This is the first title",
	    Content: "This is the first comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
    }

    _, err := Th.q.CreateArticle(context.Background(), createArticle)
    if err != nil {
	    t.Errorf("Expected no errors, got %v", err)
    }

    articles, err := Th.q.GetArticles(context.Background())
    if err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    article := articles[0]

    getTestCases := []struct{
	name 	string
	w	*httptest.ResponseRecorder
	te	*template.Template
	data	models.EditData
    } {
	{
	    name: "get with no errors",
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("edit"),
	    data: models.EditData{
		Id: int(article.ArticleID),
		Title: article.Title,
		Content: article.Content,
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: false, Message: "", Field: "content"},
		},
	    },
	},
    }

    for _, tc := range getTestCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodGet, "/edit/" + strconv.Itoa(int(article.ArticleID)), nil)

	    token, err := auth.NewToken(1)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }
	    req.AddCookie(&http.Cookie{
		Name: "auth",
		Value: token,
		HttpOnly: true,
	    })

	    ts, err := stringTemplate(tc.te, tc.data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    Th.ServeEdit(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    responseBody, err := ioutil.ReadAll(res.Body) 
	    if err != nil {
		t.Errorf("Expected no error, got %v", err)
	    }

	    if string(responseBody) != ts {
		t.Errorf("expected response to be equal to template string response body\n%s\nts\n%s", string(responseBody), ts)
	    }
	})
    }
//--data-raw 'title=a+new+post&comment=this+is+the+comment+for+the+new+post'

    postTestCases := []struct{
	name 	string
	resPath	string
	body	io.Reader
	w	*httptest.ResponseRecorder
	te	*template.Template
	data	models.EditData
    } {
	{
	    name: "post with no errors",
	    resPath: "/manage",
	    body: bytes.NewReader([]byte("title=a+new+post+with+a+very+interesting+title&content=this+comment+is+too+good+for+you+to+understand")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("edit"),
	    data: models.EditData{
		Id: int(article.ArticleID),
		Title: "",
		Content: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: false, Message: "", Field: "content"},
		},
	    },
	},
	{
	    name: "post with errors",
	    resPath: "/post",
	    body: bytes.NewReader([]byte("title=a+new+post+with+a+very+interesting+title&content=")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("edit"),
	    data: models.EditData{
		Id: int(article.ArticleID),
		Title: "a new post with a very interesting title",
		Content: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: true, Message: "This field is required", Field: "content"},
		},
	    },
	},
    }
    
    for _, tc := range postTestCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodPost, "/edit/" + strconv.Itoa(int(article.ArticleID)), tc.body)
	    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	    token, err := auth.NewToken(1)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    req.AddCookie(&http.Cookie{
		Name: "auth",
		Value: token,
		HttpOnly: true,
	    })

	    Th.ServeEdit(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    ts, err := stringTemplate(tc.te, tc.data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    if tc.name == "post with errors" {
		responseBody := tc.w.Body.String() 

		if cleanString(responseBody) != cleanString(ts) {
		    t.Errorf("expected response to be equal to template string %s \n %s", responseBody, string(ts))
		}
	    } else {
		url, err := res.Location()
		if err != nil {
		    t.Errorf("Expected no errors, got %v", err)
		}

		if url.Path != "/manage" {
		    t.Errorf("expected path to be /manage but got " + url.Path)
		}
	    }

	})
    }
}


func TestServeDelete(t *testing.T) {
    if err := start(); err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    createArticle := sqlc.CreateArticleParams{
	    Title: "This is the first title",
	    Content: "This is the first comment",
	    Date: strconv.Itoa(int(time.Now().Unix())),
    }

    _, err := Th.q.CreateArticle(context.Background(), createArticle)
    if err != nil {
	    t.Errorf("Expected no errors, got %v", err)
    }

    articles, err := Th.q.GetArticles(context.Background())
    if err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    article := articles[0]

    getTestCases := []struct{
	name 	string
	w	*httptest.ResponseRecorder
	te	*template.Template
	data	models.DeleteData
    } {
	{
	    name: "get with no errors",
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("delete"),
	    data: models.DeleteData{
		Article: article,
	    },
	},
    }

    for _, tc := range getTestCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodGet, "/delete/" + strconv.Itoa(int(article.ArticleID)), nil)

	    token, err := auth.NewToken(1)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }
	    req.AddCookie(&http.Cookie{
		Name: "auth",
		Value: token,
		HttpOnly: true,
	    })

	    ts, err := stringTemplate(tc.te, tc.data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    Th.ServeDelete(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    responseBody, err := ioutil.ReadAll(res.Body) 
	    if err != nil {
		t.Errorf("Expected no error, got %v", err)
	    }

	    if string(responseBody) != ts {
		t.Errorf("expected response to be equal to template string response body\n%s\nts\n%s", string(responseBody), ts)
	    }
	})
    }
//--data-raw 'title=a+new+post&comment=this+is+the+comment+for+the+new+post'

    postTestCases := []struct{
	name 	string
	resPath	string
	body	io.Reader
	w	*httptest.ResponseRecorder
	te	*template.Template
	data	models.DeleteData
    } {
	{
	    name: "post with no errors",
	    resPath: "/manage",
	    body: bytes.NewReader([]byte("")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("delete"),
	    data: models.DeleteData{
		Article: article,
	    },
	},
    }

    for _, tc := range postTestCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodPost, "/delete/" + strconv.Itoa(int(article.ArticleID)), tc.body)
	    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	    token, err := auth.NewToken(1)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    req.AddCookie(&http.Cookie{
		Name: "auth",
		Value: token,
		HttpOnly: true,
	    })

	    Th.ServeDelete(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    ts, err := stringTemplate(tc.te, tc.data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    if tc.name == "post with errors" {
		responseBody := tc.w.Body.String() 

		if cleanString(responseBody) != cleanString(ts) {
		    t.Errorf("expected response to be equal to template string %s \n %s", responseBody, string(ts))
		}
	    } else {
		url, err := res.Location()
		if err != nil {
		    t.Errorf("Expected no errors, got %v", err)
		}

		if url.Path != "/manage" {
		    t.Errorf("expected path to be /manage but got " + url.Path)
		}
	    }

	})
    }
}

