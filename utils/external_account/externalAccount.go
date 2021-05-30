package external_account

import "github.com/lakshay35/finlit-backend/utils/encryption"

// ConvertToAccessToken...
func ConvertToAccessToken(token string) string {
	b64d, b64dError := encryption.DecodeBase64(token)

	if b64dError != nil {
		panic(b64dError)
	}

	return string(encryption.Decrypt([]byte(b64d)))
}
