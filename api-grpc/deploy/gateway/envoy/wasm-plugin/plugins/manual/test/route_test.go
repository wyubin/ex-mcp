package test

import (
	"fmt"
	"testing"

	"main/internal/grpcio"
	userpb "main/proto/user"

	"google.golang.org/protobuf/proto"
)

const (
	prefixPath = "/user.UserService"
)

func TestRoute(t *testing.T) {
	routes := grpcio.NewRoutes()
	covUser := userpb.NewUserGrpcMapper("user")
	covUser.SetClusterName("user_service")
	routes.Add(covUser)

	route, info := routes.Match("GET", "/v1/user/42?name=yubin")
	if route != nil {
		fmt.Printf("info[Get]: %+v;\n", info)
	} else {
		fmt.Printf("info[Get] no match \n")
	}

	grpcReq, err := route.RequestCov(info, nil)
	fmt.Printf("grpcReq: %s; err: %s \n", grpcReq, err)

	req := userpb.GetUserResponse{User: &userpb.User{Id: "01", Name: "test-name", Email: "test@google.com"}}
	grpcData, _ := proto.Marshal(&req)
	jsonResp, err := route.ResponseCov(info, grpcio.GrpcFrame(grpcData))
	fmt.Printf("jsonResp: %s; err: %s \n", jsonResp, err)

}
