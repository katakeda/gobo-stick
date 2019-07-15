package main

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthClient ...
type OAuthClient struct {
	config *oauth2.Config
}

func (auth *OAuthClient) init() {
	auth.config = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

var auth OAuthClient
