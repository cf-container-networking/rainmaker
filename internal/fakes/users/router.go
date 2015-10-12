package users

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pivotal-cf-experimental/rainmaker/internal/fakes/domain"
)

func NewRouter(users *domain.Users, spaces *domain.Spaces) http.Handler {
	router := mux.NewRouter()

	router.Handle("/v2/users", listHandler{users, spaces}).Methods("GET")
	router.Handle("/v2/users", createHandler{users}).Methods("POST")
	router.Handle("/v2/users/{guid}", getHandler{users}).Methods("GET")

	return router
}