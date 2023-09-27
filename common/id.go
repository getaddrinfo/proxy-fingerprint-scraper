package common

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateID(prefix string) string {
	bytes := make([]byte, 16)
	crand.Read(bytes)
	return fmt.Sprintf("%s_%s", prefix, base64.RawURLEncoding.EncodeToString(bytes))
}
