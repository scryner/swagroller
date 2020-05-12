package static

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var mFile map[string]string

func init() {
	mFile = make(map[string]string)

	for k, data := range _escData {
		if data.isDir {
			continue
		}

		mFile[k] = data.local
	}
}

func Build(title string, jsonBody []byte, outputDir string) error {
	// build index html
	indexPath := filepath.Join(outputDir, "index.html")
	indexF, err := os.OpenFile(indexPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open 'index.html': %v", err)

	}
	defer indexF.Close()

	err = MakeIndexHTML(title, jsonBody, indexF)
	if err != nil {
		return fmt.Errorf("failed to make 'index.html': %v", err)
	}

	// write other files
	for k, v := range mFile {

		// write file
		err := func() error {
			p := filepath.Join(outputDir, v)

			finfo, err := os.Stat(p)
			if err != nil {
				if os.IsNotExist(err) {
					// create file
					content := FSMustByte(false, k)

					err = ioutil.WriteFile(p, content, 0644)
					if err != nil {
						return fmt.Errorf("failed to write content to '%s': %v", p, err)
					}

					return nil
				} else {
					return fmt.Errorf("failed to stat '%s': %v", p, err)
				}
			} else {
				if finfo.IsDir() {
					return fmt.Errorf("'%s' is a directory", p)
				}
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}
