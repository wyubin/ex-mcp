package grpcio

import "net/url"

type InfoRequest struct {
	ClusterName string            `json:"clusterName"`
	PathGrpc    string            `json:"pathGrpc"`
	PathParams  map[string]string `json:"pathParams"`
	Query       url.Values        `json:"query"`
}

type BodyCov interface {
	Json2Grpc(info InfoRequest, jsonBody []byte) ([]byte, error)
	Grpc2Json(info InfoRequest, grpcBody []byte) ([]byte, error)
}
