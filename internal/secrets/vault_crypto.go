package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

func generateRSAKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, rsaKeyBits)
}

func marshalPublicKeyPEM(pub *rsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	block := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: der}
	return string(pem.EncodeToMemory(block)), nil
}

func marshalPrivateKeyPEM(priv *rsa.PrivateKey) (string, error) {
	der := x509.MarshalPKCS1PrivateKey(priv)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}
	return string(pem.EncodeToMemory(block)), nil
}

func parsePrivateKeyPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid private key pem")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func encryptPrivateKeyWithPassword(pemStr, password string) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, aesKeySize, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(pemStr), nil)
	payload := encryptedBlob{Salt: salt, Nonce: nonce, Ciphertext: ciphertext}
	return json.Marshal(payload)
}

func decryptPrivateKeyWithPassword(data []byte, password string) (string, error) {
	var payload encryptedBlob
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	key := pbkdf2.Key([]byte(password), payload.Salt, pbkdf2Iterations, aesKeySize, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, payload.Nonce, payload.Ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

type encryptedBlob struct {
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

func wrapDataKeyRSA(pub *rsa.PublicKey, dataKey []byte) (string, error) {
	wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, dataKey, nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(wrapped), nil
}

func unwrapDataKeyRSA(priv *rsa.PrivateKey, wrappedB64 string) ([]byte, error) {
	wrapped, err := base64.StdEncoding.DecodeString(wrappedB64)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, wrapped, nil)
}

func encryptSecretsJSON(dataKey []byte, secrets map[string]string) (string, error) {
	raw, err := json.Marshal(secrets)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, raw, nil)
	payload := encryptedBlob{Nonce: nonce, Ciphertext: ciphertext}
	out, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(out), nil
}

func decryptSecretsJSON(dataKey []byte, encryptedB64 string) (map[string]string, error) {
	raw, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return nil, err
	}
	var payload encryptedBlob
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, payload.Nonce, payload.Ciphertext, nil)
	if err != nil {
		return nil, err
	}
	secrets := map[string]string{}
	if len(plain) == 0 {
		return secrets, nil
	}
	if err := json.Unmarshal(plain, &secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}

func randomAESKey() ([]byte, error) {
	key := make([]byte, aesKeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("random aes key: %w", err)
	}
	return key, nil
}
