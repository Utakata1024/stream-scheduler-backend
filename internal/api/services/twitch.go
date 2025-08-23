package services

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
)

const (
    twitchAPIURL = "https://api.twitch.tv/helix"
    twitchAuthURL = "https://id.twitch.tv/oauth2"
)

// TwitchTokenResponse は Twitch の認証トークンレスポンス
type TwitchTokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
    TokenType   string `json:"token_type"`
}

// TwitchStreamsResponse は Twitch Streams API のレスポンス
type TwitchStreamsResponse struct {
    Data []struct {
        ID          string `json:"id"`
        UserID      string `json:"user_id"`
        UserName    string `json:"user_name"`
        Type        string `json:"type"`
        Title       string `json:"title"`
        StartedAt   string `json:"started_at"`
        ThumbnailURL string `json:"thumbnail_url"`
    } `json:"data"`
}

// getAppAccessToken は Twitch App Access Token を取得する
func getAppAccessToken(clientID, clientSecret string) (string, error) {
    url := fmt.Sprintf("%s/token?client_id=%s&client_secret=%s&grant_type=client_credentials", twitchAuthURL, clientID, clientSecret)

    resp, err := http.Post(url, "application/json", nil)
    if err != nil {
        return "", fmt.Errorf("Twitch認証APIの呼び出しに失敗: %w", err)
    }
    defer resp.Body.Close()

    var tokenResponse TwitchTokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
        return "", fmt.Errorf("認証レスポンスのデコードに失敗: %w", err)
    }

    if tokenResponse.AccessToken == "" {
        return "", fmt.Errorf("Twitch認証トークンが取得できませんでした")
    }

    return tokenResponse.AccessToken, nil
}

// GetTwitchStreams は Twitch API からライブ配信情報を取得する
func GetTwitchStreams(channelIDs []string) ([]StreamData, error) {
    if len(channelIDs) == 0 {
        return []StreamData{}, nil
    }

    clientID := os.Getenv("TWITCH_CLIENT_ID")
    clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")

    if clientID == "" || clientSecret == "" {
        return nil, fmt.Errorf("Twitch APIキーが設定されていません")
    }

    accessToken, err := getAppAccessToken(clientID, clientSecret)
    if err != nil {
        return nil, err
    }

    var queryParams []string
    for _, id := range channelIDs {
        queryParams = append(queryParams, fmt.Sprintf("user_id=%s", id))
    }

    url := fmt.Sprintf("%s/streams?%s", twitchAPIURL, strings.Join(queryParams, "&"))
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("Twitchリクエストの作成に失敗: %w", err)
    }
    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Client-Id", clientID)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("Twitch Streams APIの呼び出しに失敗: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("Twitch APIからのエラーレスポンス: %d", resp.StatusCode)
    }

    var streamsResponse TwitchStreamsResponse
    if err := json.NewDecoder(resp.Body).Decode(&streamsResponse); err != nil {
        return nil, fmt.Errorf("Twitchレスポンスのデコードに失敗: %w", err)
    }

    streams := make([]StreamData, 0, len(streamsResponse.Data))
    for _, s := range streamsResponse.Data {
        streams = append(streams, StreamData{
            VideoID:      s.ID,
            ThumbnailUrl: strings.Replace(strings.Replace(s.ThumbnailURL, "{width}", "480", 1), "{height}", "270", 1), // サムネイルURLを整形
            Title:        s.Title,
            ChannelName:  s.UserName,
            DateTime:     s.StartedAt,
            Status:       "live", // TwitchのAPIはライブ配信のみを返すため、statusは常に"live"
            StreamUrl:    fmt.Sprintf("https://www.twitch.tv/%s", s.UserName),
            Platform:     "twitch",
        })
    }

    return streams, nil
}