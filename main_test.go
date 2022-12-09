package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	main "github.com/snykk/kanban-app"
	"github.com/snykk/kanban-app/entity"
	"github.com/snykk/kanban-app/utils"

	_ "github.com/jackc/pgx/v4/stdlib"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var _ = os.Setenv("DATABASE_URL", "postgres://postgres:12345678@localhost:5432/kanban_app")

func SetCookie(mux *http.ServeMux) *http.Cookie {
	login := entity.UserLogin{
		Email:    "test@mail.com",
		Password: "testing123",
	}

	body, _ := json.Marshal(login)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/users/login", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, r)

	var cookie *http.Cookie

	for _, c := range w.Result().Cookies() {
		if c.Name == "user_id" {
			cookie = c
		}
	}

	return cookie
}

var _ = Describe("TestAPIHandler", Ordered, func() {
	var apiServer *http.ServeMux
	var db *gorm.DB
	var userTest int
	var categoryIdTest int
	var categoryIdForTaskTest int
	var taskIdTest int

	BeforeAll(func() {
		conn, err := gorm.Open(postgres.New(postgres.Config{
			DriverName: "pgx",
			DSN:        os.Getenv("DATABASE_URL"),
		}), &gorm.Config{})
		if err != nil {
			panic(err)
		}

		db = conn

		db.Exec("DROP TABLE IF EXISTS tasks CASCADE")
		db.Exec("DROP TABLE IF EXISTS categories CASCADE")
		db.Exec("DROP TABLE IF EXISTS users CASCADE")

		db.AutoMigrate(entity.User{}, entity.Category{}, entity.Task{})

		apiServer = http.NewServeMux()
		apiServer = main.RunServer(db, apiServer)
	})

	AfterAll(func() {
		ctx := context.Background()

		err := db.WithContext(ctx).Exec("DELETE FROM tasks WHERE user_id = ?", userTest).Error
		if err != nil {
			panic(err)
		}

		err = db.WithContext(ctx).Exec("DELETE FROM categories WHERE user_id = ?", userTest).Error
		if err != nil {
			panic(err)
		}

		err = db.WithContext(ctx).Exec("DELETE FROM users WHERE id = ?", userTest).Error
		if err != nil {
			panic(err)
		}
	})

	Describe("/users/register", func() {
		When("send empty register request data", func() {
			It("should return a bad request", func() {
				reqRegister := entity.UserRegister{}
				reqBody, _ := json.Marshal(reqRegister)

				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewReader(reqBody))
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				errResp := entity.ErrorResponse{}
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
				Expect(errResp.Error).To(Equal("register data is empty"))
			})
		})

		When("send register request data with method POST", func() {
			It("should return a success", func() {
				reqRegister := entity.UserRegister{
					Fullname: "test",
					Email:    "test@mail.com",
					Password: "testing123",
				}

				reqBody, _ := json.Marshal(reqRegister)

				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewReader(reqBody))
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				var resp = map[string]interface{}{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusCreated))
				Expect(resp["message"]).To(Equal("register success"))

				userTest = int(resp["user_id"].(float64))
			})
		})

		When("send register twice with same data by POST method", func() {
			It("should return a bad request", func() {
				reqRegister := entity.UserRegister{
					Fullname: "test",
					Email:    "test@mail.com",
					Password: "testing123",
				}

				reqBody, _ := json.Marshal(reqRegister)

				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewReader(reqBody))
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				errResp := entity.ErrorResponse{}
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusInternalServerError))
				Expect(errResp.Error).NotTo(BeEmpty())
			})
		})
	})

	Describe("/users/login", func() {
		When("send empty email and password with POST method", func() {
			It("should return a bad request", func() {
				loginData := entity.UserLogin{
					Email:    "",
					Password: "",
				}

				body, _ := json.Marshal(loginData)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/users/login", bytes.NewReader(body))
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				errResp := entity.ErrorResponse{}
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
				Expect(errResp.Error).To(Equal("email or password is empty"))
			})
		})

		When("send email and password with POST method", func() {
			It("should return a success", func() {
				loginData := entity.UserLogin{
					Email:    "test@mail.com",
					Password: "testing123",
				}

				body, _ := json.Marshal(loginData)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/users/login", bytes.NewReader(body))
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				var resp = map[string]interface{}{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				// Expect(w.Result().Request.Cookie("user_id")).NotTo(BeNil())
				Expect(resp["message"]).To(Equal("login success"))
			})
		})
	})

	// ==============================================
	// ==============     CATEGORY     ==============
	// ==============================================
	Describe("/categories/dashboard", func() {
		When("hit endpoint without user login", func() {
			It("should return an error unauthorized", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/api/v1/categories/dashboard", nil)
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				errResp := entity.ErrorResponse{}
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(errResp.Error).To(Equal("error unauthorized user id"))
			})
		})

		When("hit endpoint with GET method", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/api/v1/categories/dashboard", nil)
				r.Header.Set("Content-Type", "application/json")

				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = []entity.CategoryData{}

				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil()) // <- error

				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(len(resp)).To(Equal(4))

				categoryIdForTaskTest = resp[0].ID
			})
		})
	})

	Describe("/categories/get", func() {
		When("hit endpoint without user login", func() {
			It("should return an error unauthorized", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/api/v1/categories/get", nil)
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				errResp := entity.ErrorResponse{}
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(errResp.Error).To(Equal("error unauthorized user id"))
			})
		})

		When("hit endpoint with GET method", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/api/v1/categories/get", nil)
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = []entity.Category{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(len(resp)).To(Equal(4))
				Expect(resp[0].Type).To(Equal("Todo"))
				Expect(resp[1].Type).To(Equal("In Progress"))
				Expect(resp[2].Type).To(Equal("Done"))
				Expect(resp[3].Type).To(Equal("Backlog"))
			})
		})
	})

	Describe("/categories/create", func() {
		When("hit endpoint without user login", func() {
			It("should return an error unauthorized", func() {
				categoryData := entity.Category{
					Type: "Testing",
				}

				body, _ := json.Marshal(categoryData)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/categories/create", bytes.NewReader(body))
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				var errResp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(errResp.Error).To(Equal("error unauthorized user id"))
			})
		})

		When("hit endpoint with GET method", func() {
			It("should return method not allowed", func() {
				categoryData := entity.Category{
					Type: "Testing",
				}

				body, _ := json.Marshal(categoryData)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/api/v1/categories/create", bytes.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusMethodNotAllowed))
				Expect(resp.Error).To(Equal("method is not allowed!"))
			})
		})

		When("hit endpoint with POST method without required data", func() {
			It("should return a bad request", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/categories/create", bytes.NewReader([]byte("{}")))
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
				Expect(resp.Error).To(Equal("invalid category request"))
			})
		})

		// should success
		When("hit endpoint with POST method", func() {
			It("should return a success", func() {
				categoryData := entity.Category{
					Type: "Testing",
				}

				body, _ := json.Marshal(categoryData)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/categories/create", bytes.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = map[string]interface{}{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusCreated))
				Expect(resp["message"]).To(Equal("success create new category"))

				categoryIdTest = int(resp["category_id"].(float64))
			})
		})
	})

	Describe("/categories/delete", func() {
		When("hit endpoint without user login", func() {
			It("should return an error unauthorized", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/categories/delete?category_id=%v", categoryIdTest), nil)
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				var errResp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(errResp.Error).To(Equal("error unauthorized user id"))
			})
		})

		When("hit endpoint with GET method", func() {
			It("should return method not allowed", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/categories/delete?category_id=%v", categoryIdTest), nil)
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusMethodNotAllowed))
				Expect(resp.Error).To(Equal("method is not allowed!"))
			})
		})

		When("hit endpoint with DELETE method", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/categories/delete?category_id=%v", categoryIdTest), nil)
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = map[string]interface{}{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(resp["message"]).To(Equal("success delete category"))
			})
		})
	})

	// ==============================================
	// ============== 		TASK 	   ==============
	// ==============================================
	Describe("/tasks/create", func() {
		// create one in one if the category, get the id
		When("hit endpoint without user login", func() {
			It("should return an error unauthorized", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/tasks/create", bytes.NewReader([]byte("{}")))
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				var errResp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(errResp.Error).To(Equal("error unauthorized user id"))
			})
		})

		When("hit endpoint with GET method", func() {
			It("should return method not allowed", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/api/v1/tasks/create", bytes.NewReader([]byte("{}")))
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusMethodNotAllowed))
				Expect(resp.Error).To(Equal("method is not allowed!"))
			})
		})

		When("hit endpoint with POST method without required data", func() {
			It("should return error bad request", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/tasks/create", bytes.NewReader([]byte("{}")))
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
				Expect(resp.Error).To(Equal("invalid task request"))
			})
		})

		When("hit endpoint with POST method", func() {
			It("should return a success", func() {
				taskData := entity.Task{
					CategoryID:  categoryIdForTaskTest,
					Title:       "Testing",
					Description: "Testing",
				}

				body, _ := json.Marshal(taskData)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/api/v1/tasks/create", bytes.NewReader(body))

				r.AddCookie(SetCookie(apiServer))
				apiServer.ServeHTTP(w, r)

				var resp = map[string]interface{}{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusCreated))
				Expect(resp["message"]).To(Equal("success create new task"))

				taskIdTest = int(resp["task_id"].(float64))
			})
		})
	})

	Describe("/tasks/get", func() {
		When("hit endpoint without user login", func() {
			It("should return an error unauthorized", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/api/v1/tasks/get", nil)
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				var errResp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(errResp.Error).To(Equal("error unauthorized user id"))
			})
		})

		// get of the one task from category, with the id
		When("hit endpoint with GET method", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/tasks/get?category_id=%v", categoryIdForTaskTest), nil)

				r.AddCookie(SetCookie(apiServer))
				apiServer.ServeHTTP(w, r)

				var resp = []entity.Task{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(len(resp)).To(Equal(1))
				Expect(resp[0].Title).To(Equal("Testing"))
				Expect(resp[0].Description).To(Equal("Testing"))
				Expect(resp[0].CategoryID).To(Equal(categoryIdForTaskTest))
			})
		})

		When("hit endpoint with GET method and set query task_id", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/tasks/get?task_id=%v", taskIdTest), nil)
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = entity.Task{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(resp.Title).To(Equal("Testing"))
				Expect(resp.Description).To(Equal("Testing"))
				Expect(resp.CategoryID).To(Equal(categoryIdForTaskTest))
			})
		})
	})

	Describe("/tasks/update", func() {
		When("hit endpoint without user login", func() {
			It("should return an error unauthorized", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("PUT", "/api/v1/tasks/update", nil)
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				var errResp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(errResp.Error).To(Equal("error unauthorized user id"))
			})
		})

		When("hit endpoint with GET method", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/api/v1/tasks/update", bytes.NewReader([]byte("{}")))
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = entity.ErrorResponse{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusMethodNotAllowed))
				Expect(resp.Error).To(Equal("method is not allowed!"))
			})
		})

		When("hit endpoint with PUT method", func() {
			It("should return a success", func() {
				taskData := entity.Task{
					ID:          taskIdTest,
					Title:       "Testing Updated",
					Description: "Testing Updated",
				}

				body, _ := json.Marshal(taskData)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/tasks/update?task_id=%v", taskIdTest), bytes.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = map[string]interface{}{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)

				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(int(resp["task_id"].(float64))).To(Equal(taskIdTest))
				Expect(resp["message"]).To(Equal("success update task"))
			})
		})
	})

	Describe("/tasks/delete", func() {
		When("hit endpoint without user login", func() {
			It("should return an error unauthorized", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("DELETE", "/api/v1/tasks/delete", nil)
				r.Header.Set("Content-Type", "application/json")

				apiServer.ServeHTTP(w, r)

				var errResp = entity.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(&errResp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(errResp.Error).To(Equal("error unauthorized user id"))
			})
		})

		When("hit endpoint with GET method", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/api/v1/tasks/delete", bytes.NewReader([]byte("{}")))
				r.Header.Set("Content-Type", "application/json")
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = entity.ErrorResponse{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusMethodNotAllowed))
				Expect(resp.Error).To(Equal("method is not allowed!"))
			})
		})

		When("hit endpoint with DELETE method", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/tasks/delete?task_id=%v", taskIdTest), nil)
				r.AddCookie(SetCookie(apiServer))

				apiServer.ServeHTTP(w, r)

				var resp = map[string]interface{}{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				Expect(err).To(BeNil())
				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(int(resp["task_id"].(float64))).To(Equal(taskIdTest))
				Expect(resp["message"]).To(Equal("success delete task"))
			})
		})
	})
})

