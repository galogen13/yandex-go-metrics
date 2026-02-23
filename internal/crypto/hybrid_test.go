package crypto_test

import (
	"os"
	"testing"

	"github.com/galogen13/yandex-go-metrics/internal/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {

	privatePEM, publicPEM, err := crypto.GenerateKeys()
	assert.NoError(t, err)

	tmpPrivate := t.TempDir() + "/private.pem"
	tmpPublic := t.TempDir() + "/public.pem"

	err = os.WriteFile(tmpPrivate, []byte(privatePEM), 0600)
	assert.NoError(t, err)
	defer os.Remove(tmpPrivate)

	err = os.WriteFile(tmpPublic, []byte(publicPEM), 0644)
	assert.NoError(t, err)
	defer os.Remove(tmpPublic)

	encryptor, err := crypto.NewEncryptor(tmpPublic)
	assert.NoError(t, err)
	require.NotNil(t, encryptor)

	decryptor, err := crypto.NewDecryptor(tmpPrivate)
	assert.NoError(t, err)
	require.NotNil(t, decryptor)

	original := []byte("Hello, World! Ооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооочень длинное сообщение!№;%:?*()Х[]{}<>")

	ciphertext, err := encryptor.Encrypt(original)
	assert.NoError(t, err)
	assert.NotEmpty(t, ciphertext)

	plaintext, err := decryptor.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.NotEmpty(t, plaintext)

	assert.Equal(t, original, plaintext)

}

func TestNoKey(t *testing.T) {

	encryptor, err := crypto.NewEncryptor("")
	assert.NoError(t, err)
	require.Nil(t, encryptor)

	decryptor, err := crypto.NewDecryptor("")
	assert.NoError(t, err)
	require.Nil(t, decryptor)

	original := []byte("Hello, World! Ооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооооочень длинное сообщение!№;%:?*()Х[]{}<>")

	ciphertext, err := encryptor.Encrypt(original)
	assert.NoError(t, err)
	assert.NotEmpty(t, ciphertext)

	assert.Equal(t, original, ciphertext)

	plaintext, err := decryptor.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.NotEmpty(t, plaintext)

	assert.Equal(t, original, plaintext)
}

func TestWrongPath(t *testing.T) {
	tmpPrivate := t.TempDir() + "/private.pem"
	tmpPublic := t.TempDir() + "/public.pem"

	encryptor, err := crypto.NewEncryptor(tmpPublic)
	assert.Error(t, err)
	require.Nil(t, encryptor)

	decryptor, err := crypto.NewDecryptor(tmpPrivate)
	assert.Error(t, err)
	require.Nil(t, decryptor)
}
