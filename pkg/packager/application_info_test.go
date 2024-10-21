package packager

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	data = []byte(`<ApplicationInfo xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" ToolVersion="1.8.4.0"><FileName>test</FileName><Name>test</Name><UnencryptedContentSize>0</UnencryptedContentSize><SetupFile>test</SetupFile><EncryptionInfo><EncryptionKey>dGVzdA==</EncryptionKey><MacKey>dGVzdA==</MacKey><InitializationVector>dGVzdA==</InitializationVector><Mac>dGVzdA==</Mac><ProfileIdentifier>ProfileVersion1</ProfileIdentifier><FileDigest>dGVzdA==</FileDigest><FileDigestAlgorithm>SHA256</FileDigestAlgorithm></EncryptionInfo></ApplicationInfo>`)
)

func TestApplicationInfoTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationInfoTestSuite))
}

type ApplicationInfoTestSuite struct {
	suite.Suite
}

func (s *ApplicationInfoTestSuite) TestMarshalXML() {
	ai := &ApplicationInfo{
		FileName:               "test",
		Name:                   "test",
		UnencryptedContentSize: 0,
		SetupFile:              "test",
		EncryptionInfo: EncryptionInfo{
			EncryptionKey:        []byte("test"),
			MACKey:               []byte("test"),
			InitializationVector: []byte("test"),
			Mac:                  []byte("test"),
			ProfileIdentifier:    "ProfileVersion1",
			FileDigest:           []byte("test"),
			FileDigestAlgorithm:  "SHA256",
		},
	}

	xmlBytes, err := xml.Marshal(ai)
	s.Require().NoError(err)

	s.Require().Equal(data, xmlBytes)
}

func (s *ApplicationInfoTestSuite) TestUnmarshalXML() {
	var ai ApplicationInfo
	err := xml.Unmarshal(data, &ai)
	s.Require().NoError(err)

	s.Require().Equal("test", ai.FileName)
	s.Require().Equal("test", ai.Name)
	s.Require().Equal(int64(0), ai.UnencryptedContentSize)
	s.Require().Equal("test", ai.SetupFile)
	s.Require().Equal([]byte("test"), ai.EncryptionInfo.EncryptionKey)
	s.Require().Equal([]byte("test"), ai.EncryptionInfo.MACKey)
	s.Require().Equal([]byte("test"), ai.EncryptionInfo.InitializationVector)
	s.Require().Equal([]byte("test"), ai.EncryptionInfo.Mac)
	s.Require().Equal("ProfileVersion1", ai.EncryptionInfo.ProfileIdentifier)
	s.Require().Equal([]byte("test"), ai.EncryptionInfo.FileDigest)
	s.Require().Equal("SHA256", ai.EncryptionInfo.FileDigestAlgorithm)
}
