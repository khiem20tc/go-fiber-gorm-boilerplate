package key

import (
	"crypto/rand"
	"math/big"
	"sync"
)

const (
	// Define the character sets for API key and API secret.
	apiKeyCharset    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	apiSecretCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%&?"
	apiKeyLength     = 32  // Adjust the length as needed.
	apiSecretLength  = 128 // Adjust the length as needed.
)

func GenerateAPIKey() (string, string) {
	apiKey, secretKey := "", ""
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		apiKey = generateRandomString(apiKeyCharset, apiKeyLength)
	}()
	go func() {
		defer wg.Done()
		secretKey = generateRandomString(apiSecretCharset, apiSecretLength)
	}()
	wg.Wait()
	return apiKey, secretKey
}

func generateRandomString(charset string, length int) string {
	// Calculate the maximum index in the charset.
	maxIndex := big.NewInt(int64(len(charset)))

	// Generate random indices and create the string.
	var result string
	for i := 0; i < length; i++ {
		index, _ := rand.Int(rand.Reader, maxIndex)

		result += string(charset[index.Int64()])
	}

	return result
}
