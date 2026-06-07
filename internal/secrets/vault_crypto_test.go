package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestParsePrivateKeyPEMInvalid(t *testing.T) {
	_, err := parsePrivateKeyPEM("not-a-pem")
	if err == nil {
		t.Fatal("expected error for invalid pem")
	}
}

func TestDecryptPrivateKeyWithPasswordInvalidJSON(t *testing.T) {
	_, err := decryptPrivateKeyWithPassword([]byte("not-json"), "password")
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestDecryptPrivateKeyWithPasswordWrongPassword(t *testing.T) {
	priv, err := generateRSAKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	pemStr, err := marshalPrivateKeyPEM(priv)
	if err != nil {
		t.Fatal(err)
	}
	enc, err := encryptPrivateKeyWithPassword(pemStr, "correctpass")
	if err != nil {
		t.Fatal(err)
	}
	_, err = decryptPrivateKeyWithPassword(enc, "wrongpass")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestUnwrapDataKeyRSAInvalidBase64(t *testing.T) {
	priv, err := generateRSAKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	_, err = unwrapDataKeyRSA(priv, "not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestEncryptPrivateKeyWithPasswordRoundTrip(t *testing.T) {
	priv, err := generateRSAKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	pemStr, err := marshalPrivateKeyPEM(priv)
	if err != nil {
		t.Fatal(err)
	}
	enc, err := encryptPrivateKeyWithPassword(pemStr, "testpass12")
	if err != nil {
		t.Fatal(err)
	}
	got, err := decryptPrivateKeyWithPassword(enc, "testpass12")
	if err != nil {
		t.Fatal(err)
	}
	if got != pemStr {
		t.Fatal("pem round trip mismatch")
	}
}

func TestWrapDataKeyRSARoundTrip(t *testing.T) {
	priv, err := generateRSAKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	dataKey, err := randomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := wrapDataKeyRSA(&priv.PublicKey, dataKey)
	if err != nil {
		t.Fatal(err)
	}
	unwrapped, err := unwrapDataKeyRSA(priv, wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if string(unwrapped) != string(dataKey) {
		t.Fatal("data key round trip mismatch")
	}
}

func TestUnwrapDataKeyRSAWrongKey(t *testing.T) {
	priv1, err := generateRSAKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	priv2, err := generateRSAKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	dataKey, err := randomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := wrapDataKeyRSA(&priv1.PublicKey, dataKey)
	if err != nil {
		t.Fatal(err)
	}
	_, err = unwrapDataKeyRSA(priv2, wrapped)
	if err == nil {
		t.Fatal("expected error for wrong private key")
	}
}

func TestDecryptSecretsJSONInvalidBase64(t *testing.T) {
	key, err := randomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	_, err = decryptSecretsJSON(key, "not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecryptSecretsJSONInvalidPayload(t *testing.T) {
	key, err := randomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	invalid := base64.StdEncoding.EncodeToString([]byte("not-json"))
	_, err = decryptSecretsJSON(key, invalid)
	if err == nil {
		t.Fatal("expected error for invalid payload json")
	}
}

func TestDecryptSecretsJSONTamperedCiphertext(t *testing.T) {
	key, err := randomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	enc, err := encryptSecretsJSON(key, map[string]string{"a": "b"})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0xff
	tampered := base64.StdEncoding.EncodeToString(raw)
	_, err = decryptSecretsJSON(key, tampered)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
}

func TestDecryptSecretsJSONNullPlaintext(t *testing.T) {
	key, err := randomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	enc, err := encryptSecretsJSON(key, nil)
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := decryptSecretsJSON(key, enc)
	if err != nil {
		t.Fatal(err)
	}
	if secrets != nil {
		t.Fatalf("expected nil secrets map, got %v", secrets)
	}
}

func TestDecryptSecretsJSONInvalidSecretsJSON(t *testing.T) {
	key, err := randomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	block, err := aesNewGCMEncryptRaw(key, []byte("not-an-object"))
	if err != nil {
		t.Fatal(err)
	}
	enc := base64.StdEncoding.EncodeToString(block)
	_, err = decryptSecretsJSON(key, enc)
	if err == nil {
		t.Fatal("expected error for invalid secrets json")
	}
}

func TestEncryptSecretsJSONInvalidKeySize(t *testing.T) {
	_, err := encryptSecretsJSON([]byte("short"), map[string]string{})
	if err == nil {
		t.Fatal("expected error for invalid key size")
	}
}

func TestEncryptDecryptSecretsJSONRoundTrip(t *testing.T) {
	key, err := randomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	enc, err := encryptSecretsJSON(key, map[string]string{"conn": "secret"})
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := decryptSecretsJSON(key, enc)
	if err != nil {
		t.Fatal(err)
	}
	if secrets["conn"] != "secret" {
		t.Fatalf("got %v", secrets)
	}

	encNil, err := encryptSecretsJSON(key, nil)
	if err != nil {
		t.Fatal(err)
	}
	secretsNil, err := decryptSecretsJSON(key, encNil)
	if err != nil {
		t.Fatal(err)
	}
	if secretsNil != nil {
		t.Fatalf("expected nil map, got %v", secretsNil)
	}
}

func TestDecryptSecretsJSONEmptyPlaintext(t *testing.T) {
	key, err := randomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	enc, err := encryptSecretsJSON(key, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := decryptSecretsJSON(key, enc)
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets) != 0 {
		t.Fatalf("expected empty map, got %v", secrets)
	}
}

func aesNewGCMEncryptRaw(dataKey, plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nil, nonce, plain, nil)
	payload := encryptedBlob{Nonce: nonce, Ciphertext: ciphertext}
	return json.Marshal(payload)
}
