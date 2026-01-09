package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	redirectURL = "urn:ietf:wg:oauth:2.0:oob" // 本地授權用
	tokenFile   = "token.json"
)

// 建立 OAuth2 配置
func getOAuthConfig() *oauth2.Config {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatalf("請設定環境變數 GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET")
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveFileScope},
		RedirectURL:  redirectURL,
	}
}

// 取得 token
func getClient(config *oauth2.Config) *http.Client {
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("請開啟瀏覽器授權後輸入驗證碼:\n%v\n", authURL)

	var code string
	fmt.Print("輸入驗證碼: ")
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("讀取驗證碼錯誤: %v", err)
	}

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("授權交換 token 失敗: %v", err)
	}
	return tok
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("儲存授權 token 到 %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("無法儲存 token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()

	config := getOAuthConfig()
	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("無法建立 Drive 服務: %v", err)
	}

	f := &drive.File{
		Name:     "hello.txt",
		MimeType: "text/plain",
	}

	content := "Hello World"
	file, err := srv.Files.Create(f).Media(strings.NewReader(content)).Do()
	if err != nil {
		log.Fatalf("無法建立檔案: %v", err)
	}

	fmt.Printf("檔案已建立: ID=%s\n", file.Id)
}
