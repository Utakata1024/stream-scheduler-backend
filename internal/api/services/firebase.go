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

type ChannelData struct {
	ChannelName string `firestore:"channelName"`
	Platform    string `firestore:"platform"`
	ThumbnailUrl string `firestore:"thumbnailUrl"`
}

var firestoreClient *firestoreClient

func init() {
	// 環境変数からサービスアカウントキーのパスを取得
	serviceAccountKeyPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY_PATH")
	if serviceAccountKeyPath == "" {
		fmt.Println("FIREBASE_SERVICE_ACCOUNT_KEY_PATH is not set")
		return
	}

	ctx := context.Background()
	conf := &firebase.Config{
		ProjectID: os.Getenv("NEXT_PUBLIC_FIREBASE_PROJECT_ID"),
	}

	// サービスアカウントキーを指定してFirebase Admin SDK初期化
	opt := option.WithCredentialsFile(serviceAccountKeyPath)
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		fmt.Printf("Firebaseアプリの初期化に失敗しました: %v\n", err)
		return
	}

	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		fmt.Printf("Firestoreクライアントの初期化に失敗しました: %v\n", err)
		return
	}
	fmt.Println("Firestoreクライアントの初期化に成功しました")
}

// GetUserChannelsは、指定されたユーザーIDに関連付けられたチャンネル情報をFirestoreから取得します。
func GetUserChannels(userID string) (map[string]ChannelData, error) {
	if firestoreClient == nil {
		return nil, fmt.Errorf("Firestoreクライアントが初期化されていません")
	}

	ctx := context.Background()
	channels := make(map[string]ChannelData)

	iter := firestoreClient.Collection("users").Doc(userID).Collection("channels").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("チャンネル情報の取得中にエラーが発生しました: %v", err)
		}

		var channelData ChannelData
		if err := doc.DataTo(&channelData); err != nil {
			return nil, fmt.Errorf("ドキュメントデータの変換中にエラーが発生しました: %v", err)
		}
		channels[doc.Ref.ID] = channelData
	}

	return channels, nil
}