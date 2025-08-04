package svc

import (
	"context"
	"fmt"
	userv1 "gen/user/v1"

	"connectrpc.com/connect"
)

type UserService struct {
	userv1.UnimplementedUserServiceHandler
	users map[string]*userv1.User
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*userv1.User),
	}
}

func (s *UserService) CreateUser(
	ctx context.Context,
	req *connect.Request[userv1.CreateUserRequest],
) (*connect.Response[userv1.CreateUserResponse], error) {
	id := fmt.Sprintf("%d", len(s.users)+1)
	user := &userv1.User{
		Id:    id,
		Name:  req.Msg.Name,
		Email: req.Msg.Email,
	}
	s.users[id] = user
	return connect.NewResponse(&userv1.CreateUserResponse{User: user}), nil
}

func (s *UserService) GetUser(
	ctx context.Context,
	req *connect.Request[userv1.GetUserRequest],
) (*connect.Response[userv1.GetUserResponse], error) {
	user, ok := s.users[req.Msg.Id]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}
	return connect.NewResponse(&userv1.GetUserResponse{User: user}), nil
}

func (s *UserService) ListUsers(
	ctx context.Context,
	_ *connect.Request[userv1.ListUsersRequest],
) (*connect.Response[userv1.ListUsersResponse], error) {
	var list []*userv1.User
	for _, u := range s.users {
		list = append(list, u)
	}
	return connect.NewResponse(&userv1.ListUsersResponse{Users: list}), nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	req *connect.Request[userv1.DeleteUserRequest],
) (*connect.Response[userv1.DeleteUserResponse], error) {
	delete(s.users, req.Msg.Id)
	return connect.NewResponse(&userv1.DeleteUserResponse{}), nil
}
