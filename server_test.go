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

func postRequest(url string, data []byte) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Origin", config.Domain)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req);

	return resp, err
}

func postRequestToken(url string, data []byte, token string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data)); if err != nil {
		return nil, err
	}
	req.Header.Set("Origin", config.Domain)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", token)
	resp, err := client.Do(req);

	return resp, err	
}

func registerCreds(t *testing.T, admin *ActiveUser) {
	server := httptest.NewServer(postDefense(registerCredentialsHandler()))
	defer server.Close()
	
	data := make(map[string]string)
	data["username"] = "foo"
	data["password"] = "bar"
	data["id"] = fmt.Sprintf("%d", admin.Id)
	data["admin"] = "false"
	res, _ := json.Marshal(data)
	
	resp, err := postRequestToken(server.URL, res, admin.AccessToken); if err != nil {
		t.Fatal("Registering credentials failed", err.Error())
	}
	defer resp.Body.Close()

	checkStatusCode(t, resp, "Registering creds has error")
	checkBody(t, resp)
}

func loginCreds(t *testing.T) *ActiveUser {
	server := httptest.NewServer(originMiddleware(postMiddleware(loginCredentialsHandler())))
	defer server.Close()

	creds := make(map[string]string)
	creds["username"]="shiba"
	creds["password"]="foobar"
	res, _ := json.Marshal(creds)

	resp, err := postRequest(server.URL, res)
	if err != nil {
		t.Fatal("Logging in credentials failed with:", err.Error())
	}
	
	checkStatusCode(t, resp, "Login credentials has error")
	checkBody(t, resp)

	// Check Set-Cookie header
	if resp.Header.Get("Set-Cookie") == "" {
		t.Fatal("Set-Cookie header not set with login response")
	}

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
	server := httptest.NewServer(postDefense(updateUsernameHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = "shiba2"
	data["id"] = fmt.Sprintf("%d", au.Id)
	res, _ := json.Marshal(data)

	resp, err := postRequestToken(server.URL, res, au.AccessToken); if err != nil {
		t.Fatal("Updating username failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Update username has error")
	checkBody(t, resp)
}

func updatePassword(t *testing.T, au *ActiveUser) {
	server := httptest.NewServer(postDefense(updatePasswordHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["new_password"] = "foobar2"
	data["old_password"] = "foobar"
	data["id"] = fmt.Sprintf("%d", au.Id)
	res, _ := json.Marshal(data)

	resp, err := postRequestToken(server.URL, res, au.AccessToken); if err != nil {
		t.Fatal("Updating password failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Update password has error")
	checkBody(t, resp)
}

func adminNewPassword(t *testing.T, admin *ActiveUser, username string) {
	server := httptest.NewServer(postDefense(adminNewPasswordHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = username
	data["id"] = fmt.Sprintf("%d", admin.Id)
	res, _ := json.Marshal(data)

	resp, err := postRequestToken(server.URL, res, admin.AccessToken); if err != nil {
		t.Fatal("Admin new password failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Admin new password has error")
	checkBody(t, resp)

	var data2 map[string]string
	err = json.NewDecoder(resp.Body).Decode(&data2)

	p, ok := data2["password"]; if !ok {
		t.Fatal("No password field in response body")
	}

	if p == "" {
		t.Fatal("New password is empty")
	}

	if len(p) < 10 {
		t.Fatal("New password is less than 10 characters")
	}

	//More tests for password?
	//Currently poor passwords can get through, but I don't want to deal with random tests for now

}

func adminMakeAdmin(t *testing.T, admin *ActiveUser, username string) {
	server := httptest.NewServer(postDefense(adminMakeAdminHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = username
	data["id"] = fmt.Sprintf("%d", admin.Id)
	res, _ := json.Marshal(data)

	resp, err := postRequestToken(server.URL, res, admin.AccessToken); if err != nil {
		t.Fatal("Admin make admin failed with:", err.Error())
	}

	checkStatusCode(t, resp, "Admin make admin error")
	checkBody(t, resp)

	stmt := prepareQuery("sql/check_admin_by_name.sql")
	var b bool
	err = stmt.QueryRow(username).Scan(&b)

	if !b {
		t.Fatal("User is not an admin")
	}
}

func adminRevokeAdmin(t *testing.T, admin *ActiveUser, username string) {
	server := httptest.NewServer(postDefense(adminRevokeAdminHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = username
	data["id"] = fmt.Sprintf("%d", admin.Id)
	res, _ := json.Marshal(data)

	resp, err := postRequestToken(server.URL, res, admin.AccessToken); if err != nil {
		t.Fatal("Admin revoke admin failed with:", err.Error())		
	}

	checkStatusCode(t, resp, "Admin revoke admin error")
	checkBody(t, resp)

	stmt := prepareQuery("sql/check_admin_by_name.sql")
	var b bool
	err = stmt.QueryRow(username).Scan(&b)

	if b {
		t.Fatal("User is an admin")
	}
}

func adminDeleteUser(t *testing.T, admin *ActiveUser, username string) {
	server := httptest.NewServer(postDefense(adminDeleteUserHandler()))
	defer server.Close()

	data := make(map[string]string)
	data["username"] = username
	data["id"] = fmt.Sprintf("%d", admin.Id)
	res, _ := json.Marshal(data)

	resp, err := postRequestToken(server.URL, res, admin.AccessToken); if err != nil {
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

func l(s string) {
	fmt.Println(s)
}

func TestIntegrationApi(t *testing.T) {

	l("Login")
	au := loginCreds(t)
	
	l("Verify")
	verifyToken(t, au.AccessToken)

	l("Update username")
	updateUsername(t, au)

	l("Update password")
	updatePassword(t, au)

	l("Register New User")
	registerCreds(t, au)	

	l("Admin password")
	adminNewPassword(t, au, "foo")

	l("Admin make admin")
	adminMakeAdmin(t, au, "foo")

	l("Admin revoke")
	adminRevokeAdmin(t, au, "foo")

	l("Admin delete")
	adminDeleteUser(t, au, "foo")	
}
