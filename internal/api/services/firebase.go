package services

import (
    "context"
    "fmt"
    "os"
    
    "cloud.google.com/go/firestore"
    firebase "firebase.google.com/go"
    "google.golang.org/api/iterator"
    "google.golang.org/api/option"
)

// ChannelData は Firestoreに保存されているチャンネル情報の型
type ChannelData struct {
    ChannelName string `firestore:"channelName"`
    Platform    string `firestore:"platform"`
    ThumbnailUrl string `firestore:"thumbnailUrl"`
}

var firestoreClient *firestore.Client

func init() {
    // 環境変数からサービスアカウントキーのパスを取得
    serviceAccountKeyPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY_PATH")
    if serviceAccountKeyPath == "" {
        fmt.Println("エラー: FIREBASE_SERVICE_ACCOUNT_KEY_PATHが設定されていません。サーバーを終了します。")
        os.Exit(1)
    }

    ctx := context.Background()
    conf := &firebase.Config{
        ProjectID: os.Getenv("NEXT_PUBLIC_FIREBASE_PROJECT_ID"),
    }
    
    // サービスアカウントキーを指定してFirebase Admin SDKを初期化
    opt := option.WithCredentialsFile(serviceAccountKeyPath)
    app, err := firebase.NewApp(ctx, conf, opt)
    if err != nil {
        fmt.Printf("エラー: Firebaseアプリの初期化に失敗しました: %v\n", err)
        os.Exit(1)
    }

    // Firestoreクライアントを取得 (変数errを再利用)
    firestoreClient, err = app.Firestore(ctx)
    if err != nil {
        fmt.Printf("エラー: Firestoreクライアントの取得に失敗しました: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Firestoreクライアントの初期化に成功しました。")
}

// GetUserChannels は指定されたユーザーIDのFirestoreからチャンネルリストを取得する
func GetUserChannels(userID string) (map[string]ChannelData, error) {
    if firestoreClient == nil {
        return nil, fmt.Errorf("Firestoreクライアントが初期化されていません")
    }
    
    ctx := context.Background()
    channels := make(map[string]ChannelData)
    
    // Firestoreのクエリを実行
    iter := firestoreClient.Collection("users").Doc(userID).Collection("channels").Documents(ctx)
    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("チャンネルドキュメントの取得に失敗しました: %w", err)
        }
        
        var channelData ChannelData
        if err := doc.DataTo(&channelData); err != nil {
            return nil, fmt.Errorf("チャンネルデータのパースに失敗しました: %w", err)
        }
        channels[doc.Ref.ID] = channelData
    }
    
    return channels, nil
}