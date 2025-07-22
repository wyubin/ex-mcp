package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// 命令列參數定義
	session := flag.String("session", "", "session ID to use in the WebSocket path")
	msgCount := flag.Int("msgcount", 5, "number of messages to send before closing")
	host := flag.String("host", "localhost:8080", "WebSocket server host (e.g., localhost:8080)")
	flag.Parse()

	if *session == "" {
		fmt.Println("session is required")
		flag.Usage()
		os.Exit(1)
	}

	u := url.URL{
		Scheme: "ws",
		Host:   *host,
		Path:   fmt.Sprintf("/ws/asr/neartime/%s", *session),
	}

	log.Printf("Connecting to %s\n", u.String())

	// 建立 TCP 連線
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// 讀取 server 傳來的訊息
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	// 開始傳送訊息
	for i := 0; i < *msgCount; i++ {
		msg := fmt.Sprintf("Message %d at %s", i+1, time.Now().Format(time.RFC3339))
		log.Printf("Sending: %s", msg)

		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Fatalf("Write error: %v", err)
		}

		time.Sleep(1 * time.Second)
	}

	log.Println("Sending close frame...")
	// 發送 close 訊號
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
	if err != nil {
		log.Println("close error:", err)
		return
	}
}
