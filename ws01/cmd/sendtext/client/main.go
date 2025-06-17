package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/coder/websocket"
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

	// 連線到 WebSocket 伺服器
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.CloseNow()

	// 開始傳送訊息
	for i := 0; i < *msgCount; i++ {
		msg := fmt.Sprintf("Message %d at %s", i+1, time.Now().Format(time.RFC3339))
		log.Printf("Sending: %s", msg)

		err := conn.Write(ctx, websocket.MessageText, []byte(msg))
		if err != nil {
			log.Fatalf("Write error: %v", err)
		}

		time.Sleep(1 * time.Second)
	}

	log.Println("Sending close frame...")
	conn.Close(websocket.StatusNormalClosure, "done")
}
