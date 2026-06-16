package auth

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const (
	// 密码字符集
	passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 默认密码长度
	defaultPasswordLength = 12
)

// HashPassword 使用 bcrypt 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码哈希
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

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
