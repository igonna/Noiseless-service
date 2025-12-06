package main

import (
	"errors"
	"os"
	"strings"
	"sync"

	supabase "github.com/supabase-community/supabase-go"
)

var (
	supabaseOnce sync.Once
	supabaseInst *supabase.Client
	supabaseErr  error
)

func getSupabaseClient() (*supabase.Client, error) {
	supabaseOnce.Do(func() {
		url := strings.TrimSpace(os.Getenv("SUPABASE_URL"))
		key := strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))
		if url == "" {
			supabaseErr = errors.New("SUPABASE_URL must be set")
			return
		}
		if key == "" {
			supabaseErr = errors.New("SUPABASE_SERVICE_ROLE_KEY must be set")
			return
		}
		supabaseInst, supabaseErr = supabase.NewClient(url, key, nil)
	})
	return supabaseInst, supabaseErr
}
