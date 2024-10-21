package zipper

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

func Zip(fsys fs.FS, out io.Writer) error {
	zipper := zip.NewWriter(out)
	defer zipper.Close()

	return addFs(zipper, fsys)
}

// We need to copypasta the AddFS method from the zip.Writer because it does not allow us to set the desired compression method
func addFs(w *zip.Writer, fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return errors.New("zip: cannot add non-regular file")
		}

		h, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		h.Name = name
		h.Method = zip.Store

		fw, err := w.CreateHeader(h)
		if err != nil {
			return err
		}

		f, err := fsys.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(fw, f)
		if err != nil {
			return err
		}

		return nil
	})
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
