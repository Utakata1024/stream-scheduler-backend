package services

import (
    "fmt"
    "sync"
)

// StreamData はフロントエンドに返す統一されたストリームデータの型
type StreamData struct {
    ThumbnailUrl string `json:"thumbnailUrl"`
    Title        string `json:"title"`
    ChannelName  string `json:"channelName"`
    DateTime     string `json:"dateTime"`
    Status       string `json:"status"`
    StreamUrl    string `json:"streamUrl"`
    VideoID      string `json:"videoId"`
    Platform     string `json:"platform"`
}

// GetCombinedStreams は登録されたチャンネルのストリームをすべて取得し、統合する
func GetCombinedStreams(userID string) ([]StreamData, error) {
    // Firestoreからチャンネルリストを取得
    userChannels, err := GetUserChannels(userID)
    if err != nil {
        fmt.Printf("ユーザーチャンネルの取得に失敗: %v\n", err)
        return nil, fmt.Errorf("ユーザーチャンネルの取得に失敗しました: %v", err)
    }

    var youtubeChannelIDs []string
    var twitchChannelIDs []string

    // 取得チャンネルをプラットフォームごとに分類
    for id, channel := range userChannels {
        if channel.Platform == "youtube" {
            youtubeChannelIDs = append(youtubeChannelIDs, id)
        } else if channel.Platform == "twitch" {
            twitchChannelIDs = append(twitchChannelIDs, id)
        }
    }

    var allStreams []StreamData
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    errChan := make(chan error, 2)

    // YouTubeストリームの取得
    wg.Add(1)
    go func() {
        defer wg.Done()
        streams, err := GetYoutubeStreams(youtubeChannelIDs)
        if err != nil {
            errChan <- fmt.Errorf("YouTubeストリームの取得に失敗: %w", err)
            return
        }
        mu.Lock()
        allStreams = append(allStreams, streams...)
        mu.Unlock()
    }()

    // Twitchストリームの取得
    wg.Add(1)
    go func() {
        defer wg.Done()
        streams, err := GetTwitchStreams(twitchChannelIDs)
        if err != nil {
            errChan <- fmt.Errorf("Twitchストリームの取得に失敗: %w", err)
            return
        }
        mu.Lock()
        allStreams = append(allStreams, streams...)
        mu.Unlock()
    }()

    wg.Wait()
    close(errChan)

    if len(errChan) > 0 {
        return nil, <-errChan
    }

    return allStreams, nil
}