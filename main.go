package main

import (
	"TinyWebServerGo/sessionmanager"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var globalSessions *sessionmanager.SessionManager

func init() {
	provider := sessionmanager.NewMemoryProvider()
	globalSessions = sessionmanager.NewSessionManager("sessionId", 3600, 3600, 60, provider)
}

func mainHandle(writer http.ResponseWriter, req *http.Request) {
	fmt.Print(req.URL.RequestURI())
	if req.URL.RequestURI() == "/" {
		showIndex(writer, req)
	} else {
		pageNotFound(writer, req)
	}

}

func showIndex(writer http.ResponseWriter, req *http.Request) {
	file, err := os.ReadFile("root/index.html")
	checkErr(err)

	fmt.Fprint(writer, string(file))

}

func checkLogin(username string, password string) bool {
	db, err := sql.Open("mysql", "root:@/tiny_web_server_go?charset=utf8")
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("SELECT * FROM user_tb WHERE username = '" + username + "' AND password = '" + password + "'") //sql注入测试代码
	checkErr(err)
	defer rows.Close()
	return rows.Next()
}

func login(writer http.ResponseWriter, req *http.Request) {
	if _, err := globalSessions.Get(req, "username"); err == nil {
		writer.Header().Set("Location", "/welcome")
		writer.WriteHeader(302)
		return
	}
	if req.Method == "GET" {
		t1, err := template.ParseFiles("root/login.html")
		checkErr(err)
		t1.Execute(writer, "")
	} else {
		req.ParseForm()
		username := req.Form.Get("username")
		password := req.Form.Get("password")
		fmt.Printf("username: %v, password: %v\n", username, password)
		if checkLogin(username, password) {
			data := make(map[string]string)
			data["username"] = username
			globalSessions.Create(&writer, req, data)
			writer.Header().Set("Location", "/welcome")
			writer.WriteHeader(302)
		} else {
			t1, err := template.ParseFiles("root/login.html")
			checkErr(err)
			t1.Execute(writer, `用户名或密码错误`)
		}
	}
}

func logout(writer http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		pageNotFound(writer, req)
	} else {
		globalSessions.Destroy(&writer, req)
		writer.Header().Set("Location", "/index")
		writer.WriteHeader(302)
	}
}

func doRegister(username string, password string) bool {
	db, err := sql.Open("mysql", "root:@/tiny_web_server_go?charset=utf8")
	if err != nil {
		panic(err)
	} else {
		defer db.Close()
	}
	checkErr(err)
	rows, err := db.Query("SELECT * FROM user_tb WHERE username = '" + username + "' AND password = '" + password + "'") //sql注入测试代码
	checkErr(err)
	if rows.Next() {
		rows.Close()
		return false
	}
	rows.Close()
	stmt, err := db.Prepare("INSERT INTO user_tb VALUES (?,?)")
	checkErr(err)

	_, err = stmt.Exec(username, password)
	if err != nil {
		panic(err)
	}
	return true
}

func register(writer http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		t1, err := template.ParseFiles("root/registereq.html")
		checkErr(err)
		t1.Execute(writer, ``)
	} else {
		req.ParseForm()
		username := req.Form.Get("username")
		password := req.Form.Get("password")
		fmt.Printf("username: %v, password: %v\n", username, password)
		if doRegister(username, password) {
			writer.Header().Set("Location", "/login")
			writer.WriteHeader(302)
		} else {
			t1, err := template.ParseFiles("root/registereq.html")
			checkErr(err)
			t1.Execute(writer, `该用户名已被注册`)
		}
	}

}

func pageNotFound(writer http.ResponseWriter, req *http.Request) {
	text, err := os.ReadFile("root/404.html")
	checkErr(err)
	writer.WriteHeader(404)
	fmt.Fprint(writer, string(text))
}

func welcome(writer http.ResponseWriter, req *http.Request) {
	if username, err := globalSessions.Get(req, "username"); err == nil {
		t1, err := template.ParseFiles("root/welcome.html")
		checkErr(err)
		t1.Execute(writer, username)
	} else {
		writer.Header().Set("Location", "/login")
		writer.WriteHeader(302)
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Print("server start at port 8888")
	http.HandleFunc("/", mainHandle)
	http.HandleFunc("/login", login)
	http.HandleFunc("/register", register)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/welcome", welcome)
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
