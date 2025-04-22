package serv

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// 協助建立一個sql 的query
func promptSqlQuery(name string) mcp.Prompt {
	return mcp.NewPrompt(name,
		mcp.WithPromptDescription("SQL query builder assistance"),
		mcp.WithArgument("table",
			mcp.ArgumentDescription("Name of the table to query"),
			mcp.RequiredArgument(),
		),
	)
}

func promptSqlQueryHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	tableName := request.Params.Arguments["table"]
	if tableName == "" {
		return nil, fmt.Errorf("table name is required")
	}

	return mcp.NewGetPromptResult(
		"SQL query builder assistance",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleAssistant,
				mcp.NewTextContent("You are a SQL expert. Help construct efficient and safe queries."),
			),
			mcp.NewPromptMessage(
				mcp.RoleAssistant,
				mcp.NewEmbeddedResource(mcp.BlobResourceContents{
					URI:      fmt.Sprintf("db://schema/%s", tableName),
					MIMEType: "application/json",
				}),
			),
		},
	), nil
}
