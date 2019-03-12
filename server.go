package main

import (
	"log"
	"database/sql"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"html/template"
	"strconv"
	"fmt"
	"time"
	"github.com/BurntSushi/toml"
	"github.com/robfig/cron"
	_ "github.com/lib/pq"
)

func dbConnection() *sql.DB {

	tomlData, err := ioutil.ReadFile("db.toml"); if err != nil {
		log.Fatal(err.Error())
	}
	
	var conn map[string]string
	_, err = toml.Decode(string(tomlData), &conn); if err != nil {
		log.Fatal(err.Error())
	}

	driver, ok := conn["driver"]; if !ok {
		log.Fatal("db.toml missing driver field")
	}

	user, ok := conn["user"]; if !ok {
		log.Fatal("db.toml missing user field")
	}

	password, ok := conn["password"]; if !ok {
		log.Fatal("db.toml missing password field")
	}

	dbname, ok := conn["dbname"]; if !ok {
		log.Fatal("db.toml missing dbname field")
	}

	connection := fmt.Sprintf("user=%s password=%s dbname=%s", user, password, dbname)
	
	db, err := sql.Open(driver, connection); if err != nil {
		log.Fatal(err.Error())
	}

	return db
}

var db *sql.DB = dbConnection()

func prepareQuery(filename string) *sql.Stmt {
	content, err := ioutil.ReadFile(filename); if err != nil {
		log.Fatal(err.Error())
	}

	stmt, err := db.Prepare(string(content)); if err != nil {
		log.Fatal(err.Error())
	}

	return stmt
}

type Apps struct{
	Map map[string]string
	List []string
}

func loadApps() *Apps {
	tomlData, err := ioutil.ReadFile("apps.toml"); if err != nil {
		log.Fatal(err.Error())
	}
	
	var apps map[string]string
	_, err = toml.Decode(string(tomlData), &apps); if err != nil {
		log.Fatal(err.Error())
	}

	appNames := make([]string, 0)
	for k, _ := range apps {
		appNames = append(appNames, k)
	}

	return &Apps{
		Map: apps,
		List: appNames,
	}
}

func (a *Apps) Get(app string) (string, bool) {
	v, ok := a.Map[app]
	return v, ok
}

var apps *Apps = loadApps()

type User struct{
	Id int64
	Name string
}

type ActiveUser struct{
	Id int64 `json:"id"`
	AccessToken string `json:"accessToken"`
	Name string `json:"name"`
	LoginAt time.Time
}

func (a *ActiveUser) Expired(now time.Time) bool {
	dif := a.LoginAt.Sub(now)
	return dif.Hours() < 2//TODO: Should be a configurable parameter
}

type ActiveUsers map[string]*ActiveUser

var activeUsers ActiveUsers = make(ActiveUsers)

func (a ActiveUsers) GarbageCollect() {
	c := cron.New()
	c.AddFunc("@every 2h", func() {//TODO: Should be a configurable parameter
		now := time.Now()
		for token, user := range a {
			if user.Expired(now) {
				delete(a, token)
			}
		}
	})
	c.Start()
}

func verifyAccessToken(token string) bool {
	au, ok := activeUsers[token]
	return ok && au.Expired(time.Now())
}

func verifyUserAccess(token string, id int64) bool {
	au, ok := activeUsers[token]; if !ok {
		return false
	}

	if au.Id != id {
		return false
	}

	return true
}

func activateUser(user *User) *ActiveUser {
	var token string
	au := &ActiveUser{
		Id: user.Id,
		Name: user.Name,
		AccessToken: token,
		LoginAt: time.Now(),
	}
	activeUsers[token] = au
	return au
}

type Welcome struct{
	ActiveUser *ActiveUser
	Apps []string
}

func welcomePageHandler() func(http.ResponseWriter, *http.Request) {
	t, err := template.ParseFiles("./static/welcome.html"); if err != nil {
		log.Fatal(err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		
		q := r.URL.Query()
		
		if q["access_token"] == nil || q["user_id"] == nil {
			http.Error(w, "Must include access_token and user_id in query params to access this page", 401)
			return			
			
		}

		id, err := strconv.ParseInt(q["user_id"][0], 10, 64); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		
		if !verifyAccessToken(q["access_token"][0]) {
			http.Error(w, "Access token is unauthorized", 401)
			return
		}

		if !verifyUserAccess(q["access_token"][0], id) {
			http.Error(w, "Acccess token unauthorized for user", 401)
			return
		}

		//Not checking for nil value, since redundant with above verification
		au, _ := activeUsers[q["access_token"][0]]
		t.Execute(w, &Welcome{
			ActiveUser: au,
			Apps: apps.List,
		})
	}
}

type Credentials struct{
	UserName string `json:"username"`
	Password string `json:"password"`
}

func loginCredentialsHandler() func(http.ResponseWriter, *http.Request) {
	
	stmt := prepareQuery("sql/check_login_credentials.sql")
	
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "This route only accepts POST requests", 400)
			return
		}
		
		if r.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
		}

		var creds Credentials
		err := json.NewDecoder(r.Body).Decode(&creds); if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		var u User
		err = stmt.QueryRow(creds.UserName, creds.Password).Scan(&u.Id, &u.Name); if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		au := activateUser(&u)
		json.NewEncoder(w).Encode(&au)
	}
}

