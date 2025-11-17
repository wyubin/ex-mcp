package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"main/internal/grpcio"

	"google.golang.org/protobuf/proto"
)

type UserGrpcMapper struct {
	*grpcio.RequestMapper
}

func NewUserGrpcMapper(name string) *UserGrpcMapper {
	updater := UserGrpcMapper{
		RequestMapper: grpcio.NewRequestMapper(name),
	}
	updater.RegisterRoutes()
	return &updater
}

func (s *UserGrpcMapper) RegisterRoutes() {
	s.Register("GET /v1/user/{id}", "/user.UserService/GetUser", &covUserGet{})
	s.Register("POST /v1/user", "/user.UserService/CreateUser", &covUserCreate{})
}

type covUserCreate struct{}

func (s *covUserCreate) Json2Grpc(info grpcio.InfoRequest, jsonBody []byte) ([]byte, error) {
	var req CreateUserRequest
	if err := json.Unmarshal(jsonBody, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON for CreateUser\n->%w", err)
	}
	grpcData, err := proto.Marshal(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gRPC request\n->%w", err)
	}
	return grpcio.GrpcFrame(grpcData), nil
}

func (s *covUserCreate) Grpc2Json(info grpcio.InfoRequest, grpcBody []byte) ([]byte, error) {
	if len(grpcBody) < 5 {
		return nil, errors.New("grpc body not enough")
	}
	grpcPayload := grpcBody[5:]
	var resp CreateUserResponse
	err := proto.Unmarshal(grpcPayload, &resp)
	if err != nil {
		return nil, fmt.Errorf("grpcBody encode error\n->%w", err)
	}
	jsonData, _ := json.Marshal(&resp)
	return jsonData, nil
}

type covUserGet struct{}

func (s *covUserGet) Json2Grpc(info grpcio.InfoRequest, jsonBody []byte) ([]byte, error) {
	// check user id
	id, found := info.PathParams["id"]
	if !found {
		return nil, fmt.Errorf("%s field not exist", "id")
	}
	req := GetUserRequest{Id: id}
	grpcData, err := proto.Marshal(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gRPC request\n->%w", err)
	}
	return grpcio.GrpcFrame(grpcData), nil
}

func (s *covUserGet) Grpc2Json(info grpcio.InfoRequest, grpcBody []byte) ([]byte, error) {
	if len(grpcBody) < 5 {
		return nil, errors.New("grpc body not enough")
	}
	grpcPayload := grpcBody[5:]
	var resp GetUserResponse
	err := proto.Unmarshal(grpcPayload, &resp)
	if err != nil {
		return nil, fmt.Errorf("grpcBody encode error\n->%w", err)
	}
	jsonData, _ := json.Marshal(&resp)
	return jsonData, nil
}
