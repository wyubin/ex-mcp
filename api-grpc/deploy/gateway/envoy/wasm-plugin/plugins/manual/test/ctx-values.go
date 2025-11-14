package test

import (
	"context"
	"errors"
	"net/url"
)

var (
	errorNotFoundByKey = errors.New("no value in context")
)

type ctxParamKeys struct{}

func CtxWithParamKeys(ctx context.Context, keys []string) context.Context {
	return context.WithValue(ctx, ctxParamKeys{}, keys)
}
func ParamKeysFromContext(ctx context.Context) ([]string, error) {
	if val, ok := ctx.Value(ctxParamKeys{}).([]string); ok {
		return val, nil
	}
	return nil, errorNotFoundByKey
}

type ctxInfoRequest struct{}

type InfoRequest struct {
	ClusterName string            `json:"clusterName"`
	PathGrpc    string            `json:"pathGrpc"`
	PathParams  map[string]string `json:"pathParams"`
	Query       url.Values        `json:"query"`
}

func CtxWithInfoRequest(ctx context.Context, val InfoRequest) context.Context {
	return context.WithValue(ctx, ctxInfoRequest{}, val)
}
func InfoRequestFromContext(ctx context.Context) (InfoRequest, error) {
	if val, ok := ctx.Value(ctxInfoRequest{}).(InfoRequest); ok {
		return val, nil
	}
	return InfoRequest{}, errorNotFoundByKey
}
