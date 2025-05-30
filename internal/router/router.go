package router

import (
	"log"
	"net/http"

	"go-sample/internal/handlers"

	"github.com/gorilla/mux"
)

func SetupRouter(userHandler *handlers.UserHandler, teamHandler *handlers.TeamHandler, importHandler *handlers.ImportHandler) *mux.Router {
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	// User routes
	router.HandleFunc("/api/users", userHandler.Create).Methods("POST")
	router.HandleFunc("/api/users/{id}", userHandler.Update).Methods("PUT")
	router.HandleFunc("/api/users/{id}", userHandler.Delete).Methods("DELETE")
	router.HandleFunc("/api/users/{id}", userHandler.GetByID).Methods("GET")
	router.HandleFunc("/api/users", userHandler.List).Methods("GET")

	// Team routes
	router.HandleFunc("/api/teams", teamHandler.Create).Methods("POST")
	router.HandleFunc("/api/teams/{id}", teamHandler.Update).Methods("PUT")
	router.HandleFunc("/api/teams/{id}", teamHandler.Delete).Methods("DELETE")
	router.HandleFunc("/api/teams/{id}", teamHandler.GetByID).Methods("GET")
	router.HandleFunc("/api/teams", teamHandler.List).Methods("GET")
	router.HandleFunc("/api/teams/{id}/users", teamHandler.AddUser).Methods("POST")

	// Import route
	router.HandleFunc("/api/import", importHandler.ImportCSV).Methods("POST")

	// Add a health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return router
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
} 