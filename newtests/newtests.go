package newtests

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



*/

//--data-raw 'title=a+new+post&comment=this+is+the+comment+for+the+new+post'
    for _, tc := range postTestCases {
	t.Run(tc.name, func(t *testing.T){
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




*/

    for _, tc := range postTestCases {
	t.Run(tc.name, func(t *testing.T){
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




*/



//--data-raw 'title=a+new+post&comment=this+is+the+comment+for+the+new+post'
    for _, tc := range postTestCases {
	t.Run(tc.name, func(t *testing.T){
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


/*



DELETE
DELETE
DELETE
DELETE
DELETE
DELETE
DELETE


*/

    for _, tc := range postTestCases {
	t.Run(tc.name, func(t *testing.T){
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


