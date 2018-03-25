package todo

var allowedStatuses = []string{"New", "In Progress", "Closed"}

// Todos is a list of todo.Todo structs
type Todos struct {
	TodoList []Todo `json:"todos"`
}

// Todo is a struct containing the ID of a todo, as well as, title and status
type Todo struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// CreateTodo is the expected payload for a create todo request
type CreateTodo struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}
