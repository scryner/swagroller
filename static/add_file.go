package static

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"path"
	"time"
)

func AddFile(filePath string, content []byte) error {
	// encoding to base64
	compressed, err := getCompressed(content)
	if err != nil {
		return err
	}

	// add file
	cleanedPath := path.Clean(filePath)
	escFile := &_escFile{
		local:      cleanedPath,
		size:       int64(len(content)),
		modtime:    time.Now().Unix(),
		compressed: string(compressed),
	}

	key := fmt.Sprintf("/%s", cleanedPath)
	_escData[key] = escFile

	_escStatic.prepare(key)

	return nil
}

func getCompressed(content []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	base64Encoder := base64.NewEncoder(base64.StdEncoding, buf)
	compressor := gzip.NewWriter(base64Encoder)

	_, err := compressor.Write(content)
	if err != nil {
		return nil, err
	}

	err = compressor.Flush()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
