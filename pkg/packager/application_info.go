package packager

import (
	"encoding/base64"
	"encoding/xml"
)

const (
	ToolVersion  = "1.8.4.0"
	XSDNamespace = "http://www.w3.org/2001/XMLSchema"
	XSINamespace = "http://www.w3.org/2001/XMLSchema-instance"
)

var _ xml.Marshaler = &ApplicationInfo{}
var _ xml.Unmarshaler = &ApplicationInfo{}

type ApplicationInfo struct {
	FileName               string         `xml:"FileName"`
	Name                   string         `xml:"Name"`
	UnencryptedContentSize int64          `xml:"UnencryptedContentSize"`
	SetupFile              string         `xml:"SetupFile"`
	EncryptionInfo         EncryptionInfo `xml:"EncryptionInfo"`
}

func (a *ApplicationInfo) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type Alias ApplicationInfo
	aux := &struct {
		XSI         string `xml:"xmlns:xsi,attr"`
		XSD         string `xml:"xmlns:xsd,attr"`
		ToolVersion string `xml:"ToolVersion,attr"`
		Alias
	}{
		ToolVersion: ToolVersion,
		XSI:         XSINamespace,
		XSD:         XSDNamespace,
		Alias:       (Alias)(*a),
	}

	return e.EncodeElement(aux, start)
}

func (a *ApplicationInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias ApplicationInfo
	aux := &struct {
		XSI         string `xml:"xmlns:xsi,attr"`
		XSD         string `xml:"xmlns:xsd,attr"`
		ToolVersion string `xml:"ToolVersion,attr"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := d.DecodeElement(aux, &start); err != nil {
		return err
	}

	a.FileName = aux.FileName
	a.Name = aux.Name
	a.UnencryptedContentSize = aux.UnencryptedContentSize
	a.SetupFile = aux.SetupFile
	a.EncryptionInfo = aux.EncryptionInfo

	return nil
}

var _ xml.Marshaler = &EncryptionInfo{}
var _ xml.Unmarshaler = &EncryptionInfo{}

type EncryptionInfo struct {
	EncryptionKey        []byte `xml:"EncryptionKey"`
	MACKey               []byte `xml:"MacKey"`
	InitializationVector []byte `xml:"InitializationVector"`
	Mac                  []byte `xml:"Mac"`
	ProfileIdentifier    string `xml:"ProfileIdentifier"`
	FileDigest           []byte `xml:"FileDigest"`
	FileDigestAlgorithm  string `xml:"FileDigestAlgorithm"`
}

func (i *EncryptionInfo) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	aux := struct {
		EncryptionKey        string `xml:"EncryptionKey"`
		MACKey               string `xml:"MacKey"`
		InitializationVector string `xml:"InitializationVector"`
		Mac                  string `xml:"Mac"`
		ProfileIdentifier    string `xml:"ProfileIdentifier"`
		FileDigest           string `xml:"FileDigest"`
		FileDigestAlgorithm  string `xml:"FileDigestAlgorithm"`
	}{
		EncryptionKey:        base64.StdEncoding.EncodeToString(i.EncryptionKey),
		MACKey:               base64.StdEncoding.EncodeToString(i.MACKey),
		InitializationVector: base64.StdEncoding.EncodeToString(i.InitializationVector),
		Mac:                  base64.StdEncoding.EncodeToString(i.Mac),
		ProfileIdentifier:    i.ProfileIdentifier,
		FileDigest:           base64.StdEncoding.EncodeToString(i.FileDigest),
		FileDigestAlgorithm:  i.FileDigestAlgorithm,
	}

	return e.EncodeElement(aux, start)
}

func (i *EncryptionInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	aux := struct {
		EncryptionKey        string `xml:"EncryptionKey"`
		MACKey               string `xml:"MacKey"`
		InitializationVector string `xml:"InitializationVector"`
		Mac                  string `xml:"Mac"`
		ProfileIdentifier    string `xml:"ProfileIdentifier"`
		FileDigest           string `xml:"FileDigest"`
		FileDigestAlgorithm  string `xml:"FileDigestAlgorithm"`
	}{}

	if err := d.DecodeElement(&aux, &start); err != nil {
		return err
	}

	var err error
	i.EncryptionKey, err = base64.StdEncoding.DecodeString(aux.EncryptionKey)
	if err != nil {
		return err
	}

	i.MACKey, err = base64.StdEncoding.DecodeString(aux.MACKey)
	if err != nil {
		return err
	}

	i.InitializationVector, err = base64.StdEncoding.DecodeString(aux.InitializationVector)
	if err != nil {
		return err
	}

	i.Mac, err = base64.StdEncoding.DecodeString(aux.Mac)
	if err != nil {
		return err
	}

	i.ProfileIdentifier = aux.ProfileIdentifier

	i.FileDigest, err = base64.StdEncoding.DecodeString(aux.FileDigest)
	if err != nil {
		return err
	}

	i.FileDigestAlgorithm = aux.FileDigestAlgorithm

	return nil
}
