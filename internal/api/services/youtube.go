package services

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"
)

type YoutubeSearchResponse struct {
    Items []struct {
        ID struct {
            VideoID string `json:"videoId"`
        } `json:"id"`
    } `json:"items"`
}

type YoutubeVideosResponse struct {
    Items []struct {
        ID      string `json:"id"`
        Snippet struct {
            PublishedAt   time.Time `json:"publishedAt"`
            ChannelTitle  string    `json:"channelTitle"`
            Title         string    `json:"title"`
            Thumbnails    struct {
                Default struct {
                    URL string `json:"url"`
                } `json:"default"`
                Medium struct {
                    URL string `json:"url"`
                } `json:"medium"`
                High struct {
                    URL string `json:"url"`
                } `json:"high"`
            } `json:"thumbnails"`
        } `json:"snippet"`
        LiveStreamingDetails struct {
            ActualStartTime    time.Time `json:"actualStartTime"`
            ActualEndTime      time.Time `json:"actualEndTime"`
            ScheduledStartTime time.Time `json:"scheduledStartTime"`
        } `json:"liveStreamingDetails"`
    } `json:"items"`
}

// GetYoutubeStreams は YouTube API からストリーム情報を取得する
func GetYoutubeStreams(channelIDs []string) ([]StreamData, error) {
    if len(channelIDs) == 0 {
        return []StreamData{}, nil
    }
    
    youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")
    if youtubeAPIKey == "" {
        return nil, fmt.Errorf("YouTube APIキーが設定されていません")
    }

    // チャンネル内の動画IDを検索
    searchURL := fmt.Sprintf(
        "https://www.googleapis.com/youtube/v3/search?part=id,snippet&channelId=%s&type=video&order=date&maxResults=50&key=%s",
        strings.Join(channelIDs, ","), youtubeAPIKey,
    )
    resp, err := http.Get(searchURL)
    if err != nil {
        return nil, fmt.Errorf("YouTube検索APIの呼び出しに失敗: %w", err)
    }
    defer resp.Body.Close()

    var searchData YoutubeSearchResponse
    if err := json.NewDecoder(resp.Body).Decode(&searchData); err != nil {
        return nil, fmt.Errorf("YouTube検索レスポンスのデコードに失敗: %w", err)
    }
    
    videoIDs := make([]string, 0, len(searchData.Items))
    for _, item := range searchData.Items {
        videoIDs = append(videoIDs, item.ID.VideoID)
    }
    
    if len(videoIDs) == 0 {
        return []StreamData{}, nil
    }

    // 動画詳細情報を取得
    videosURL := fmt.Sprintf(
        "https://www.googleapis.com/youtube/v3/videos?part=snippet,liveStreamingDetails&id=%s&key=%s",
        strings.Join(videoIDs, ","), youtubeAPIKey,
    )
    resp, err = http.Get(videosURL)
    if err != nil {
        return nil, fmt.Errorf("YouTube動画APIの呼び出しに失敗: %w", err)
    }
    defer resp.Body.Close()
    
    var videosData YoutubeVideosResponse
    if err := json.NewDecoder(resp.Body).Decode(&videosData); err != nil {
        return nil, fmt.Errorf("YouTube動画レスポンスのデコードに失敗: %w", err)
    }

    streams := make([]StreamData, 0, len(videosData.Items))
    for _, item := range videosData.Items {
        liveDetails := item.LiveStreamingDetails
        
        // liveStreamingDetailsが存在しない場合はライブ配信ではないためスキップ
        if liveDetails == (struct {
            ActualStartTime    time.Time "json:\"actualStartTime\""
            ActualEndTime      time.Time "json:\"actualEndTime\""
            ScheduledStartTime time.Time "json:\"scheduledStartTime\""
        }{}) {
            continue
        }
        
        var status string
        var dateTime time.Time

        if !liveDetails.ActualEndTime.IsZero() {
            status = "ended"
            dateTime = liveDetails.ActualEndTime
        } else if !liveDetails.ActualStartTime.IsZero() {
            status = "live"
            dateTime = liveDetails.ActualStartTime
        } else if !liveDetails.ScheduledStartTime.IsZero() {
            status = "upcoming"
            dateTime = liveDetails.ScheduledStartTime
        } else {
            // ライブ配信情報がない場合はスキップ
            continue
        }

        thumbnailURL := item.Snippet.Thumbnails.High.URL
        if thumbnailURL == "" {
            thumbnailURL = item.Snippet.Thumbnails.Medium.URL
        }
        if thumbnailURL == "" {
            thumbnailURL = item.Snippet.Thumbnails.Default.URL
        }

        streams = append(streams, StreamData{
            VideoID:      item.ID,
            ThumbnailUrl: thumbnailURL,
            Title:        item.Snippet.Title,
            ChannelName:  item.Snippet.ChannelTitle,
            DateTime:     dateTime.Format(time.RFC3339),
            Status:       status,
            StreamUrl:    fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID),
            Platform:     "youtube",
        })
    }
    
    return streams, nil
}