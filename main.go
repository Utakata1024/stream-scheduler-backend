package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"stream-scheduler-backend/internal/api"
)

func main() {
	// .envから環境変数を読み込む
	if err := godotenv.Load(); err != nil {
		log.Println("警告： .envファイルが見つかりません。")
	}

	// APIハンドラーを初期化
	http.HandleFunc("/api/streams", api.StreamsHandler)
	http.HandleFunc("/api/channels", api.ChannelsHandler)

	// サーバー起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("サーバーをポート %s で起動中...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
