package packager

import (
	"content-prep/pkg/cryptostream"
	"content-prep/pkg/zipper"
	"context"
	"encoding/xml"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

var _ KeyGenerator = &mykeygen{}

type mykeygen struct {
}

func (m mykeygen) GenerateKey(length int) ([]byte, error) {
	return []byte(strings.Repeat(".", length)), nil
}

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

func TestPackagerTestSuite(t *testing.T) {
	suite.Run(t, new(PackagerTestSuite))
}

type PackagerTestSuite struct {
	suite.Suite

	testDir string

	fs fs.FS
}

func (s *PackagerTestSuite) SetupSuite() {
	s.fs = fstest.MapFS{
		"test.exe": {
			Data: []byte("test"),
		},
	}
	var err error
	s.testDir, err = os.MkdirTemp("", "packager-test-*")
	s.Require().NoError(err)

	s.T().Log("Running tests in: " + s.testDir)
}

func (s *PackagerTestSuite) TearDownSuite() {
	s.Require().NoError(os.RemoveAll(s.testDir))
}

func (s *PackagerTestSuite) TestPackager() {
	p := &packager{
		keygen: &mykeygen{},
	}

	out, err := os.Create(path.Join(s.testDir, "test.intunewin"))

	err = p.CreatePackage(context.Background(), s.fs, "test.exe", out)
	s.Require().NoError(err)

	err = zipper.Unzip(out, path.Join(s.testDir, "test.unzip"))
	s.Require().NoError(err)

	detection, err := os.Open(path.Join(s.testDir, "test.unzip", "IntuneWinPackage", "Metadata", "Detection.xml"))
	s.Require().NoError(err)

	detectionData, err := io.ReadAll(detection)
	s.Require().NoError(err)

	ai := &ApplicationInfo{}
	err = xml.Unmarshal(detectionData, ai)
	s.Require().NoError(err)

	s.Require().Equal("test", ai.Name)
	s.Require().Equal("IntunePackage.intunewin", ai.FileName)
	s.Require().Equal("test.exe", ai.SetupFile)
	s.Require().Less(int64(0), ai.UnencryptedContentSize)

	kg := mykeygen{}

	aesKey, err := kg.GenerateKey(32)
	s.Require().NoError(err)

	iv, err := kg.GenerateKey(cryptostream.IvSize)
	s.Require().NoError(err)

	hmacKey, err := kg.GenerateKey(cryptostream.HMACKeySize)
	s.Require().NoError(err)

	s.Require().Equal(aesKey, ai.EncryptionInfo.EncryptionKey)
	s.Require().Equal(hmacKey, ai.EncryptionInfo.MACKey)
	s.Require().Equal(iv, ai.EncryptionInfo.InitializationVector)
	s.Require().Equal("ProfileVersion1", ai.EncryptionInfo.ProfileIdentifier)
	s.Require().Equal("SHA256", ai.EncryptionInfo.FileDigestAlgorithm)

	err = p.DecryptPackage(context.Background(), out, path.Join(s.testDir, "test.decrypted"))
	s.Require().NoError(err)

	decryptedArchive, err := os.Open(path.Join(s.testDir, "test.decrypted", "IntuneWinPackage", "Contents", "IntunePackage.intunewin.zip"))
	s.Require().NoError(err)

	err = zipper.Unzip(decryptedArchive, path.Join(s.testDir, "test.decrypted", "unzipped"))
	s.Require().NoError(err)

}