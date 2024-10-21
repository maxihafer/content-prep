package zipper

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestZipperTestSuite(t *testing.T) {
	suite.Run(t, new(ZipperTestSuite))
}

type ZipperTestSuite struct {
	suite.Suite

	testDir string

	srcDir   string
	destFile *os.File
}

func (s *ZipperTestSuite) SetupTest() {
	var err error

	s.testDir, err = os.MkdirTemp("", "zipper-test-*")
	s.Require().NoError(err)

	s.srcDir = path.Join(s.testDir, "src")

	err = os.Mkdir(s.srcDir, 0755)
	s.Require().NoError(err)

	srcFile, err := os.Create(path.Join(s.srcDir, "test"))
	s.Require().NoError(err)
	defer srcFile.Close()

	_, err = srcFile.Write([]byte("Hello, World!"))
	s.Require().NoError(err)

	err = os.MkdirAll(path.Join(s.srcDir, "subdir"), 0755)
	s.Require().NoError(err)

	srcFile2, err := os.Create(path.Join(s.srcDir, "subdir", "test2"))
	s.Require().NoError(err)
	defer srcFile2.Close()

	_, err = srcFile2.Write([]byte("Hello, World 2!"))
	s.Require().NoError(err)

	s.destFile, err = os.Create(path.Join(s.testDir, "src.zip"))
	s.Require().NoError(err)

	s.T().Log("Running tests in: " + s.testDir)
}

func (s *ZipperTestSuite) TearDownTest() {
	s.Require().NoError(os.RemoveAll(s.testDir))
}

func (s *ZipperTestSuite) TestZip() {
	fs := os.DirFS(s.srcDir)

	err := Zip(fs, s.destFile)
	s.Require().NoError(err)
}

func (s *ZipperTestSuite) TestUnzip() {
	fs := os.DirFS(s.srcDir)

	err := Zip(fs, s.destFile)
	s.Require().NoError(err)

	err = Unzip(s.destFile, path.Join(s.testDir, "unzip"))
	s.Require().NoError(err)

	resultFile, err := os.Open(path.Join(s.testDir, "unzip", "test"))
	s.Require().NoError(err)
	defer resultFile.Close()

	buf := make([]byte, 13)
	n, err := resultFile.Read(buf)
	s.Require().NoError(err)
	s.Require().Equal(13, n)
	s.Require().Equal("Hello, World!", string(buf))

	resultFile2, err := os.Open(path.Join(s.testDir, "unzip", "subdir", "test2"))
	s.Require().NoError(err)
	defer resultFile2.Close()

	buf2 := make([]byte, 15)
	n, err = resultFile2.Read(buf2)
	s.Require().NoError(err)
	s.Require().Equal(15, n)
	s.Require().Equal("Hello, World 2!", string(buf2))

}
