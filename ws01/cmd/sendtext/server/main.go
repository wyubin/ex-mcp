package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

var (
	session2msgs = map[string][]string{}
)

func wsHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")
	fmt.Printf("sessionID: %s\n", sessionID)
	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	fmt.Printf("created conn\n")
	if err != nil {
		log.Printf("failed to upgrade connection: %v", err)
		return
	}
	defer conn.CloseNow()

	session2msgs[sessionID] = []string{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	for {
		op, reader, err := conn.Reader(ctx)

		if op == websocket.MessageText {
			textByte, _ := io.ReadAll(reader)

			// 暫存訊息
			cacheMsgs(sessionID, string(textByte))

			fmt.Printf("Session [%s][code: %d] received: %s\n", sessionID, op, textByte)
		}
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			fmt.Printf("session[%s] has %d msgs\n", sessionID, len(session2msgs[sessionID]))
			for idx, msg := range session2msgs[sessionID] {
				fmt.Printf("msg[%d]: %s\n", idx+1, msg)
			}
			break
		}
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}
	}
}

// cacheMsgs
func cacheMsgs(sessionID, msg string) {
	session2msgs[sessionID] = append(session2msgs[sessionID], msg)
}

func main() {
	r := gin.Default()
	r.GET("/ws/asr/neartime/:sessionId", wsHandler)
	r.Run(":8080")
}
