package config

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// func InitFirebase(credentialsFile string) *auth.Client {
// 	opt := option.WithCredentialsFile(credentialsFile)

// 	app, err := firebase.NewApp(context.Background(), nil, opt)
// 	if err != nil {
// 		log.Fatalf("Failed to initialize Firebase app: %v", err)
// 	}

// 	authClient, err := app.Auth(context.Background())
// 	if err != nil {
// 		log.Fatalf("Failed to initialize Firebase Auth: %v", err)
// 	}

// 	log.Println("Connected to Firebase")
// 	return authClient
// }

func InitFirebase(credentialsFile string) *auth.Client {
	ctx := context.Background()
	var opt option.ClientOption

	// use docker's credFile if FIREBASE_CREDENTIALS_JSON is exist, use local if it's not
	if credJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON"); credJSON != "" {
		opt = option.WithCredentialsJSON([]byte(credJSON))
	} else {
		opt = option.WithCredentialsFile(credentialsFile)
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Failed to get Firebase Auth client: %v", err)
	}

	return client
}
