package zipkit

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

type FileToZip struct {
	Name string
	Path string
}

func WriteZip(w io.Writer, files []FileToZip) error {
	zw := zip.NewWriter(w)
	defer func() { _ = zw.Close() }()

	for _, f := range files {
		info, err := os.Stat(f.Path)
		if err != nil {
			return err
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(f.Name)
		hdr.Method = zip.Deflate
		wf, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		rf, err := os.Open(f.Path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(wf, rf); err != nil {
			_ = rf.Close()
			return err
		}
		if err := rf.Close(); err != nil {
			return err
		}
	}
	return nil
}

// func ReadZip(w)
