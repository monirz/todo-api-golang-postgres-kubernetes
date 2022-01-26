package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

const (
	DB_HOST   = "db"
	DB_PORT   = "5432"
	DB_USER   = "postgres"
	DB_PASS   = "password"
	DB_NAME   = "tododb"
	Migration = `
	  create table if not exists todos(
		  id serial primary key,
		  title text not null,
		  task text not null,
		  completed boolean not null default false,
		  created_at timestamp with time zone DEFAULT current_timestamp,
		  updated_at timestamp with time zone DEFAULT current_timestamp
		)
	`
)

type ToDo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Task      string    `json:"Task"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var db *sql.DB

//getToDos returns list of todos
func getToDos() ([]ToDo, error) {
	rows, err := db.Query("SELECT * FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []ToDo{}
	for rows.Next() {
		var id int64
		var title string
		var task string
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&id, &title, &task, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, ToDo{ID: id, Title: title, Task: task, CreatedAt: createdAt, UpdatedAt: updatedAt})
	}

	return todos, nil
}

func addTodo(title, task string) error {
	_, err := db.Exec("INSERT INTO todos(title, task) VALUES(:1, :2)", title, task)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	var err error
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome!"))
	})

	r.Post("/todos", createToDoHandler)
	r.Get("/todos", getToDosHandler)
	r.Get("/todos/{id}", getToDoHandler)
	r.Patch("/todos", updateToDoHandler)

	dbInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", DB_HOST, DB_USER, DB_PASS, DB_NAME)

	db, err = sql.Open("postgres", dbInfo)

	if err != nil {
		panic(err)
	}
	_, err = db.Query(Migration)

	if err != nil {
		log.Println("error creating migration", err)
		return
	}

	log.Println("Server started on port :8090")
	log.Fatal(http.ListenAndServe(":8090", r))

}

//createToDoHandler creates new todo
func createToDoHandler(w http.ResponseWriter, r *http.Request) {
	author := r.FormValue("author")
	content := r.FormValue("content")
	if author == "" || content == "" {
		http.Error(w, "author and content are required", http.StatusBadRequest)
		return
	}

	if err := addTodo(author, content); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success!"))
}

//getToDosHandler returns list of todos
func getToDosHandler(w http.ResponseWriter, r *http.Request) {

	todos, err := getToDos()

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

//getToDoHandler returns todo by id
func getToDoHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	var todo ToDo
	err := db.QueryRow("SELECT * FROM todos WHERE id = :1", id).Scan(&todo.ID, &todo.Title, &todo.Task, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

//updateToDoHandler updates todo
func updateToDoHandler(w http.ResponseWriter, r *http.Request) {

	var todo ToDo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE todos SET title = :1, task = :2, completed = :3, updated_at = :4 WHERE id = :5", todo.Title, todo.Task, todo.Completed, todo.UpdatedAt, todo.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success!"))
}
