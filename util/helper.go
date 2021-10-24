package util

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/asticode/go-astikit"
	"github.com/c4milo/unpackit"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip unzips a src into a dst.
// Possible src formats are /path/to/zip.zip or /path/to/zip.zip/internal/path.
func Unzip(ctx context.Context, src, dst string) (err error) {
	// Clean up on error
	defer func(err *error) {
		if *err != nil || ctx.Err() != nil {
			log.Debugf("Removing %s...", dst)
			os.RemoveAll(dst)
		}
	}(&err)

	// Unzipping
	log.Debugf("Unzipping %s into %s", src, dst)
	if err = astikit.Unzip(ctx, dst, src); err != nil {
		err = fmt.Errorf("unzipping %s into %s failed: %w", src, dst, err)
		return
	}
	return
}

//压缩 使用gzip压缩成tar.gz
func Compress(ctx context.Context, files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, file := range files {
		err := compress(file, "", tw)
		if err != nil {
			return err
		}
	}
	return nil
}

func compress(file *os.File, prefix string, tw *tar.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, tw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := tar.FileInfoHeader(info, "")
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// unarchive tar.gz, tar.bzip2, tar.xz, zip and tar files.
func Unpack(ctx context.Context, src, dst string) (depth string, err error) {
	// Clean up on error
	defer func(err *error) {
		if *err != nil || ctx.Err() != nil {
			log.Debugf("Removing %s...", dst)
			os.RemoveAll(dst)
		}
	}(&err)
	file, err := os.Open(src)
	if err != nil {
		return "", err
	}
	depth, err = unpackit.Unpack(file, dst)
	return depth, err
}

//解压 tar.gz
func DeCompress(ctx context.Context, tarFile, dest string) (err error) {
	// Clean up on error
	defer func(err *error) {
		if *err != nil || ctx.Err() != nil {
			log.Debugf("Removing %s...", dest)
			os.RemoveAll(dest)
		}
	}(&err)
	// Make sure the destination exists
	if err = os.MkdirAll(dest, astikit.DefaultDirMode); err != nil {
		return fmt.Errorf("astikit: mkdirall %s failed: %w", dest, err)
	}

	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		filename := filepath.Dir(dest) + "/" + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			return err
		}
		io.Copy(file, tr)
	}
	return nil
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}
