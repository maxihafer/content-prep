package cryptostream

import (
	"crypto/sha256"
	"io"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type mywriter struct {
	buf []byte
	pos int
}

func (m *mywriter) Write(p []byte) (n int, err error) {
	minCap := m.pos + len(p)
	if minCap > cap(m.buf) { // Make sure buf has enough capacity:
		buf2 := make([]byte, len(m.buf), minCap+len(p)) // add some extra
		copy(buf2, m.buf)
		m.buf = buf2
	}
	if minCap > len(m.buf) {
		m.buf = m.buf[:minCap]
	}
	copy(m.buf[m.pos:], p)
	m.pos += len(p)
	return len(p), nil
}

func (m *mywriter) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.buf) {
		return 0, io.EOF
	}
	n = copy(p, m.buf[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mywriter) Seek(offset int64, whence int) (int64, error) {
	newPos, offs := 0, int(offset)
	switch whence {
	case io.SeekStart:
		newPos = offs
	case io.SeekCurrent:
		newPos = m.pos + offs
	case io.SeekEnd:
		newPos = len(m.buf) + offs
	}
	if newPos < 0 {
		return 0, errors.New("negative result pos")
	}
	m.pos = newPos
	return int64(newPos), nil
}

func TestAESStreamTestSuite(t *testing.T) {
	suite.Run(t, new(AESStreamTestSuite))
}

type AESStreamTestSuite struct {
	suite.Suite

	aesKey  []byte
	iv      []byte
	hmacKey []byte
}

func (s *AESStreamTestSuite) SetupSuite() {
	s.aesKey = []byte(strings.Repeat("a", 32))

	s.iv = []byte(strings.Repeat("i", 16))

	s.hmacKey = []byte(strings.Repeat("h", 32))
}

func (s *AESStreamTestSuite) TestEncrypt() {
	plaintext := strings.NewReader("test")

	ciphertext := &mywriter{}

	err := Encrypt(plaintext, ciphertext, s.aesKey, s.iv, s.hmacKey)
	s.Require().NoError(err)

	_, err = ciphertext.Seek(0, io.SeekStart)
	s.Require().NoError(err)

	s.Require().Equal(s.iv, ciphertext.buf[sha256.Size:sha256.Size+len(s.iv)])

	decryptedPlaintext := &mywriter{}

	err = Decrypt(ciphertext, decryptedPlaintext, s.aesKey, s.hmacKey)
	s.Require().NoError(err)

	s.Require().Equal("test", string(decryptedPlaintext.buf))
}
