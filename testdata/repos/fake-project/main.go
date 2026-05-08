package main

// This is a fake source file with intentionally planted secrets
// for testing the VaultGuard scanner. NONE of these are real.

const (
	AppName = "my-app"
	Version = "1.0.0"

	// These should be detected:
	APIKey    = "EXAMPLE_sk_test_51HG3kJKlmnopQRSTuvwxyz1234567890abcdef"
	SecretKey = "EXAMPLE_super_secret_value_that_should_not_be_here_1234"
)

func getConfig() map[string]string {
	return map[string]string{
		"api_key": "EXAMPLE_AKIAIOSFODNN7EXAMPLE",
		"region":  "us-east-1", // safe
	}
}
