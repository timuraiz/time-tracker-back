package supabase

import (
	"log"
	"os"
)

type Client struct {
	URL        string
	AnonKey    string
	ServiceKey string
}

var SupabaseClient *Client

func InitClient() {
	url := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if url == "" || anonKey == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_ANON_KEY environment variables are required")
	}

	SupabaseClient = &Client{
		URL:        url,
		AnonKey:    anonKey,
		ServiceKey: serviceKey,
	}

	log.Println("Supabase client initialized successfully")
}

func GetClient() *Client {
	if SupabaseClient == nil {
		InitClient()
	}
	return SupabaseClient
}
