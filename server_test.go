package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"fmt"
)

func checkBody(t *testing.T, r *http.Response) {
	if r.Body == nil {
		t.Fatal("Response body is nil")
	}
}

func checkStatusCode(t *testing.T, r *http.Response, message string) {
	if r.StatusCode != 200 {
		body, _ := ioutil.ReadAll(r.Body)
		t.Fatal(message, string(body))		
	}
}

func registerCreds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(registerCredentialsHandler()))
	defer server.Close()
	
	creds := make(map[string]string)
	creds["username"] = "shiba"
	creds["password"] = "foobar"

	res, _ := json.Marshal(creds)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(res)); if err != nil {
		t.Fatal("Registering credentials failed", err.Error())
	}

	checkStatusCode(t, resp, "Registering creds has error")
	checkBody(t, resp)	
}

func loginCreds(t *testing.T) *ActiveUser {
	server := httptest.NewServer(http.HandlerFunc(loginCredentialsHandler()))
	defer server.Close()

	creds := make(map[string]string)
	creds["username"]="shiba"
	creds["password"]="foobar"
	res, _ := json.Marshal(creds)

	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(res)); if err != nil {
		t.Fatal("Logging in credentials failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Login credentials has error")
	checkBody(t, resp)

	var au ActiveUser
	err = json.NewDecoder(resp.Body).Decode(&au); if err != nil {
		t.Fatal("Decoding active user failed", err.Error())
	}

	if au.Name != "shiba" {
		t.Fatal("Active user name is incorrect")
	}

	return &au
}

func verifyToken(t *testing.T, token string) {
	server := httptest.NewServer(http.HandlerFunc(verifyTokenHandler))
	defer server.Close()

	secret := "supersecret"

	req, _ := http.NewRequest("GET", server.URL, nil)
	q := req.URL.Query()
	q.Add("access_token", token)
	q.Add("secret", secret)
	q.Add("app_name", "canban")
	q.Add("user_id", fmt.Sprintf("%d", 1))
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req); if err != nil {
		t.Fatal("Verifying token failed with", err.Error())
	}

	checkStatusCode(t, resp, "Verifying token failed with")
	checkBody(t, resp)	

	var data map[string]string
	err = json.NewDecoder(resp.Body).Decode(&data); if err != nil {
		t.Fatal("Verifying token failed with", err.Error())
	}

	message, ok := data["message"]; if !ok {
		t.Fatal("Message field missing from response body")
	}

	if message != "Authorized" {
		t.Fatal("Access token is not authorized")
	}
}

func TestIntegrationApi(t *testing.T) {
	registerCreds(t)
	au := loginCreds(t)
	verifyToken(t, au.AccessToken)
}
