package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/snykk/kanban-app/client"
	"github.com/snykk/kanban-app/handler/api"
	"github.com/snykk/kanban-app/handler/web"
	"github.com/snykk/kanban-app/middleware"
	"github.com/snykk/kanban-app/repository"
	"github.com/snykk/kanban-app/service"
	"github.com/snykk/kanban-app/utils"

	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

type APIHandler struct {
	UserAPIHandler     api.UserAPI
	TaskAPIHandler     api.TaskAPI
	CategoryAPIHandler api.CategoryAPI
}

type ClientHandler struct {
	AuthWeb      web.AuthWeb
	DashboardWeb web.DashboardWeb
	ModifyWeb    web.ModifyWeb
	HomeWeb      web.HomeWeb
}

//go:embed views/*
var Resources embed.FS

func main() {
	os.Setenv("DATABASE_URL", "postgres://postgres:12345678@localhost:5432/kanban_app")

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		mux := http.NewServeMux()

		err := utils.ConnectDB()
		if err != nil {
			panic(err)
		}

		db := utils.GetDBConnection()

		mux = RunServer(db, mux)
		mux = RunClient(mux, Resources)

		fmt.Println("Server is running on port 8080")
		err = http.ListenAndServe(":8080", mux)
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}

func RunServer(db *gorm.DB, mux *http.ServeMux) *http.ServeMux {
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	userService := service.NewUserService(userRepo, categoryRepo)
	taskService := service.NewTaskService(taskRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo, taskRepo)

	userAPIHandler := api.NewUserAPI(userService)
	taskAPIHandler := api.NewTaskAPI(taskService)
	categoryAPIHandler := api.NewCategoryAPI(categoryService)

	apiHandler := APIHandler{
		UserAPIHandler:     userAPIHandler,
		TaskAPIHandler:     taskAPIHandler,
		CategoryAPIHandler: categoryAPIHandler,
	}

	MuxRoute(mux, "POST", "/api/v1/users/login", middleware.Post(http.HandlerFunc(apiHandler.UserAPIHandler.Login)))
	MuxRoute(mux, "POST", "/api/v1/users/register", middleware.Post(http.HandlerFunc(apiHandler.UserAPIHandler.Register)))
	MuxRoute(mux, "POST", "/api/v1/users/logout", middleware.Post(http.HandlerFunc(apiHandler.UserAPIHandler.Logout)))
	MuxRoute(mux, "GET", "/api/v1/users/get", middleware.Get(middleware.Auth(http.HandlerFunc(apiHandler.UserAPIHandler.GetUserById))), "?user_id=")
	MuxRoute(mux, "DELETE", "/api/v1/users/delete", middleware.Delete(http.HandlerFunc(apiHandler.UserAPIHandler.Delete)), "?user_id=")

	MuxRoute(mux, "GET", "/api/v1/tasks/get", middleware.Get(middleware.Auth(http.HandlerFunc(apiHandler.TaskAPIHandler.GetTask))), "?task_id=")
	MuxRoute(mux, "POST", "/api/v1/tasks/create", middleware.Post(middleware.Auth(http.HandlerFunc(apiHandler.TaskAPIHandler.CreateNewTask))))
	MuxRoute(mux, "PUT", "/api/v1/tasks/update", middleware.Put(middleware.Auth(http.HandlerFunc(apiHandler.TaskAPIHandler.UpdateTask))), "?task_id=")
	MuxRoute(mux, "PUT", "/api/v1/tasks/update/category", middleware.Put(middleware.Auth(http.HandlerFunc(apiHandler.TaskAPIHandler.UpdateTaskCategory))), "?task_id=")
	MuxRoute(mux, "DELETE", "/api/v1/tasks/delete", middleware.Delete(middleware.Auth(http.HandlerFunc(apiHandler.TaskAPIHandler.DeleteTask))), "?task_id=")

	MuxRoute(mux, "GET", "/api/v1/categories/get", middleware.Get(middleware.Auth(http.HandlerFunc(apiHandler.CategoryAPIHandler.GetCategory))))
	MuxRoute(mux, "GET", "/api/v1/categories/dashboard", middleware.Get(middleware.Auth(http.HandlerFunc(apiHandler.CategoryAPIHandler.GetCategoryWithTasks))))
	MuxRoute(mux, "POST", "/api/v1/categories/create", middleware.Post(middleware.Auth(http.HandlerFunc(apiHandler.CategoryAPIHandler.CreateNewCategory))))
	MuxRoute(mux, "DELETE", "/api/v1/categories/delete", middleware.Delete(middleware.Auth(http.HandlerFunc(apiHandler.CategoryAPIHandler.DeleteCategory))), "?category_id=")

	return mux
}

func RunClient(mux *http.ServeMux, embed embed.FS) *http.ServeMux {
	userClient := client.NewUserClient()
	categoryClient := client.NewCategoryClient()
	taskClient := client.NewTaskClient()

	authWeb := web.NewAuthWeb(userClient, embed)
	dashboardWeb := web.NewDashboardWeb(categoryClient, userClient, embed)
	modifyWeb := web.NewModifyWeb(taskClient, categoryClient, embed)
	homeWeb := web.NewHomeWeb(embed)

	client := ClientHandler{
		authWeb, dashboardWeb, modifyWeb, homeWeb,
	}

	mux.HandleFunc("/login", client.AuthWeb.Login)
	mux.HandleFunc("/login/process", client.AuthWeb.LoginProcess)

	mux.HandleFunc("/register", client.AuthWeb.Register)
	mux.HandleFunc("/register/process", client.AuthWeb.RegisterProcess)

	mux.HandleFunc("/logout", client.AuthWeb.Logout)

	mux.Handle("/dashboard", middleware.Auth(http.HandlerFunc(client.DashboardWeb.Dashboard)))

	mux.Handle("/category/add", middleware.Auth(http.HandlerFunc(client.ModifyWeb.AddCategory)))
	mux.Handle("/category/create", middleware.Auth(http.HandlerFunc(client.ModifyWeb.AddCategoryProcess)))

	mux.Handle("/task/add", middleware.Auth(http.HandlerFunc(client.ModifyWeb.AddTask)))
	mux.Handle("/task/create", middleware.Auth(http.HandlerFunc(client.ModifyWeb.AddTaskProcess)))

	mux.Handle("/task/update", middleware.Auth(http.HandlerFunc(client.ModifyWeb.UpdateTask)))
	mux.Handle("/task/update/process", middleware.Auth(http.HandlerFunc(client.ModifyWeb.UpdateTaskProcess)))

	mux.Handle("/task/delete", middleware.Auth(http.HandlerFunc(client.ModifyWeb.DeleteTask)))
	mux.Handle("/category/delete", middleware.Auth(http.HandlerFunc(client.ModifyWeb.DeleteCategory)))

	mux.HandleFunc("/", client.HomeWeb.Index)

	return mux
}

func MuxRoute(mux *http.ServeMux, method string, path string, handler http.Handler, opt ...string) {
	if len(opt) > 0 {
		fmt.Printf("[%s]: %s %v \n", method, path, opt)
	} else {
		fmt.Printf("[%s]: %s \n", method, path)
	}

	mux.Handle(path, handler)
}