func registerCredentialsHandler() func(http.ResponseWriter, *http.Request) {
	
	stmt := prepareQuery("sql/new_user_credentials.sql")
	stmt2 := prepareQuery("sql/check_admin.sql")
	
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "This route only accepts POST requests", 400)
			return
		}
		
		if r.Body == nil  {
			http.Error(w, "Please send a request body", 400)
			return
		}

		var data map[string]string
		err := json.NewDecoder(r.Body).Decode(&data); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if !verifyAccessToken(data["access_token"]) {
			http.Error(w, "Access token is unauthorized", 401)
			return
		}		

		id, err := strconv.ParseInt(data["id"], 10, 64); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if !verifyUserAccess(data["access_token"], id) {
			http.Error(w, "Access token is not authorized for user", 401)
			return
		}

		var admin bool
		err = stmt2.QueryRow(id).Scan(&admin); if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		newAdmin := false
		if data["admin"] == "true" {
			newAdmin = true
		}

		_, err = stmt.Exec(data["username"], data["password"], newAdmin); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "%s", "New user has been registered")
	}
}

func verifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	
	q := r.URL.Query()
	if q["access_token"] == nil || q["user_id"] == nil || q["secret"] == nil || q["app_name"] == nil {
		http.Error(w, "Must include access_token, user_id, app_name, and secret in query params to access this page", 401)
		return	
	}

	secret, ok := apps.Get(q["app_name"][0]); if !ok {
		http.Error(w, "App name is unrecognized", 401)
		return
	}

	if secret != q["secret"][0] {
		http.Error(w, "Incorrect secret for application", 401)
		return
	}

	if !verifyAccessToken(q["access_token"][0]) {
		http.Error(w, "Access token is unauthorized", 401)
		return
	}

	id, err := strconv.ParseInt(q["user_id"][0], 10, 64); if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !verifyUserAccess(q["access_token"][0], id) {
		http.Error(w, "Access token is not authorized for user", 401)
		return
	}

	data := make(map[string]string)
	data["message"]="Authorized"
	json.NewEncoder(w).Encode(&data)
}

func updatePasswordHandler() func(http.ResponseWriter, *http.Request) {
	stmt := prepareQuery("sql/get_password.sql")
	stmt2 := prepareQuery("sql/update_user_password.sql")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This route only accepts POST requests", 400)
			return
		}
		
		if r.Body == nil  {
			http.Error(w, "Please send a request body", 400)
			return
		}

		var data map[string]string
		err := json.NewDecoder(r.Body).Decode(&data); if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		//verify access token
		if !verifyAccessToken(data["access_token"]) {
			http.Error(w, "Access token is unauthorized", 401)
			return
		}


		id, err := strconv.ParseInt(data["id"], 10, 64); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		
		if !verifyUserAccess(data["access_token"], id) {
			http.Error(w, "Access token is not authorized for user", 401)
			return
		}
		
		//get password
		var password string
		err = stmt.QueryRow(id).Scan(&password); if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		if password != data["old_password"] {
			http.Error(w, "Old password is incorrect", 401)
			return
		}

		_, err = stmt2.Exec(id, data["new_password"]); if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
	}
}

func updateUsernameHandler() func(http.ResponseWriter, *http.Request) {
	stmt := prepareQuery("sql/update_user_name.sql")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This route only accepts POST requests", 400)
			return
		}
		
		if r.Body == nil  {
			http.Error(w, "Please send a request body", 400)
			return
		}

		var data map[string]string
		err := json.NewDecoder(r.Body).Decode(&data); if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		//verify access token
		if !verifyAccessToken(data["access_token"]) {
			http.Error(w, "Access token is unauthorized", 401)
			return
		}

		id, err := strconv.ParseInt(data["id"], 10, 64); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		
		if !verifyUserAccess(data["access_token"], id) {
			http.Error(w, "Access token is not authorized for user", 401)
			return
		}

		_, err = stmt.Exec(id, data["username"]); if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
	}
}

