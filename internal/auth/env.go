package auth

import "os"

// getenv is a wrapper so envOrEmpty in oauth.go can be stubbed in tests
// without leaking os.Getenv into the test surface.
func getenv(key string) string { return os.Getenv(key) }
