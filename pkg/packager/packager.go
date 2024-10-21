package packager

import (
	"content-prep/pkg/cryptostream"
	"content-prep/pkg/logger"
	"content-prep/pkg/zipper"
	"context"
	"crypto/sha256"
	"encoding/xml"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

var Default = &packager{
	keygen: defaultKeyGenerator{},
}

type KeyGenerator interface {
	GenerateKey(length int) ([]byte, error)
}

type packager struct {
	keygen KeyGenerator
}

const (
	packageFileName      = "IntunePackage.intunewin"
	PackageFileExtension = ".intunewin"
)

func (p *packager) CreatePackage(ctx context.Context, source fs.FS, setupFile string, output io.Writer) error {
	log := logger.FromContext(ctx).With("component", "packager", "action", "create")

	log.Info("creating package", "source", source, "setupFile", setupFile, "output", output)

	tempDirPath, err := os.MkdirTemp(os.TempDir(), "content-prep-packager-*")
	if err != nil {
		return errors.Wrapf(err, "failed to create temporary directory")
	}
	log.Debug("created temporary directory", "path", tempDirPath)

	workDirPath := path.Join(tempDirPath, "IntuneWinPackage")
	if err := os.Mkdir(workDirPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create work directory")
	}
	log.Debug("created work directory", "path", workDirPath)

	contentsFolderPath := path.Join(workDirPath, "Contents")
	if err := os.Mkdir(contentsFolderPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create 'Contents' folder")
	}
	log.Debug("created contents folder", "path", contentsFolderPath)

	compressedPackageFilePath := path.Join(contentsFolderPath, packageFileName+".zip")
	compressedPackageFile, err := os.Create(compressedPackageFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open compressed package file")
	}
	defer compressedPackageFile.Close()
	log.Debug("created compressed package file", "path", compressedPackageFilePath)

	if err := zipper.Zip(source, compressedPackageFile); err != nil {
		return errors.Wrapf(err, "failed to create compressed package")
	}
	log.Info("compressed source folder", "source", source, "archive", compressedPackageFilePath)

	encryptedPackageFilePath := path.Join(contentsFolderPath, packageFileName)
	encryptedPackageFile, err := os.Create(encryptedPackageFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open compressed package file")
	}
	defer encryptedPackageFile.Close()
	log.Debug("created encrypted package file", "path", encryptedPackageFilePath)

	aesKey, err := p.keygen.GenerateKey(32)
	if err != nil {
		return errors.Wrapf(err, "failed to generate AES key")
	}

	iv, err := p.keygen.GenerateKey(cryptostream.IvSize)
	if err != nil {
		return errors.Wrapf(err, "failed to generate initialization vector")
	}

	hmacKey, err := p.keygen.GenerateKey(cryptostream.HMACKeySize)
	if err != nil {
		return errors.Wrapf(err, "failed to generate HMAC key")
	}

	_, err = compressedPackageFile.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrapf(err, "failed to seek to start of compressed package file")
	}

	if err := cryptostream.Encrypt(compressedPackageFile, encryptedPackageFile, aesKey, iv, hmacKey); err != nil {
		return errors.Wrapf(err, "failed to encrypt compressed package file")
	}
	log.Debug("encrypted archive", "archive", compressedPackageFilePath, "encrypted", encryptedPackageFilePath)

	mac := make([]byte, sha256.Size)
	_, err = encryptedPackageFile.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrapf(err, "failed to seek to start of encrypted package file")
	}
	_, err = encryptedPackageFile.Read(mac)
	if err != nil {
		return errors.Wrapf(err, "failed to read HMAC from encrypted package file")
	}

	compressedPackageFileInfo, err := compressedPackageFile.Stat()
	if err != nil {
		return errors.Wrapf(err, "failed to get compressed package file info")
	}
	log.Debug("got compressed package file info", "size", compressedPackageFileInfo.Size())

	hash := sha256.New()
	if _, err := io.Copy(hash, compressedPackageFile); err != nil {
		return errors.Wrapf(err, "failed to hash compressed package file")
	}

	digest := hash.Sum(nil)
	log.Debug("generated digest of compressed Package file", "digest", digest)

	setupFileName := strings.Trim(path.Base(setupFile), path.Ext(setupFile))

	applicationInfo := &ApplicationInfo{
		FileName:               packageFileName,
		Name:                   setupFileName,
		UnencryptedContentSize: compressedPackageFileInfo.Size(),
		SetupFile:              path.Base(setupFile),
		EncryptionInfo: EncryptionInfo{
			EncryptionKey:        aesKey,
			MACKey:               hmacKey,
			InitializationVector: iv,
			Mac:                  mac,
			ProfileIdentifier:    "ProfileVersion1",
			FileDigest:           digest,
			FileDigestAlgorithm:  "SHA256",
		},
	}

	metadataFolderPath := path.Join(workDirPath, "Metadata")
	if err := os.Mkdir(metadataFolderPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create 'Metadata' folder")
	}
	log.Debug("created metadata folder", "path", metadataFolderPath)

	detectionFilePath := path.Join(metadataFolderPath, "Detection.xml")
	detectionFile, err := os.Create(detectionFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to create detection file")
	}
	defer detectionFile.Close()
	log.Debug("created detection file", "path", detectionFilePath)

	if err := xml.NewEncoder(detectionFile).Encode(applicationInfo); err != nil {
		return errors.Wrapf(err, "failed to write application info")
	}
	log.Debug("wrote application info to detection file")

	if err := os.Remove(compressedPackageFilePath); err != nil {
		return errors.Wrapf(err, "failed to remove compressed package file")
	}
	log.Debug("removed compressed package file", "path", compressedPackageFilePath)

	packageFS := os.DirFS(tempDirPath)

	if err := zipper.Zip(packageFS, output); err != nil {
		return errors.Wrapf(err, "failed to create output package")
	}

	return nil
}

func (p *packager) DecryptPackage(ctx context.Context, packageFile *os.File, destDir string) error {
	_ = logger.FromContext(ctx).With("component", "packager", "action", "decrypt")

	if err := zipper.Unzip(packageFile, destDir); err != nil {
		return errors.Wrapf(err, "failed to extract package")
	}

	uncompressedFilePath := path.Join(destDir, "IntuneWinPackage")

	detectionFilePath := path.Join(uncompressedFilePath, "Metadata", "Detection.xml")
	detectionFile, err := os.Open(detectionFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open detection file")
	}
	defer detectionFile.Close()

	var applicationInfo ApplicationInfo
	if err := xml.NewDecoder(detectionFile).Decode(&applicationInfo); err != nil {
		return errors.Wrapf(err, "failed to read application info")
	}

	encryptedPackageFilePath := path.Join(uncompressedFilePath, "Contents", applicationInfo.FileName)
	encryptedPackageFile, err := os.Open(encryptedPackageFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open encrypted package file")
	}
	defer encryptedPackageFile.Close()

	decryptedPackageFilePath := path.Join(uncompressedFilePath, "Contents", applicationInfo.FileName+".zip")
	decryptedPackageFile, err := os.Create(decryptedPackageFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to create decrypted package file")
	}
	defer decryptedPackageFile.Close()

	digester := sha256.New()

	out := io.MultiWriter(decryptedPackageFile, digester)

	err = cryptostream.Decrypt(encryptedPackageFile, out, applicationInfo.EncryptionInfo.EncryptionKey, applicationInfo.EncryptionInfo.MACKey)
	if err != nil {
		return errors.Wrapf(err, "failed to decrypt package file")
	}

	return nil
}
