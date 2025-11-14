package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

const (
	prefixPath = "/user.UserService"
)

func TestRoute(t *testing.T) {
	updater := NewCtxUpdater()
	updater.RegisterRoutes()
	updater.SetClusterName("user_service")

	req := httptest.NewRequest("GET", "/v1/user/42?name=yubin", nil)
	w := httptest.NewRecorder()

	updater.Route.ServeHTTP(w, req)
	fmt.Printf("w.Code: %d\n", w.Code)
	fmt.Printf("w.Body: %s\n", w.Body)
}

type CtxUpdater struct {
	Route       *http.ServeMux
	name        string
	clusterName string
}

func (s *CtxUpdater) Name() string {
	return s.name
}

func (s *CtxUpdater) SetClusterName(name string) {
	s.clusterName = name
}

func (s *CtxUpdater) RegisterRoutes() {
	r := http.NewServeMux()

	r.HandleFunc("POST /v1/user", s.Handle("/user.UserService/CreateUser"))
	r.HandleFunc("GET /v1/user/{id}", s.Handle("/user.UserService/GetUser"))
	s.Route = r
}

func (s *CtxUpdater) ctxUpdateParamkeys(req *http.Request, pathGrpc string) InfoRequest {
	// get pattern and parse keys
	_, pattern := s.Route.Handler(req)
	paramKeys := extractParamkeys(pattern)
	info := InfoRequest{
		ClusterName: s.clusterName,
		PathGrpc:    pathGrpc,
		PathParams:  PathValuesMap(req, paramKeys),
		Query:       req.URL.Query(),
	}

	return info
}

func (s *CtxUpdater) Handle(pathGrpc string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		info := s.ctxUpdateParamkeys(r, pathGrpc)
		infoByte, err := json.Marshal(info)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		w.Write(infoByte)
	}
}

func NewCtxUpdater() *CtxUpdater {
	inst := CtxUpdater{name: "user"}
	return &inst
}

func PathValuesMap(req *http.Request, keys []string) map[string]string {
	res := map[string]string{}
	if keys == nil {
		return res
	}
	for _, keyTmp := range keys {
		res[keyTmp] = req.PathValue(keyTmp)
	}
	return res
}

func extractParamkeys(pattern string) []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)

	var params []string
	for _, m := range matches {
		params = append(params, m[1])
	}
	return params
}
