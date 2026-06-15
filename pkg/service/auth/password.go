package auth

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

const (
	// 密码字符集
	passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 默认密码长度
	defaultPasswordLength = 12
)

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int) (string, error) {
	if length <= 0 {
		length = defaultPasswordLength
	}

	password := make([]byte, length)
	charsLen := big.NewInt(int64(len(passwordChars)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsLen)
		if err != nil {
			return "", err
		}
		password[i] = passwordChars[num.Int64()]
	}

	return string(password), nil
}

// GenerateSecretKey 生成 JWT 密钥
func GenerateSecretKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
