package todo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Create will allow a user to create a new todo
// The supported body is {"title": "", "status": ""}
// POST
func Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	db, success := getDb(w)
	if !success {
		return
	}

	todo, success := receiveTodo(w, r)
	if ! success {
		return
	}

	insertStmt := fmt.Sprintf(`INSERT INTO todo (title, status) VALUES ('%s', '%s') RETURNING id`, todo.Title, todo.Status)

	var todoID int

	// Insert and get back newly created todo ID
	if err := db.QueryRow(insertStmt).Scan(&todoID); err != nil {
		fmt.Printf("Failed to save to db: %s", err.Error())
	}

	fmt.Printf("Todo Created -- ID: %d\n", todoID)

	newTodo := Todo{}
	db.QueryRow("SELECT id, title, status FROM todo WHERE id=$1", todoID).Scan(&newTodo.ID, &newTodo.Title, &newTodo.Status)

	jsonResp, _ := json.Marshal(newTodo)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, string(jsonResp))
}

// Read will provide a list of all current to-dos
// GET
func Read(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	db, success := getDb(w)
	if !success {
		return
	}

	todoList := []Todo{}

	rows, err := db.Query("SELECT id, title, status FROM todo")
	if err != nil {
		failStatus(w, "Failed to get rows from db")
		return
	}
	defer rows.Close()

	for rows.Next() {
		todo := Todo{}
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Status); err != nil {
			failStatus(w, "Failed to get rows from todo list")
			return
		}

		todoList = append(todoList, todo)
	}

	jsonResp, _ := json.Marshal(Todos{TodoList: todoList})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, string(jsonResp))
}

// Update an item by id,
// supported body format:{"title": "", "status": ""}
// PUT
func Update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Setup
	db, success := getDb(w)
	if !success {
		return
	}

	id := p.ByName("id")
	if !idInTable(id, w, db) {
		return
	}

	todo, success := receiveTodo(w, r)

	if !success {
		return
	}

	// Execution
	updateStmt := `UPDATE todo SET title = $2, status = $3 WHERE id = $1;`
	res, err := db.Exec(updateStmt, id, todo.Title, todo.Status)

	if opSuccess(w, "Update", res, err) {
		iid, _ := strconv.Atoi(id)
		jsonResp, _ := json.Marshal(Todo{ID: iid, Title: todo.Title, Status: todo.Status})
		w.WriteHeader(200)
		fmt.Fprintf(w, string(jsonResp))
	}
}

// Delete will delete an item by id
// DELETE
func Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Setup
	db, success := getDb(w)
	if !success {
		return
	}

	id := p.ByName("id")
	if !idInTable(id, w, db) {
		return
	}

	// Extract todo we're supposed to delete
	retTodo := &Todo{}
	queryStatement := `SELECT id, title, status FROM todo WHERE id=$1`
	db.QueryRow(queryStatement, id).Scan(&retTodo.ID, &retTodo.Title, &retTodo.Status)

	// Execution
	deleteStmt := `DELETE FROM todo WHERE id = $1;`
	res, err := db.Exec(deleteStmt, id)

	if opSuccess(w, "Delete", res, err) {
		jsonResp, _ := json.Marshal(retTodo)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, string(jsonResp))
	}
}

func receiveTodo(w http.ResponseWriter, r *http.Request) (*CreateTodo, bool) {
	var todo *CreateTodo
	json.NewDecoder(r.Body).Decode(&todo)

	if todo.Status == "" || todo.Title == "" {
		http.Error(w, "Todo request is missing status or title", http.StatusBadRequest)
		return todo, false
	}

	validStatus := false
	for _, status := range allowedStatuses {
		if todo.Status == status {
			validStatus = true
			break
		}
	}

	if !validStatus {
		http.Error(w, "Invalid todo status", http.StatusBadRequest)
	}

	return todo, validStatus
}

func getDb(w http.ResponseWriter) (*sql.DB, bool) {
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		failStatus(w, "Error connecting to database", err.Error())
		return nil, false
	}
	return db, true
}

func failStatus(w http.ResponseWriter, m string, msgs ...string) {
	w.WriteHeader(500)
	if len(msgs) != 0 {
		m += ": " + strings.Join(msgs, ", ")
	}
	fmt.Fprintf(w, m)
	fmt.Println(m)
	return
}

func idInTable(id string, w http.ResponseWriter, db *sql.DB) bool {
	rows, err := db.Query("SELECT count(1) FROM todo WHERE id = $1;", id)
	defer rows.Close()

	if err != nil {
		failStatus(w, "Encountered error trying to find item", err.Error())
		return false
	}

	var count int
	rows.Next()
	if err = rows.Scan(&count); err != nil {
		failStatus(w, "Encountered error trying to find item", err.Error())
		return false
	}

	if count == 0 {
		failStatus(w, "todo item not found")
		return false
	}

	return true
}

func opSuccess(w http.ResponseWriter, op string, res sql.Result, err error) bool {
	if err != nil {
		failStatus(w, op + ": Error Encounted", err.Error())
		return false
	}

	count, err := res.RowsAffected()
	if err != nil {
		failStatus(w, op + ": Error Encountered", err.Error())
		return false
	}
	if count == 0 {
		failStatus(w, ": Operation Failed - no rows affected")
		return false
	}
	return true
}
