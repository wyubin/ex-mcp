package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	chunkSize = 32000 // 16kHz mono 1秒資料（32000 bytes = 2 bytes/sample * 16000 samples）
)

func main() {
	// CLI 參數
	sessionID := flag.String("token", "asura-token", "token from asura")
	filePath := flag.String("file", "audio.pcm", "Path to PCM audio file")
	host := flag.String("host", "localhost:8080", "WebSocket host (e.g., localhost:8080)")
	flag.Parse()

	// 建立 WS 連線
	u := url.URL{
		Scheme:   "wss",
		Host:     *host,
		Path:     "asura/neartime/ws/v3",
		RawQuery: fmt.Sprintf("token=%s", *sessionID),
	}

	log.Printf("Connecting to %s\n", u.String())

	// 連線到 WebSocket 伺服器
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.CloseNow()

	// 開啟檔案
	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatal("failed to open pcm:", err)
	}
	defer file.Close()

	buffer := make([]byte, chunkSize)

	log.Printf("Start streaming %s to %s with session ID %s", *filePath, *host, *sessionID)

	// 讀取 server 傳來的訊息
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			opCode, message, err := conn.Read(ctx)
			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
					log.Println("complete transcriptions...")
				} else {
					log.Println("read error:", err)
				}
				wg.Done()
				return
			}
			log.Printf("recv[code:%d]: %s", opCode, message)
		}
	}()

	for {
		n, err := file.Read(buffer)
		if err != nil {
			log.Println("streaming finished:", err)
			break
		}
		err = conn.Write(ctx, websocket.MessageBinary, buffer[:n])
		if err != nil {
			log.Println("write error:", err)
			break
		}
		log.Printf("send: %d bytes audio", len(buffer))
		time.Sleep(1 * time.Second)
	}

	log.Println("Client finished sending audio.")
	log.Println("Wait for transcriptions response...")
	wg.Wait()
	log.Println("Close websocket...")
	conn.Close(websocket.StatusNormalClosure, "done")

}