func adminNewPasswordHandler() func(http.ResponseWriter, *http.Request) {
	stmt := prepareQuery("sql/check_admin.sql")
	stmt2 := prepareQuery("sql/update_other_user_password.sql")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This route only accepts POST requests", 400)
			return
		}
		
		if r.Body == nil  {
			http.Error(w, "Please send a request body", 400)
			return
		}

		var data map[string]string
		err := json.NewDecoder(r.Body).Decode(&data); if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		//verify access token
		if !verifyAccessToken(data["access_token"]) {
			http.Error(w, "Access token is unauthorized", 401)
			return
		}

		id, err := strconv.ParseInt(data["id"], 10, 64); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var admin bool
		err = stmt.QueryRow(id).Scan(&admin); if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		if !admin {
			http.Error(w, "User is not an admin. Unauthorized action.", 401)
			return
		}

		newPassword := "supersecure"

		_, err = stmt2.Exec(data["username"], newPassword); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var body map[string]string = make(map[string]string)
		body["password"] = newPassword
		json.NewEncoder(w).Encode(&body)
	}
}

func adminMakeAdminHandler() func(http.ResponseWriter, *http.Request) {
	stmt := prepareQuery("sql/check_admin.sql")	
	stmt2 := prepareQuery("sql/update_admin.sql")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This route only accepts POST requests", 400)
			return
		}
		
		if r.Body == nil  {
			http.Error(w, "Please send a request body", 400)
			return
		}

		var data map[string]string
		err := json.NewDecoder(r.Body).Decode(&data); if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		//verify access token
		if !verifyAccessToken(data["access_token"]) {
			http.Error(w, "Access token is unauthorized", 401)
			return
		}

		id, err := strconv.ParseInt(data["id"], 10, 64); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var admin bool
		err = stmt.QueryRow(id).Scan(&admin); if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		if !admin {
			http.Error(w, "User is not an admin. Unauthorized action.", 401)
			return
		}

		_, err = stmt2.Exec(data["username"], true); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func adminRevokeAdminHandler() func(http.ResponseWriter, *http.Request) {
	stmt := prepareQuery("sql/check_admin.sql")
	stmt2 := prepareQuery("sql/get_user_name.sql")
	stmt3 := prepareQuery("sql/update_admin.sql")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This route only accepts POST requests", 400)
			return
		}
		
		if r.Body == nil  {
			http.Error(w, "Please send a request body", 400)
			return
		}

		var data map[string]string
		err := json.NewDecoder(r.Body).Decode(&data); if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		//verify access token
		if !verifyAccessToken(data["access_token"]) {
			http.Error(w, "Access token is unauthorized", 401)
			return
		}

		id, err := strconv.ParseInt(data["id"], 10, 64); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var admin bool
		err = stmt.QueryRow(id).Scan(&admin); if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		if !admin {
			http.Error(w, "User is not an admin. Unauthorized action.", 401)
			return
		}

		var name string
		err = stmt2.QueryRow(id).Scan(&name); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if data["username"] == name {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = stmt3.Exec(data["username"], false); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func adminDeleteUserHandler() func(http.ResponseWriter, *http.Request) {
	stmt := prepareQuery("sql/check_admin.sql")
	stmt2 := prepareQuery("sql/get_user_name.sql")
	stmt3 := prepareQuery("sql/delete_user.sql")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This route only accepts POST requests", 400)
			return
		}
		
		if r.Body == nil  {
			http.Error(w, "Please send a request body", 400)
			return
		}

		var data map[string]string
		err := json.NewDecoder(r.Body).Decode(&data); if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		//verify access token
		if !verifyAccessToken(data["access_token"]) {
			http.Error(w, "Access token is unauthorized", 401)
			return
		}

		id, err := strconv.ParseInt(data["id"], 10, 64); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var admin bool
		err = stmt.QueryRow(id).Scan(&admin); if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		if !admin {
			http.Error(w, "User is not an admin. Unauthorized action.", 401)
			return
		}

		var name string
		err = stmt2.QueryRow(id).Scan(&name); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if data["username"] == name {
			http.Error(w, "Cannot delete yourself", 500)
			return
		}

		_, err = stmt3.Exec(data["username"]); if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/welcome", welcomePageHandler())
	http.HandleFunc("/login/credentials", loginCredentialsHandler())
	http.HandleFunc("/register/credentials", registerCredentialsHandler())
	http.HandleFunc("/verify/token", verifyTokenHandler)
	http.HandleFunc("/update/username", updateUsernameHandler())
	http.HandleFunc("/update/password", updatePasswordHandler())
	http.HandleFunc("/admin/password", adminNewPasswordHandler())
	http.HandleFunc("/admin/new", adminMakeAdminHandler())
	http.HandleFunc("/admin/revoke", adminRevokeAdminHandler())
	http.HandleFunc("/admin/delete/user", adminDeleteUserHandler())
	
	fmt.Println("Running Portal server at port 3333")
	log.Fatal(http.ListenAndServe(":3333", nil))
}
