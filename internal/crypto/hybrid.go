// Изначально в проекте использовалось rsa-шифрование, но размер сообщения превышал
// длину ключа даже с ключем 4096 бит, поэтому пришлось использовать гибридный метод.
// Для выполнения задачи была использована готовая библиотека.
package crypto

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/threatflux/cryptum-go/pkg/encryption"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"go.uber.org/zap"
)

// Encryptor для шифрования данных публичным ключом
type Encryptor struct {
	publicKey *rsa.PublicKey
}

// Decryptor для расшифровки данных приватным ключом
type Decryptor struct {
	privateKey *rsa.PrivateKey
}

// NewEncryptor создает новый шифратор из публичного ключа
func NewEncryptor(publicKeyPath string) (*Encryptor, error) {
	if publicKeyPath == "" {
		return nil, nil // Шифрование отключено
	}

	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &Encryptor{publicKey: publicKey}, nil
}

// NewDecryptor создает новый дешифратор из приватного ключа
func NewDecryptor(privateKeyPath string) (*Decryptor, error) {
	if privateKeyPath == "" {
		return nil, nil // Дешифрование отключено
	}

	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return &Decryptor{privateKey: privateKey}, nil
}

// Encrypt шифрует данные с использованием публичного ключа
func (e *Encryptor) Encrypt(data []byte) ([]byte, error) {
	if e == nil || e.publicKey == nil {
		// Если шифратор не инициализирован, возвращаем данные как есть
		return data, nil
	}

	return encryption.EncryptBlob(data, e.publicKey)
}

// Decrypt расшифровывает данные с использованием приватного ключа
func (d *Decryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	if d == nil || d.privateKey == nil {
		// Если дешифратор не инициализирован, возвращаем данные как есть
		return ciphertext, nil
	}

	return encryption.DecryptBlob(ciphertext, d.privateKey)
}

// loadPublicKey загружает публичный ключ из PEM файла
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return encryption.ParsePublicKey(string(pemData))

}

// loadPrivateKey загружает приватный ключ из PEM файла
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return encryption.ParsePrivateKey(string(pemData))

}

// GenerateKeys генерирует пару ключей RSA
func GenerateKeys() (privateKeyPEM, publicKeyPEM string, err error) {

	return encryption.GenerateKeyPair()
}

func DecryptMiddleware(d *Decryptor, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if d == nil {
			next.ServeHTTP(w, r)
			return
		}

		ciphertext, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error("unexpected error reading request body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body.Close()

		plaintext, err := d.Decrypt(ciphertext)
		if err != nil {
			logger.Log.Error("Decryption failed", zap.Error(err))
			http.Error(w, "Decryption failed", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(plaintext))
		r.ContentLength = int64(len(plaintext))

		next.ServeHTTP(w, r)
	}
}
