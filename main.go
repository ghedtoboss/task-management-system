package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"task-management-system/db"
	"task-management-system/handlers"
	"task-management-system/middleware"

	_ "task-management-system/docs"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Task Management API
// @version 1.0
// @description API for managing tasks and users
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
func main() {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY env is not set")
	}

	db := db.InitDB("root:Eses147852@tcp(127.0.0.1:3306)/task_management?parseTime=true")
	defer db.Close()
	fmt.Println("Veritabanına bağlanıldı.")

	r := mux.NewRouter()

	appHandler := &handlers.AppHandler{DB: db}

	//routes
	r.Handle("/register", appHandler.Register()).Methods("POST")
	r.Handle("/login", appHandler.Login()).Methods("POST")
	r.Handle("/tasks", middleware.JWTMiddleware(middleware.RoleMiddleware("admin")(appHandler.CreateTask()))).Methods("POST")
	r.Handle("/tasks/{task_id}", middleware.JWTMiddleware(middleware.RoleMiddleware("admin")(appHandler.UpdateTask()))).Methods("PUT")
	r.Handle("/tasks/{task_id}", middleware.JWTMiddleware(middleware.RoleMiddleware("admin")(appHandler.DeleteTask()))).Methods("DELETE")
	r.Handle("/tasks", middleware.JWTMiddleware(middleware.RoleMiddleware("admin", "user")(appHandler.GetTasks()))).Methods("GET")
	r.Handle("/user/stats", middleware.JWTMiddleware(appHandler.GetStats())).Methods("GET")
	r.Handle("/friends", middleware.JWTMiddleware(appHandler.CreateFriendship())).Methods("POST")
	r.Handle("/friends/accept", middleware.JWTMiddleware(appHandler.AcceptFriendRequest())).Methods("POST")
	r.Handle("/friends/reject", middleware.JWTMiddleware(appHandler.RejectFriendRequest())).Methods("POST")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
