package packager

import "crypto/rand"

var _ KeyGenerator = &defaultKeyGenerator{}

type defaultKeyGenerator struct{}

func (g defaultKeyGenerator) GenerateKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}
