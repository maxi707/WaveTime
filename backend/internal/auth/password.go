package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const (
	hashIterations = 120000
	saltSize       = 16
)

func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 chars")
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	h := derive([]byte(password), salt, hashIterations)
	return fmt.Sprintf("sha256$%d$%s$%s", hashIterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(h)), nil
}

func VerifyPassword(encodedHash, password string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[0] != "sha256" {
		return false
	}

	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	actual := derive([]byte(password), salt, iter)
	if len(actual) != len(expected) {
		return false
	}

	var diff byte
	for i := range actual {
		diff |= actual[i] ^ expected[i]
	}
	return diff == 0
}

func derive(password, salt []byte, iter int) []byte {
	buf := make([]byte, 0, len(password)+len(salt))
	buf = append(buf, password...)
	buf = append(buf, salt...)
	sum := sha256.Sum256(buf)
	out := sum[:]
	for i := 1; i < iter; i++ {
		s := sha256.Sum256(out)
		out = s[:]
	}
	res := make([]byte, len(out))
	copy(res, out)
	return res
}