var _ = Describe("TestWebHandler", Ordered, func() {
	var clientHandler *http.ServeMux
	var db *gorm.DB
	var userClientID int

	BeforeAll(func() {
		err := utils.ConnectDB()
		if err != nil {
			panic(err)
		}

		db = utils.GetDBConnection()

		clientHandler = http.NewServeMux()

		clientHandler = main.RunServer(db, clientHandler)
		clientHandler = main.RunClient(clientHandler, main.Resources)

		register := entity.UserRegister{
			Fullname: "testing client",
			Email:    "test@mail.com",
			Password: "testing123",
		}

		byte, _ := json.Marshal(register)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewReader(byte))
		r.Header.Set("Content-Type", "application/json")

		clientHandler.ServeHTTP(w, r)

		var resp = map[string]interface{}{}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			panic(err)
		}

		userClientID = int(resp["user_id"].(float64))
	})

	AfterAll(func() {
		ctx := context.Background()

		err := db.WithContext(ctx).Exec("DELETE FROM users WHERE id = ?", userClientID).Error
		if err != nil {
			panic(err)
		}
	})

	Describe("/", func() {
		When("hit endpoint", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/", nil)

				clientHandler.ServeHTTP(w, r)

				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("Kanban App"))
			})
		})
	})

	Describe("/login", func() {
		When("hit endpoint with GET method", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/login", nil)

				clientHandler.ServeHTTP(w, r)

				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("Login"))
			})
		})
	})

	Describe("/register", func() {
		When("hit endpoint with GET method", func() {
			It("should return a success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/register", nil)

				clientHandler.ServeHTTP(w, r)

				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("Register"))
			})
		})
	})

	Describe("/dashboard", func() {
		When("user is not logged in", func() {
			It("should redirect to login", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/dashboard", nil)

				clientHandler.ServeHTTP(w, r)

				Expect(w.Result().StatusCode).To(Equal(http.StatusSeeOther))
				Expect(w.Body.String()).To(ContainSubstring(`<a href="/login">See Other</a>`))
			})
		})
	})

	Describe("/category/add", func() {
		When("user is not logged in", func() {
			It("should redirect to login", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/category/add", nil)

				clientHandler.ServeHTTP(w, r)

				Expect(w.Result().StatusCode).To(Equal(http.StatusSeeOther))
				Expect(w.Body.String()).To(ContainSubstring(`<a href="/login">See Other</a>`))
			})
		})

		When("user is logged in", func() {
			It("should return success", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/category/add", nil)
				cookie := SetCookie(clientHandler)

				r.AddCookie(cookie)

				clientHandler.ServeHTTP(w, r)

				Expect(w.Result().StatusCode).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("Add Category"))
			})
		})
	})

	Describe("/task/add", func() {
		When("user is not logged in", func() {
			It("should redirect to login", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/task/add", nil)

				clientHandler.ServeHTTP(w, r)

				Expect(w.Result().StatusCode).To(Equal(http.StatusSeeOther))
				Expect(w.Body.String()).To(ContainSubstring(`<a href="/login">See Other</a>`))
			})
		})
	})

	Describe("/task/update", func() {
		When("user is not logged in", func() {
			It("should redirect to login", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/task/update", nil)

				clientHandler.ServeHTTP(w, r)

				Expect(w.Result().StatusCode).To(Equal(http.StatusSeeOther))
				Expect(w.Body.String()).To(ContainSubstring(`<a href="/login">See Other</a>`))
			})
		})
	})
})
