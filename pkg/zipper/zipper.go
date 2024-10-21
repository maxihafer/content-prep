package zipper

import (
	"archive/zip"
	"content-prep/pkg/logger"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func Zip(fsys fs.FS, out io.Writer) error {
	zipper := zip.NewWriter(out)
	defer zipper.Close()

	return zipper.AddFS(fsys)
}

func Unzip(archive *os.File, dest string) error {
	archiveInfo, err := archive.Stat()
	if err != nil {
		return errors.Wrap(err, "failed to get archive file info")
	}

	zipper, err := zip.NewReader(archive, archiveInfo.Size())
	if err != nil {
		return errors.Wrap(err, "failed to create zip reader")
	}

	extractAndWriteFile := func(f *zip.File) error {
		var rc io.ReadCloser
		rc, err = f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		p := path.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(p, path.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", p)
		}

		err = os.MkdirAll(path.Dir(p), os.ModePerm)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			return nil
		}

		var file *os.File

		file, err = os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, rc)
		if err != nil {
			return err
		}

		return nil
	}

	for _, f := range zipper.File {
		err = extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func ZipDirectory(ctx context.Context, sourceDirectoryPath string, w io.Writer) error {
	log := logger.FromContext(ctx).With("component", "zipper", "action", "zip")

	fsys := os.DirFS(sourceDirectoryPath)

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	log.Info("compressing source directory", "source", sourceDirectoryPath)

	return errors.Wrap(zipWriter.AddFS(fsys), "failed to add file system to zip writer")
}

func UnzipToDirectory(ctx context.Context, archiveFilePath, outputFolderPath string) error {
	log := logger.FromContext(ctx).With("component", "zipper", "action", "unzip")

	archiveFile, err := os.Open(archiveFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open archive file")
	}
	defer archiveFile.Close()

	archiveFileInfo, err := os.Stat(archiveFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to get archive file info")
	}

	zipReader, err := zip.NewReader(archiveFile, archiveFileInfo.Size())
	if err != nil {
		return errors.Wrapf(err, "failed to create zip reader")
	}

	log.Info("decompressing archive", "archive", archiveFilePath, "output", outputFolderPath)

	for _, f := range zipReader.File {
		filePath := path.Join(outputFolderPath, f.Name)
		log.Info("extracting file", "file", f.Name, "path", filePath)

		if !strings.HasPrefix(filePath, path.Clean(outputFolderPath)+string(os.PathSeparator)) {
			return errors.Errorf("illegal file path: %s", filePath)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	return nil
}
