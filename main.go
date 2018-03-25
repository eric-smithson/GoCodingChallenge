package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"

	"github.com/rackerlabs/GoCodingChallenge/todo"
)

// Status :=
func Status(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Status Request Received")
	w.WriteHeader(200)
	fmt.Fprint(w, "OK\n")
}

func main() {
	router := httprouter.New()
	router.GET("/", Status)
	router.POST("/todos", todo.Create)
	router.GET("/todos", todo.List)

	log.Println("Starting server...")

	// Make sure you have DB_USER, DB_PASSWORD and DB_NAME environment variables set.
	// We use them elsewhere
	log.Fatal(http.ListenAndServe(":8080", router))
}
