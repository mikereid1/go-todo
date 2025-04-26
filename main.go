package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TodoDTO struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type Todo struct {
	Id          bson.ObjectID `bson:"_id"`
	Title       string        `bson:"title"`
	Description string        `bson:"description"`
	Done        bool          `bson:"done"`
	CreatedAt   string        `bson:"createdAt"`
	UpdatedAt   string        `bson:"updatedAt"`
}

type MemoryDB struct {
	mtx  sync.RWMutex
	data map[string]Todo
}

var db MemoryDB

func main() {
	// init db
	db = MemoryDB{
		mtx:  sync.RWMutex{},
		data: make(map[string]Todo),
	}

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Build the path to your static HTML file
		htmlPath := filepath.Join("templates", "index.html")

		// Read the file
		html, err := os.ReadFile(htmlPath)
		if err != nil {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		// Set content type and write to response
		w.Header().Set("Content-Type", "text/html")
		w.Write(html)
	})
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1/todos", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				var request TodoDTO
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					log.Printf("Failed to decode request: %v", err)
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}

				db.mtx.Lock()
				defer db.mtx.Unlock()

				todo := Todo{
					Id:          bson.NewObjectID(),
					Title:       request.Title,
					Description: request.Description,
					Done:        false,
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				}
				db.data[todo.Id.Hex()] = todo

				response := TodoDTO{
					Id:          todo.Id.Hex(),
					Title:       todo.Title,
					Description: todo.Description,
					Done:        todo.Done,
					CreatedAt:   todo.CreatedAt,
					UpdatedAt:   todo.UpdatedAt,
				}

				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(&response)
			})
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {

				db.mtx.RLock()
				defer db.mtx.RUnlock()

				var todos []Todo
				for _, todo := range db.data {
					todos = append(todos, todo)
				}

				var result []TodoDTO
				for _, todo := range todos {
					result = append(result, TodoDTO{
						Id:          todo.Id.Hex(),
						Title:       todo.Title,
						Description: todo.Description,
						Done:        todo.Done,
						CreatedAt:   todo.CreatedAt,
						UpdatedAt:   todo.UpdatedAt,
					})
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&result)
			})
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")

				db.mtx.RLock()
				defer db.mtx.RUnlock()

				todo, exists := db.data[id]
				if !exists {
					http.Error(w, "todo not found", http.StatusNotFound)
					return
				}

				result := TodoDTO{
					Id:          todo.Id.Hex(),
					Title:       todo.Title,
					Description: todo.Description,
					Done:        todo.Done,
					CreatedAt:   todo.CreatedAt,
					UpdatedAt:   todo.UpdatedAt,
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&result)
			})
			r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")

				var request TodoDTO
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					log.Printf("Failed to decode request: %v", err)
					http.Error(w, "something went wrong", http.StatusBadRequest)
					return
				}

				db.mtx.Lock()
				defer db.mtx.Unlock()

				todo, exists := db.data[id]
				if !exists {
					http.Error(w, "todo not found", http.StatusNotFound)
					return
				}

				todo.Title = request.Title
				todo.Description = request.Description
				todo.Done = request.Done
				todo.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

				db.data[id] = todo

				response := TodoDTO{
					Id:          todo.Id.Hex(),
					Title:       todo.Title,
					Description: todo.Description,
					Done:        todo.Done,
					CreatedAt:   todo.CreatedAt,
					UpdatedAt:   todo.UpdatedAt,
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(&response)
			})

			r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")

				db.mtx.Lock()
				_, exists := db.data[id]
				db.mtx.Unlock()

				if !exists {
					http.Error(w, "todo not found", http.StatusNotFound)
					return
				}

				delete(db.data, id)

				w.WriteHeader(http.StatusNoContent)
			})
		})
	})

	srv := http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", "8080"),
		Handler: r,
	}

	log.Printf("Running server on %s...", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}
