package scan

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Hashes struct {
	SHA1  string
	MD5   string
	CRC32 string
	Size  int64
}

func Hash(path string) (Hashes, error) {
	if strings.EqualFold(filepath.Ext(path), ".zip") {
		return hashZip(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return Hashes{}, err
	}
	defer f.Close()

	return hashReader(f)
}

func hashZip(path string) (Hashes, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return Hashes{}, err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.FileInfo().IsDir() || strings.HasPrefix(filepath.Base(f.Name), ".") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return Hashes{}, err
		}
		h, hashErr := hashReader(rc)
		closeErr := rc.Close()
		if hashErr != nil {
			return Hashes{}, hashErr
		}
		if closeErr != nil {
			return Hashes{}, closeErr
		}
		return h, nil
	}
	return Hashes{}, fmt.Errorf("zip contains no ROM file: %s", path)
}

func hashReader(r io.Reader) (Hashes, error) {
	sha := sha1.New()
	md := md5.New()
	crc := crc32.NewIEEE()

	n, err := io.Copy(io.MultiWriter(sha, md, crc), r)
	if err != nil {
		return Hashes{}, err
	}

	return Hashes{
		SHA1:  hex.EncodeToString(sha.Sum(nil)),
		MD5:   hex.EncodeToString(md.Sum(nil)),
		CRC32: hex.EncodeToString(crc.Sum(nil)),
		Size:  n,
	}, nil
}
