package grpcio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type Route interface {
	SetClusterName(name string)
	Name() string
	MatchInfo(method, path string) (InfoRequest, error)
	RequestCov(info InfoRequest, jsonBody []byte) ([]byte, error)
	ResponseCov(info InfoRequest, grpcBody []byte) ([]byte, error)
}

type RequestMapper struct {
	name        string
	clusterName string
	mux         *http.ServeMux
	path2Cov    map[string]BodyCov
}

func NewRequestMapper(name string) *RequestMapper {
	return &RequestMapper{
		name:     name,
		mux:      http.NewServeMux(),
		path2Cov: map[string]BodyCov{},
	}
}

// implement by hoster
func (s *RequestMapper) RegisterRoutes() {
	// s.Register("GET /v1/user/{id}", "/user.UserService/GetUser")
	// s.Register("POST /v1/user", "/user.UserService/CreateUser")
}

// shared method
func (s *RequestMapper) SetClusterName(name string) {
	s.clusterName = name
}

func (s *RequestMapper) Name() string {
	return s.name
}

func (s *RequestMapper) Register(pattern, grpcPath string, converter BodyCov) {
	s.mux.HandleFunc(pattern, s.handle(grpcPath))
	s.path2Cov[grpcPath] = converter
}

func (s *RequestMapper) MatchInfo(method, path string) (InfoRequest, error) {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		return InfoRequest{}, errors.New(w.Body.String())
	}
	var res InfoRequest
	json.Unmarshal(w.Body.Bytes(), &res)
	return res, nil
}

func (s *RequestMapper) RequestCov(info InfoRequest, jsonBody []byte) ([]byte, error) {
	converter, found := s.path2Cov[info.PathGrpc]
	if !found {
		return nil, fmt.Errorf("no path match for %s", info.PathGrpc)
	}
	return converter.Json2Grpc(info, jsonBody)
}

func (s *RequestMapper) ResponseCov(info InfoRequest, grpcBody []byte) ([]byte, error) {
	converter, found := s.path2Cov[info.PathGrpc]
	if !found {
		return nil, fmt.Errorf("no path match for %s", info.PathGrpc)
	}
	return converter.Grpc2Json(info, grpcBody)
}

// -- private method --  //
func (s *RequestMapper) handle(pathGrpc string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		info, err := s.ctxUpdateParamkeys(r, pathGrpc)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
		infoByte, err := json.Marshal(info)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
		w.Write(infoByte)
	}
}

func (s *RequestMapper) ctxUpdateParamkeys(req *http.Request, pathGrpc string) (InfoRequest, error) {
	// get pattern and parse keys
	handler, pattern := s.mux.Handler(req)
	if handler == nil {
		return InfoRequest{}, fmt.Errorf("no route match for %s %s", req.Method, req.URL.Path)
	}
	paramKeys := extractParamkeys(pattern)
	info := InfoRequest{
		ClusterName: s.clusterName,
		PathGrpc:    pathGrpc,
		PathParams:  mapRequestPathValue(req, paramKeys),
		Query:       req.URL.Query(),
	}

	return info, nil
}
