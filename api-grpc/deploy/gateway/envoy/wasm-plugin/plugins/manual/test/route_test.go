package test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRoute(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/user/", hanCreateUser)
	mux.HandleFunc("GET /v1/user/:id", hanGetUser)
}

func hanCreateUser(http.ResponseWriter, *http.Request) {
	fmt.Printf("it's hanCreateUser")
}

func hanGetUser(http.ResponseWriter, *http.Request) {
	fmt.Printf("it's hanGetUser")
}
