package main


import (
	"log"
	"net/http"
	"os"

	"github.com/enzdor/goblog/controllers"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)


func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	errEnv := godotenv.Load()
	if errEnv != nil {
	    log.Fatal(errEnv)
	}
	user := os.Getenv("DBUSER")
	pass := os.Getenv("DBPASS")
	name := os.Getenv("DBNAME")
	key := os.Getenv("JWTKEY")

	db := controllers.NewDB(user, pass, name)
	h := controllers.NewHandler(db, key)

	http.HandleFunc("/", h.ServeIndex)
	http.HandleFunc("/login", h.ServeLogin)
	http.HandleFunc("/logout", h.ServeLogout)
	http.HandleFunc("/manage", h.ServeManage)
	http.HandleFunc("/article/", h.ServeArticle)
	http.HandleFunc("/post", h.ServePost)
	http.HandleFunc("/edit/", h.ServeEdit)
	http.HandleFunc("/delete/", h.ServeDelete)
	http.HandleFunc("/error/", h.ServeError)
	
	log.Print("Listening on port :3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
	    log.Fatal(err)
	}
}









