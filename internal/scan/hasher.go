package scan

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
)

type Hashes struct {
	SHA1  string
	MD5   string
	CRC32 string
	Size  int64
}

func Hash(path string) (Hashes, error) {
	f, err := os.Open(path)
	if err != nil {
		return Hashes{}, err
	}
	defer f.Close()

	sha := sha1.New()
	md := md5.New()
	crc := crc32.NewIEEE()

	n, err := io.Copy(io.MultiWriter(sha, md, crc), f)
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
