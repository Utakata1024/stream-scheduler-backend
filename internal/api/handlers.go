package api

import (
	"encoding/json"
	"net/http"
	"stream-scheduler-backend/internal/api/services"
)

// /api/streams のハンドラー
func StreamsHandler(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダー設定
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Content-Type", "application/json")

	// ここでフロントエンドから渡されたユーザーIDを取得
	// 現状はダミー
	userID := "dummy-user-id"

	// ストリーム情報を取得
	streams, err := services.GetStreamsByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンスを返す
	if err := json.NewEncoder(w).Encode(streams); err != nil {
		http.Error(w, "JSONエンコードエラー", http.StatusInternalServerError)
		return
	}
}

// /api/channels のハンドラー
func ChannelsHandler(w http.ResponseWriter, r *http.Request) {
}
