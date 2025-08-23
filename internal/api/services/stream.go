package services

import (
	"fmt"
	"os"
	"sync"
)

// StreamData はフロントエンドに返すデータ型
type StreamData struct {
	ThumbnailUrl string `json:"thumbnailUrl"`
	Title        string `json:"title"`
	ChannelName  string `json:"channelName"`
	DateTime     string `json:"dateTime"`
	Status       string `json:"status"`
	StreamUrl   string `json:"streamUrl"`
	VideoID     string `json:"videoId"`
	Platform    string `json:"platform"`
}

// 登録されたチャンネルのストリームをすべて取得し、統合する
func GetCombinedStreams(userID string) ([]StreamData, error) {
	// ダミーのチャンネルID使用
	youtubeChannelIDs := []string{"UC...", "UC..."} // 本来はfirestoreから
	twitchChannelIDs := []string{"streamer1", "streamer2"}   // 本来はfirestoreから

	var allStreams []StreamData
	var wg sync.WaitGroup
	var mu sync.Mutex

	errChan := make(chan error, 2)

	// YouTube配信取得
	wg.Add(1)
	go func() {
		defer wg.Done()
		streams, err := GetYouTubeStreams(youtubeChannelIDs)
		if err != nil {
			errChan <- fmt.Errorf("YouTubeストリーム取得エラー: %w", err)
			return
		}
		mu.Lock()
		allStreams = append(allStreams, streams...)
		mu.Unlock()
	}()

	// Twitch配信取得
	wg.Add(1)
	go func() {
		defer wg.Done()
		streams, err := GetTwitchStreams(twitchChannelIDs)
		if err != nil {
			errChan <- fmt.Errorf("Twitchストリーム取得エラー: %w", err)
			return
		}
		mu.Lock()
		allStreams = append(allStreams, streams...)
		mu.Unlock()
	}()

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan // 任意のエラー返す
	}

	return allStreams, nil
}
