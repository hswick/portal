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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	q.Add("user_id", fmt.Sprintf("%d", 2))
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req); if err != nil {
		t.Fatal("Verifying token failed with", err.Error())
	}
	defer resp.Body.Close()

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

func updateUsername(t *testing.T, au *ActiveUser) {
	server := httptest.NewServer(http.HandlerFunc(updateUsernameHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = "shiba2"
	data["id"] = fmt.Sprintf("%d", au.Id)
	data["access_token"] = au.AccessToken
	res, _ := json.Marshal(data)

	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(res)); if err != nil {
		t.Fatal("Updating username failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Update username has error")
	checkBody(t, resp)
}

func updatePassword(t *testing.T, au *ActiveUser) {
	server := httptest.NewServer(http.HandlerFunc(updatePasswordHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["new_password"] = "foobar2"
	data["old_password"] = "foobar"
	data["id"] = fmt.Sprintf("%d", au.Id)
	data["access_token"] = au.AccessToken
	res, _ := json.Marshal(data)

	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(res)); if err != nil {
		t.Fatal("Updating password failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Update password has error")
	checkBody(t, resp)
}

func adminNewPassword(t *testing.T, admin *ActiveUser, username string) {
	server := httptest.NewServer(http.HandlerFunc(adminNewPasswordHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = username
	data["access_token"] = admin.AccessToken
	res, _ := json.Marshal(data)

	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(res)); if err != nil {
		t.Fatal("Admin new password failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Admin new password has error")
	checkBody(t, resp)

	var data2 map[string]string
	err = json.NewDecoder(resp.Body).Decode(&data2)

	p, ok := data2["password"]; if !ok {
		t.Fatal("No password field in response body")
	}

	if p != "supersecure" {
		t.Fatal("Expecting password to be 'supersecure'")
	}
}

func adminMakeAdmin(t *testing.T, admin *ActiveUser, username string) {
	server := httptest.NewServer(http.HandlerFunc(adminMakeAdminHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = username
	data["access_token"] = admin.AccessToken
	res, _ := json.Marshal(data)

	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(res)); if err != nil {
		t.Fatal("Admin make admin failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Admin make admin error")
	checkBody(t, resp)

	var b bool
	err = db.QueryRow("SELECT admin FROM users WHERE name = $1").Scan(&b)

	if !b {
		t.Fatal("User is not an admin")
	}
}

func adminRevokeAdmin(t *testing.T, admin *ActiveUser, username string) {
	server := httptest.NewServer(http.HandlerFunc(adminRevokeAdminHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = username
	data["access_token"] = admin.AccessToken
	res, _ := json.Marshal(data)

	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(res)); if err != nil {
		t.Fatal("Admin revoke admin failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Admin revoke admin error")
	checkBody(t, resp)

	var b bool
	err = db.QueryRow("SELECT admin FROM users WHERE name = $1").Scan(&b)

	if b {
		t.Fatal("User is an admin")
	}
}

func adminDeleteUser(t *testing.T, admin *ActiveUser, username string) {
	server := httptest.NewServer(http.HandlerFunc(adminDeleteUserHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = username
	data["access_token"] = admin.AccessToken
	res, _ := json.Marshal(data)

	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(res)); if err != nil {
		t.Fatal("Admin delete user failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Admin delete has error")
	checkBody(t, resp)

	var n int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&n)
	if n > 1 {
		t.Fatal("There can only be one")
	}
}

func TestIntegrationApi(t *testing.T) {
	registerCreds(t)
	au := loginCreds(t)
	verifyToken(t, au.AccessToken)

	updateUsername(t, au)
	updatePassword(t, au)

	adminNewPassword(t, au, "foo")
	adminMakeAdmin(t, au, "foo")
	adminRevokeAdmin(t, au, "foo")
	adminDeleteUser(t, au, "foo")	
}
