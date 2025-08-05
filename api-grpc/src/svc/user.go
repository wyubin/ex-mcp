package svc

import (
	"context"
	"fmt"
	"sync"

	userv1 "github.com/wyubin/ex-mcp/api-grpc/src/gen/user/v1"
)

type UserService struct {
	userv1.UnimplementedUserServiceServer
	mu    sync.RWMutex
	users map[string]*userv1.User
	idSeq int
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*userv1.User),
		idSeq: 1,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("%d", s.idSeq)
	s.idSeq++

	user := &userv1.User{
		Id:    id,
		Name:  req.Name,
		Email: req.Email,
	}
	s.users[id] = user

	return &userv1.CreateUserResponse{User: user}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[req.Id]
	if !ok {
		return nil, fmt.Errorf("user with id %s not found", req.Id)
	}

	return &userv1.GetUserResponse{User: user}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userList []*userv1.User
	for _, user := range s.users {
		userList = append(userList, user)
	}

	return &userv1.ListUsersResponse{
		Users: userList,
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[req.Id]; !ok {
		return nil, fmt.Errorf("user with id %s not found", req.Id)
	}

	delete(s.users, req.Id)
	return &userv1.DeleteUserResponse{}, nil
}
