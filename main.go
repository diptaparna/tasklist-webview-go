package main

// #cgo pkg-config: gtk+-3.0
// #include<gtk/gtk.h>
import "C"
import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/webview/webview"
	"log"
	"os"
)

var w webview.WebView
var db *sql.DB
var err error

type Task struct {
	Id          int
	Title       string
	Description string
	Done        bool
}

func AddTask(task Task) {
	taskStr, err := json.Marshal(task)
	if err != nil {
		log.Println("Cannot convert Task object to JSON.")
		return
	}
	w.Eval(fmt.Sprintf(`addTask(%s)`, string(taskStr)))
}

func onLoad() {
	rows, err := db.Query("SELECT id, title, description, done FROM task ORDER BY created DESC")
	if err != nil {
		createDB()
		rows, err = db.Query("SELECT id, title, description, done FROM task ORDER BY created DESC")
	}
	defer rows.Close()
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.Id, &task.Title, &task.Description, &task.Done)
		if err != nil {
			log.Fatal(err)
		}
		AddTask(task)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func updateTask(task Task) {
	_, err := db.Exec("UPDATE task SET title=?, description=?, done=? WHERE id=?", task.Title, task.Description, task.Done, task.Id)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteTask(task Task) {
	_, err := db.Exec("DELETE FROM task WHERE id=?", task.Id)
	if err != nil {
		log.Fatal(err)
	}
}

func createTask(task Task) int64 {
	res, err := db.Exec("INSERT INTO task (title, description) VALUES (?,?)", task.Title, task.Description)
	if err != nil {
		log.Fatal(err)
	}
	taskId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	return taskId
}

func createDB() {
	_, err = db.Exec(`
					CREATE TABLE task (id INTEGER PRIMARY KEY AUTOINCREMENT,  
					title STRING DEFAULT 'New Task', 
					description STRING DEFAULT 'No description.', 
					created TIMESTAMP DEFAULT current_timestamp, 
					done INTEGER DEFAULT 0)
					`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO task (title,description) VALUES (?,?)", "Demo Task #1", "Please add a description.")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO task (title,description) VALUES (?,?)", "Demo Task #2", "Click on the title or description to edit them.")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	db, err = sql.Open("sqlite3", "assets/taskdb.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	debug := true
	w = webview.New(debug)
	defer w.Destroy()

	w.SetTitle("Minimal webview example")
	w.SetSize(600, 600, webview.HintFixed)
	C.gtk_window_set_keep_above((*C.GtkWindow)(w.Window()), 1)

	PWD, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	w.Navigate(fmt.Sprintf("file://%s/assets/index.html", PWD))
	w.Bind("onLoad", onLoad)
	w.Bind("updateTask", updateTask)
	w.Bind("createTask", createTask)
	w.Bind("deleteTask", deleteTask)

	w.Run()
}
