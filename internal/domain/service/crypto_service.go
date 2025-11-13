package service

type CryptoService interface {
	GenerateAESKey() ([]byte, error)
	EncryptAESGCM(key, plaintext []byte) ([]byte, []byte, error)
	DecryptAESGCM(key, nonce, ciphertext []byte) ([]byte, error)
}

