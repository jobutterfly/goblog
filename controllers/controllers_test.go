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
	name		string
	req 		*http.Request
	w		*httptest.ResponseRecorder
	te		*template.Template
	te_data		any
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
		t.Errorf("Expected data to be equal to ts: %s\n\n got responseBody: %s", ts, string(responseBody))
	    }
	})
    }
}

type postCase struct {
	name 		string
	withCookie	bool
	withError	bool
	body		io.Reader
	w		*httptest.ResponseRecorder
	te		*template.Template
	te_data		any
}

func testPostCases(t *testing.T, reqPath string, testCases []postCase, serveFunc func(w http.ResponseWriter, r *http.Request )) {
    for _, tc := range testCases {
	t.Run(tc.name, func(t *testing.T){
	    req := httptest.NewRequest(http.MethodPost, reqPath, tc.body)
	    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	    if tc.withCookie {
		token, err := auth.NewToken(1)
		if err != nil {
		    t.Errorf("Expected no errors, got %v", err)
		}

		req.AddCookie(&http.Cookie{
		    Name: "auth",
		    Value: token,
		    HttpOnly: true,
		})
	    }

	    serveFunc(tc.w, req)
	    res := tc.w.Result()
	    defer res.Body.Close()

	    ts , err := stringTemplate(tc.te, tc.te_data)
	    if err != nil {
		t.Errorf("Expected no errors, got %v", err)
	    }

	    if tc.withError {
		responseBody := tc.w.Body.String()

		if cleanString(string(responseBody)) != cleanString(ts) {
		    t.Errorf("expected response: \n%s\n\n to be equal to ts: \n\n %s", cleanString(string(responseBody)), cleanString(ts))
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

    testCases := []getCase {
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

    testGetCases(t, testCases, Th.ServeIndex)

    t.Run("redirect", func(t *testing.T){
	if err := testRedirect("/akldfjk", Th.ServeIndex); err != nil {
	    t.Errorf("expected no error and got %v", err)
	}
    })


}

func TestServeArticle(t *testing.T) {
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

    testCases := []getCase {
	{
	    name: "not found",
	    req: httptest.NewRequest(http.MethodGet, "/error/" + 
		strconv.Itoa(http.StatusNotFound), nil),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("error"),
	    te_data: utils.CreateErrorData(http.StatusNotFound),
	},
	{
	    name: "internal server error",
	    req: httptest.NewRequest(http.MethodGet, "/error/" + 
		strconv.Itoa(http.StatusInternalServerError), nil),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("error"),
	    te_data: utils.CreateErrorData(http.StatusInternalServerError),
	},
	{
	    name: "other status",
	    req: httptest.NewRequest(http.MethodGet, "/error/" + 
		strconv.Itoa(http.StatusForbidden), nil),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("error"),
	    te_data: utils.CreateErrorData(http.StatusForbidden),
	},
    }

    testGetCases(t, testCases, Th.ServeError)
}


func TestServePost(t *testing.T) {
    if err := start(); err != nil {
	t.Errorf("expected no error, got %v", err)
    }

    getReq := httptest.NewRequest(http.MethodGet, "/post", nil);

    token, err := auth.NewToken(1)
    if err != nil {
	t.Errorf("Expected no errors, got %v", err)
    }

    getReq.AddCookie(&http.Cookie{
	Name: "auth",
	Value: token,
	HttpOnly: true,
    })

    getTestCases := []getCase {
	{
	    name: "get with no errors",
	    req: getReq,
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("post"),
	    te_data: models.PostData{
		Title: "",
		Content: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: false, Message: "", Field: "content"},
		},
	    },
	},
    }

    testGetCases(t, getTestCases, Th.ServePost);

//--data-raw 'title=a+new+post&comment=this+is+the+comment+for+the+new+post'

    postTestCases := []postCase {
	{
	    name: "post with no errors",
	    body: bytes.NewReader([]byte("title=a+new+post+with+a+very+interesting+title&content=this+comment+is+too+good+for+you+to+understand")),
	    w: httptest.NewRecorder(), 
	    withCookie: true,
	    withError: false,
	    te: utils.Serve("post"),
	    te_data: models.PostData{
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
	    body: bytes.NewReader([]byte("title=a+new+post+with+a+very+interesting+title&content=")),
	    w: httptest.NewRecorder(), 
	    withCookie: true,
	    withError: true,
	    te: utils.Serve("post"),
	    te_data: models.PostData{
		Title: "a new post with a very interesting title",
		Content: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "title"},
		    {Bool: true, Message: "This field is required", Field: "content"},
		},
	    },
	},
    }

    testPostCases(t, "/post", postTestCases, Th.ServePost)
}


// FIX ME: not working

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

    getTestCases := []getCase {
	{
	    name: "get with no errors",
	    req: httptest.NewRequest(http.MethodGet, "/login", nil),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("login"),
	    te_data: models.LoginData{
		Name: "",
		Password: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "name"},
		    {Bool: false, Message: "", Field: "password"},
		},
	    },
	},
    }

    testGetCases(t, getTestCases, Th.ServePost)

    postTestCases := []postCase {
	{
	    name: "post with no errors",
	    withCookie: false,
	    withError: false,
	    body: bytes.NewReader([]byte("name=" + name + "&password=" + pass)),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("login"),
	    te_data: models.LoginData{
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
	    withCookie: false,
	    withError: true,
	    body: bytes.NewReader([]byte("name=" + name + "&password=")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("login"),
	    te_data: models.LoginData{
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
	    withCookie: false,
	    withError: true,
	    body: bytes.NewReader([]byte("name=" + name + "&password=joemamma")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("login"),
	    te_data: models.LoginData{
		Name: name,
		Password: "",
		Errors: [2]models.FormError{
		    {Bool: false, Message: "", Field: "name"},
		    {Bool: true, Message: "Invalid credentials", Field: "password"},
		},
	    },
	},
    }

    testPostCases(t, "/login", postTestCases, Th.ServePost)
    
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

    getReq := httptest.NewRequest(http.MethodGet, "/edit/" + strconv.Itoa(int(article.ArticleID)), nil);

    token, err := auth.NewToken(1)
    if err != nil {
	t.Errorf("Expected no errors, got %v", err)
    }

    getReq.AddCookie(&http.Cookie{
	Name: "auth",
	Value: token,
	HttpOnly: true,
    })

    getTestCases := []getCase {
	{
	    name: "get with no errors",
	    req: getReq,
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("edit"),
	    te_data: models.EditData{
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

    testGetCases(t, getTestCases, Th.ServeEdit)

//--data-raw 'title=a+new+post&comment=this+is+the+comment+for+the+new+post'

    postTestCases := []postCase {
	{
	    name: "post with no errors",
	    withCookie: true,
	    withError: false,
	    body: bytes.NewReader([]byte("title=a+new+post+with+a+very+interesting+title&content=this+comment+is+too+good+for+you+to+understand")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("edit"),
	    te_data: models.EditData{
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
	    withCookie: true,
	    withError: true,
	    body: bytes.NewReader([]byte("title=a+new+post+with+a+very+interesting+title&content=")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("edit"),
	    te_data: models.EditData{
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

    testPostCases(t, "/edit/" + strconv.Itoa(int(article.ArticleID)), postTestCases, Th.ServeEdit)
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

    getReq := httptest.NewRequest(http.MethodGet, "/delete/" + strconv.Itoa(int(article.ArticleID)), nil)

    token, err := auth.NewToken(1)
    if err != nil {
	t.Errorf("Expected no errors, got %v", err)
    }
    getReq.AddCookie(&http.Cookie{
	Name: "auth",
	Value: token,
	HttpOnly: true,
    })

    getTestCases := []getCase {
	{
	    name: "get with no errors",
	    req: getReq,
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("delete"),
	    te_data: models.DeleteData{
		Article: article,
	    },
	},
    }

    testGetCases(t, getTestCases, Th.ServeDelete)
//--data-raw 'title=a+new+post&comment=this+is+the+comment+for+the+new+post'

    postTestCases := []postCase {
	{
	    name: "post with no errors",
	    withCookie: true,
	    withError: false,
	    body: bytes.NewReader([]byte("")),
	    w: httptest.NewRecorder(), 
	    te: utils.Serve("delete"),
	    te_data: models.DeleteData{
		Article: article,
	    },
	},
    }

    testPostCases(t, "/delete/" + strconv.Itoa(int(article.ArticleID)), postTestCases, Th.ServeDelete)
}


/*

MANAGE
LOGOUT

*/
