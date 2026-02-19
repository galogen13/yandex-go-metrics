package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
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

	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, e.publicKey, data, nil)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	return ciphertext, nil
}

// Decrypt расшифровывает данные с использованием приватного ключа
func (d *Decryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	if d == nil || d.privateKey == nil {
		// Если дешифратор не инициализирован, возвращаем данные как есть
		return ciphertext, nil
	}

	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, d.privateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// loadPublicKey загружает публичный ключ из PEM файла
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

// loadPrivateKey загружает приватный ключ из PEM файла
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Пробуем PKCS1
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	rsaPriv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaPriv, nil
}

// GenerateKeys генерирует пару ключей RSA
func GenerateKeys(bits int) (privateKeyPEM, publicKeyPEM []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	// Приватный ключ в PKCS8
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}
	privateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Публичный ключ
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	publicKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return privateKeyPEM, publicKeyPEM, nil
}
