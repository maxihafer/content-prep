package cryptostream

import (
	"crypto/aes"
	"crypto/cipher"
	HMAC "crypto/hmac"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/sirupsen/logrus"
)

const BufferSize int = 2097152
const IvSize int = 16
const HMACKeySize = 64

type loggingWriter struct {
	w    io.Writer
	mode string
}

func (l loggingWriter) Write(p []byte) (n int, err error) {
	logrus.WithField("mode", l.mode).Info("Write:", p)

	return l.w.Write(p)
}

// Encrypt the stream using the given AES-CTR and SHA256-HMAC key
func Encrypt(in io.ReadSeeker, out io.WriteSeeker, keyAes []byte, iv []byte, hmacKey []byte) error {
	AES, err := aes.NewCipher(keyAes)
	if err != nil {
		return err
	}

	switch len(keyAes) {
	case 16, 24, 32:
		break
	default:
		return errors.New("invalid AES key length, expected 16, 24 or 32 bytes")
	}

	if len(iv) != IvSize {
		return errors.New("invalid IV length, expected 16 bytes")
	}

	if len(hmacKey) != HMACKeySize {
		return errors.New("invalid HMAC key length, expected 64 bytes")
	}

	hasher := HMAC.New(sha256.New, hmacKey)
	ctr := cipher.NewCTR(AES, iv)

	_, err = out.Seek(sha256.Size, io.SeekStart)
	if err != nil {
		return err
	}

	w := io.MultiWriter(out, hasher)

	_, err = w.Write(iv)
	if err != nil {
		return err
	}

	var n int
	buf := make([]byte, BufferSize)
	for {
		n, err = in.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if n > 0 {
			outBuf := make([]byte, n)
			ctr.XORKeyStream(outBuf, buf[:n])
			_, err = w.Write(outBuf)
			if err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		}
	}

	_, err = out.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = out.Write(hasher.Sum(nil))
	if err != nil {
		return err
	}

	return nil
}

// Decrypt the stream and verify HMAC using the given AES-CTR and SHA512-HMAC key
// Do not trust the out io.Writer contents until the function returns the result
// of validating the ending HMAC hash.
func Decrypt(in io.Reader, out io.Writer, keyAes []byte, hmacKey []byte) error {
	hash := make([]byte, sha256.Size)

	_, err := io.ReadFull(in, hash)
	if err != nil {
		return err
	}

	iv := make([]byte, IvSize)
	_, err = io.ReadFull(in, iv)
	if err != nil {
		return err
	}

	AES, err := aes.NewCipher(keyAes)
	if err != nil {
		return err
	}

	ctr := cipher.NewCTR(AES, iv)
	hasher := HMAC.New(sha256.New, hmacKey)

	_, err = hasher.Write(iv)
	if err != nil {
		return err
	}

	var n int
	buf := make([]byte, BufferSize)
	for {
		n, err = in.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if n > 0 {
			outBuf := make([]byte, n)

			_, err := hasher.Write(buf[:n])
			if err != nil {
				return err
			}

			ctr.XORKeyStream(outBuf, buf[:n])
			_, err = out.Write(outBuf)
			if err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		}
	}

	if !HMAC.Equal(hasher.Sum(nil), hash) {
		return errors.New("HMAC mismatch")
	}

	return nil
}
