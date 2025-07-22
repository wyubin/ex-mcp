package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	session2msgs = map[string][]string{}
)

// 建立一個 websocket 的 upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // production 時建議驗證 Origin
	},
}

func handleWebSocket(c *gin.Context) {
	sessionID := c.Param("sessionId") // 取得 URL 中的 sessionId
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to upgrade:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Client connected with sessionID: %s\n", sessionID)
	rwLock := sync.RWMutex{}

	for {
		// 讀取 client 傳來的訊息
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Session [%s][code:%d] Read error: %v\n", sessionID, msgType, err)
			break
		}

		fmt.Printf("Received from [%s][code:%d]: %s\n", sessionID, msgType, string(msg))

		// 回傳 echo 訊息
		err = conn.WriteMessage(msgType, []byte("Echo from server: "+string(msg)))
		if err != nil {
			fmt.Printf("Session [%s] Write error: %v\n", sessionID, err)
			break
		}
		rwLock.Lock()
		cacheMsgs(sessionID, string(msg))
		rwLock.Unlock()
	}

	fmt.Printf("Client disconnected: %s\n", sessionID)
}

// cacheMsgs
func cacheMsgs(sessionID, msg string) {
	session2msgs[sessionID] = append(session2msgs[sessionID], msg)
}

func main() {
	r := gin.Default()
	r.GET("/ws/asr/neartime/:sessionId", handleWebSocket)
	r.Run(":8080")
}
