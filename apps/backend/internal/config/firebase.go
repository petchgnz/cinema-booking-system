package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

func InitFirebase(credentialsFile string) *auth.Client {
	opt := option.WithCredentialsFile(credentialsFile)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth: %v", err)
	}

	log.Println("Connected to Firebase")
	return authClient
}
