package tool

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/pkg/hash"
	"github.com/OmnTeam/ppanel-pro/pkg/jwt"
	"github.com/OmnTeam/ppanel-pro/pkg/random"
	"github.com/OmnTeam/ppanel-pro/pkg/snowflake"
)

func MicrosecondsStr(elapsed time.Duration) string {
	return fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)
}

// KeyNew generates a random key
func KeyNew(length int, keyType int) string {
	return random.KeyNew(length, keyType)
}

// EncodeBase36 encodes an int64 to base36 string
func EncodeBase36(id int64) string {
	return random.EncodeBase36(id)
}

// StrToDashedString converts a string number to dashed format
func StrToDashedString(strNum string) string {
	return random.StrToDashedString(strNum)
}

// ParseJWT parses a JWT token with the given secret
func ParseJWT(tokenString string, secret string) (map[string]interface{}, error) {
	claims, err := jwt.ParseJwtToken(tokenString, secret)
	if err != nil {
		return nil, err
	}
	// Convert jwt.MapClaims to map[string]interface{}
	result := make(map[string]interface{})
	for k, v := range claims {
		result[k] = v
	}
	return result, nil
}

// GenerateInviteCode generates an invite code from user ID
func GenerateInviteCode(userID int64) string {
	// Generate invite code using Base36 encoding of snowflake ID
	id := snowflake.GetID()
	code := random.EncodeBase36(id)
	return random.StrToDashedString(code)
}

// GenerateReferCode generates a refer code from user ID (alias for GenerateInviteCode)
func GenerateReferCode(userID int64) string {
	return GenerateInviteCode(userID)
}

// GenerateSubscribeToken generates a subscription token from order number
func GenerateSubscribeToken(orderNo string) string {
	// Generate a token based on order number
	data := fmt.Sprintf("%s-%d", orderNo, time.Now().UnixNano())
	return hash.Md5Hex([]byte(data))
}

// StructToJSON converts a struct to JSON string
func StructToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// JSONToStruct converts JSON string to struct
func JSONToStruct(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// StringToStringSlice splits a comma-separated string to slice
func StringToStringSlice(str string) []string {
	if str == "" {
		return []string{}
	}
	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GenerateRandomString generates a random string with given length
func GenerateRandomString(length int) string {
	return random.KeyNew(length, 0)
}

// GenerateUUID generates a UUID v4 string
func GenerateUUID() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		randBytes(4),
		randBytes(2),
		randBytes(2),
		randBytes(2),
		randBytes(6))
}

// GenerateJWT generates a JWT token
func GenerateJWT(secret string, expireSeconds int64, claims map[string]interface{}) (string, error) {
	opts := make([]jwt.Option, 0, len(claims))
	for k, v := range claims {
		opts = append(opts, jwt.WithOption(k, v))
	}
	return jwt.NewJwtToken(secret, time.Now().Unix(), expireSeconds, opts...)
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
