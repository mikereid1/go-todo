package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"net"
	"net/http"
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

func main() {
	// init mongo
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Error connecting to mongo: %v", err)
	}
	defer client.Disconnect(context.Background())

	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1/todos", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				var request TodoDTO
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					log.Printf("Failed to decode request: %v", err)
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}

				todo := Todo{
					Id:          bson.NewObjectID(),
					Title:       request.Title,
					Description: request.Description,
					Done:        false,
					CreatedAt:   time.Now().UTC().Format(time.RFC3339),
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				}

				_, err = client.Database("todos").
					Collection("todos").
					InsertOne(r.Context(), todo)
				if err != nil {
					log.Printf("Failed to perist Todo: %v", err)
					http.Error(w, "something went wrong", http.StatusInternalServerError)
					return
				}

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
				cursor, err := client.Database("todos").Collection("todos").Find(r.Context(), bson.M{})
				if err != nil {
					http.Error(w, "something went wrong", http.StatusInternalServerError)
					return
				}
				defer cursor.Close(r.Context())

				var todos []Todo
				if err = cursor.All(r.Context(), &todos); err != nil {
					http.Error(w, "something went wrong", http.StatusInternalServerError)
					return
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

				var todo Todo
				err := client.Database("todos").Collection("todos").FindOne(r.Context(), bson.M{"id": id}).Decode(&todo)
				if err != nil {
					if errors.Is(err, mongo.ErrNoDocuments) {
						http.Error(w, "todo not found", http.StatusNotFound)
						return
					}
					log.Printf("Failed to decode request: %v", err)
					http.Error(w, "something went wrong", http.StatusInternalServerError)
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

				var todo Todo
				err = client.Database("todos").Collection("todos").FindOne(r.Context(), bson.M{"id": id}).Decode(&todo)
				if err != nil {
					if errors.Is(err, mongo.ErrNoDocuments) {
						http.Error(w, "todo not found", http.StatusNotFound)
						return
					}
					log.Printf("Failed to updated todo: %v", err)
					http.Error(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				todo.Title = request.Title
				todo.Description = request.Description
				todo.Done = request.Done
				todo.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

				_, err = client.Database("todos").Collection("todos").InsertOne(r.Context(), todo)
				if err != nil {
					log.Printf("Failed to update Todo: %v", err)
					http.Error(w, "something went wrong", http.StatusInternalServerError)
					return
				}

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
				result, err := client.Database("todos").Collection("todos").DeleteOne(r.Context(), bson.M{"id": id})
				if err != nil {
					log.Printf("Failed to delete todo: %v", err)
					http.Error(w, "something went wrong", http.StatusInternalServerError)
					return

				}

				if result.DeletedCount == 0 {
					http.Error(w, "todo not found", http.StatusNotFound)
					return
				}

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
