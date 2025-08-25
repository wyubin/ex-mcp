package user

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	userpb "github.com/wyubin/ex-mcp/api-grpc/src/gen/user/v1"
)

type Config struct {
	GRPCAddr string `json:"grpcAddr"` // ex: "localhost:50051"
}

func CreateConfig() *Config {
	return &Config{
		GRPCAddr: "localhost:50051",
	}
}

type UserGateway struct {
	next   http.Handler
	client userpb.UserServiceClient
	mux    *http.ServeMux
}

var once sync.Once
var conn *grpc.ClientConn

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	var err error
	once.Do(func() {
		conn, err = grpc.NewClient(config.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	})
	if err != nil {
		return nil, err
	}

	client := userpb.NewUserServiceClient(conn)
	mux := http.NewServeMux()

	g := &UserGateway{
		next:   next,
		client: client,
		mux:    mux,
	}

	// Bind paths to handlers
	mux.Handle("POST /user/v1/user", http.HandlerFunc(g.handleCreateUser))
	mux.Handle("GET /user/v1/users", http.HandlerFunc(g.handleListUsers))
	mux.Handle("GET /user/v1/user/{id}", http.HandlerFunc(g.handleGetUsers))
	mux.Handle("DELETE /user/v1/user/{id}", http.HandlerFunc(g.handleDeleteUser))

	return g, nil
}

func (p *UserGateway) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if _, pattern := p.mux.Handler(req); pattern != "" {
		p.mux.ServeHTTP(rw, req)
	} else {
		p.next.ServeHTTP(rw, req)
	}
}

// ---------------- Handlers ----------------

func (p *UserGateway) handleCreateUser(rw http.ResponseWriter, req *http.Request) {
	var createReq userpb.CreateUserRequest
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		http.Error(rw, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := p.client.CreateUser(req.Context(), &createReq)
	writeResponse(rw, resp, err)
}

func (p *UserGateway) handleListUsers(rw http.ResponseWriter, req *http.Request) {
	resp, err := p.client.ListUsers(req.Context(), &userpb.ListUsersRequest{})
	writeResponse(rw, resp, err)
}

func (p *UserGateway) handleGetUsers(rw http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(map[string]string{"error": "missing id field"})
		return
	}
	resp, err := p.client.GetUser(req.Context(), &userpb.GetUserRequest{Id: id})
	writeResponse(rw, resp, err)
}

func (p *UserGateway) handleDeleteUser(rw http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(map[string]string{"error": "missing id field"})
		return
	}
	resp, err := p.client.DeleteUser(req.Context(), &userpb.DeleteUserRequest{Id: id})
	writeResponse(rw, resp, err)
}

// ---------------- gRPC → HTTP Response Helper ----------------

func writeResponse(rw http.ResponseWriter, resp any, err error) {
	rw.Header().Set("Content-Type", "application/json")

	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(grpcCodeToHTTP(st.Code()))
		json.NewEncoder(rw).Encode(map[string]string{"error": st.Message()})
		return
	}

	json.NewEncoder(rw).Encode(resp)
}

func grpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
