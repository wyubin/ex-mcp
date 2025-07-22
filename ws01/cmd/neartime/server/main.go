package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func wsHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")
	fmt.Printf("sessionID: %s\n", sessionID)
	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()

	// create file
	pathTmpFile := fmt.Sprintf("/tmp/ws01/%s.pcm", sessionID)
	pathFile := fmt.Sprintf("/workspaces/ex-mcp/ws01/tmp/audio-backup/%s.pcm", sessionID)
	f, err := os.Create(pathTmpFile)
	if err != nil {
		fmt.Println("file create error:", err)
		return
	}
	defer f.Close()

	for {
		op, data, err := conn.Read(ctx)

		if op == websocket.MessageBinary {

			// 暫存音檔
			nbyte, err := f.Write(data)
			if err != nil {
				fmt.Println("file write error:", err)
				break
			}

			fmt.Printf("Session [%s][code: %d] received: %d bytes\n", sessionID, op, nbyte)
		}
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			f.Close()
			err := os.Rename(pathTmpFile, pathFile)
			if err != nil {
				fmt.Printf("err : %s\n", err)
			}
			break
		}
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}
	}
	// c.JSON will error
	// c.JSON(http.StatusBadRequest, "WsStreamTranscriptions finish")
}

func main() {
	r := gin.Default()
	r.GET("/ws/asr/neartime/:sessionId", wsHandler)
	r.Run(":8080")
}
