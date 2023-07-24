package main

import (
	"database/sql"
	"log"

	// "strings"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

// Task represents the structure of a task
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
}

var db *sql.DB

// This is the entry point of the program and where the API server starts. It sets up the SQLite database, defines the API endpoints, and runs the server.
func main() {

	// Connect to the SQLite database
	var err error
	db, err = sql.Open("sqlite", "./tasks.db")
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Create the tasks table if it does not exist
	createTable()

	// set up the Gin router
	router := gin.Default()

	// Define API endpoints and their respective handler functions
	router.POST("/tasks", createTask)
	router.GET("/tasks/:id", getTask)
	router.PUT("/tasks/:id", updateTask)
	router.DELETE("/tasks/:id", deleteTask)
	router.GET("/tasks", listTasks)

	// Run the server
	router.Run(":8080")
}

// This function is responsible for creating the tasks table in the SQLite database.
// The table will store task-related information like ID, title, description, due date, and status.
// The CREATE TABLE SQL statement is used to define the table schema with the required columns.
func createTable() {
	// Create the tasks table if it does not exist
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		description TEXT,
		due_date TEXT,
		status TEXT
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("Table creation failed:", err)
	}
}

// This function handles the POST request to create a new task.
// It accepts a JSON payload containing task details like title, description, due date and status.
// It generates a unique ID for the task, stores it in the database, and returns the created task with the assigned ID.
func createTask(context *gin.Context) {
	var task Task
	if err := context.ShouldBindJSON(&task); err != nil {
		// log.Printf("Error parsing JSON: %v", err)
		context.JSON(400, gin.H{"error": "Invalid task data"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		context.JSON(500, gin.H{"error": "Database transaction failed"})
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO tasks (title, description, due_date, status) VALUES (?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		context.JSON(500, gin.H{"error": "Database statement preparation failed"})
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(task.Title, task.Description, task.DueDate, task.Status)
	if err != nil {
		tx.Rollback()
		context.JSON(500, gin.H{"error": "Database insertion failed"})
		return
	}

	taskID, _ := res.LastInsertId()
	task.ID = int(taskID)

	tx.Commit()
	context.JSON(201, task)
}

// This function handles the GET request to retrieve a task by its ID.
// It accepts the task ID as a parameter, queries the database to find the corresponding task, and returns the task details if found.
// If the task is not found, it returns an appropriate error message.
func getTask(context *gin.Context) {
	id := context.Param("id")

	row := db.QueryRow("SELECT * FROM tasks WHERE id = ?", id)
	var task Task
	err := row.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)

	if err != nil {
		context.JSON(404, gin.H{"error": "Task not found"})
		return
	}

	context.JSON(200, task)
}

// This function handles the PUT request to update an existing task.
// It accepts the task ID as a parameter and a JSON payload containing the updated task details (title, description, due date, status).
// It updates the corresponding task in the database and returns the updated task if successful.
// If the task is not found, it returns an appropriate error message.
func updateTask(context *gin.Context) {
	id := context.Param("id")

	var task Task
	if err := context.ShouldBindJSON(&task); err != nil {
		context.JSON(400, gin.H{"error": "Invalid task data"})
		return
	}

	stmt, err := db.Prepare("UPDATE tasks SET title=?, description=?, due_date=?, status=? WHERE id=?")
	if err != nil {
		context.JSON(500, gin.H{"error": "Database statement preparation failed"})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(task.Title, task.Description, task.DueDate, task.Status, id)
	if err != nil {
		context.JSON(500, gin.H{"error": "Database update failed"})
		return
	}

	context.JSON(200, task)
}

// This function handles the DELETE request to delete a task by its ID.
// It accepts the task ID as a parameter, deletes the corresponding task from the database, and returns a success message if the deletion is successful.
// If the task is not found, it returns an appropriate error message.
func deleteTask(context *gin.Context) {
	id := context.Param("id")

	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		context.JSON(500, gin.H{"error": "Database delete failed"})
		return
	}

	context.JSON(200, gin.H{"message": "Task deleted successfully"})
}

//	This function handles the GET request to retrieve all tasks from the database.
//
// It queries the database to get all tasks and returns a list of tasks, including their details (title, description, due date).
func listTasks(context *gin.Context) {
	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		context.JSON(500, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
		if err != nil {
			context.JSON(500, gin.H{"error": "Database scan failed"})
			return
		}
		tasks = append(tasks, task)
	}

	context.JSON(200, tasks)
}
