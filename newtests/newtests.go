package newtests


/*


INDEX
INDEX
INDEX
INDEX
INDEX
INDEX
INDEX
INDEX


we can


*/


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



/*



ARTICLE
ARTICLE
ARTICLE
ARTICLE
ARTICLE
ARTICLE
ARTICLE
ARTICLE
ARTICLE



probably not able to use the same function
because serveArticle uses the function
serveArticleTemplate which none of the other
handlers use


*/



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





/*

ERROR
ERROR
ERROR
ERROR
ERROR
ERROR
ERROR
ERROR
ERROR
ERROR


we can

*/





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


/*




POST
POST
POST
POST
POST
POST
POST
POST
POST
POST

we can


*/



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


/*



LOGIN
LOGIN
LOGIN
LOGIN
LOGIN
LOGIN
LOGIN
LOGIN
LOGIN
LOGIN


we can


*/




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


/*


EDIT
EDIT
EDIT
EDIT
EDIT
EDIT
EDIT
EDIT
EDIT
EDIT
EDIT
EDIT


we can


*/





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



/*



DELETE
DELETE
DELETE
DELETE
DELETE
DELETE
DELETE

we can

*/



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
