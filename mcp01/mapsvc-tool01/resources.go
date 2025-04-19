package main

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

func resProfile(name string) mcp.ResourceTemplate {
	return mcp.NewResourceTemplate(
		"users://{id}/profile",
		name,
		mcp.WithTemplateDescription("Returns user profile information"),
		mcp.WithTemplateMIMEType("application/json"),
	)
}

func resProfileHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// get user id from request
	url := request.Params.URI
	profile, err := os.ReadFile(url) // Your DB/API call here
	if err != nil {
		return nil, err
	}
	// return user profile
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(profile),
		},
	}, nil
}
