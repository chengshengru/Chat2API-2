package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"
	"runtime"
	"strings"
)

var (
	// ErrInvalidKeySize 密钥长度错误
	ErrInvalidKeySize = errors.New("invalid key size: must be 16, 24, or 32 bytes")

	// ErrCipherTextShort 密文太短
	ErrCipherTextShort = errors.New("ciphertext too short")

	// ErrDecryptionFailed 解密失败
	ErrDecryptionFailed = errors.New("decryption failed")
)

// deriveEncryptionKey 从机器信息派生加密密钥
// 与 Electron 版保持一致的密钥派生逻辑
func deriveEncryptionKey() []byte {
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}

	// 与 Electron 版保持兼容的密钥派生
	machineData := hostname + username + runtime.GOOS + runtime.GOARCH
	hash := sha256.Sum256([]byte(machineData))
	return hash[:]
}

// EncryptData 使用 AES-GCM 加密数据
func EncryptData(plaintext string) string {
	if plaintext == "" {
		return ""
	}

	key := deriveEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return plaintext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return plaintext
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// 将 nonce + ciphertext 合并为 base64 编码
	result := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(result)
}

// DecryptData 使用 AES-GCM 解密数据
func DecryptData(encrypted string) string {
	if encrypted == "" {
		return ""
	}

	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return encrypted
	}

	key := deriveEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return encrypted
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return encrypted
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return encrypted
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return encrypted
	}

	return string(plaintext)
}

// EncryptCredentials 加密凭证映射
func EncryptCredentials(credentials map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range credentials {
		result[key] = EncryptData(value)
	}
	return result
}

// DecryptCredentials 解密凭证映射
func DecryptCredentials(credentials map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range credentials {
		result[key] = DecryptData(value)
	}
	return result
}

// GenerateId 生成唯一 ID
func GenerateId() string {
	// 使用随机数生成器创建 ID
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return strings.Replace(base64.URLEncoding.EncodeToString(b)[:16], "-", "", -1)
}
