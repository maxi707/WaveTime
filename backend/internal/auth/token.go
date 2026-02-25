package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
}

func IssueToken(secret string, userID int64, role string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		Exp:    time.Now().Add(ttl).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	payloadEnc := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(secret, payloadEnc)
	return payloadEnc + "." + sig, nil
}

func ParseToken(secret, token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, fmt.Errorf("invalid token format")
	}

	payloadEnc := parts[0]
	sig := parts[1]
	if !hmac.Equal([]byte(sig), []byte(sign(secret, payloadEnc))) {
		return Claims{}, fmt.Errorf("invalid token signature")
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(payloadEnc)
	if err != nil {
		return Claims{}, fmt.Errorf("invalid token payload")
	}

	var claims Claims
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return Claims{}, fmt.Errorf("invalid token claims")
	}

	if time.Now().Unix() > claims.Exp {
		return Claims{}, fmt.Errorf("token expired")
	}

	return claims, nil
}

func sign(secret, payload string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
