package storage

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"

	"maphraohom.app/maphraohom-backoffice/config"
)

func (f *FileSystem) getFileFromLocalStorage(path string) (*os.File, error) {
	return os.Open(fmt.Sprintf("%s/%s", config.Global.FileSystem.LocalPath, path))
}

func (f *FileSystem) putFileToLocalStorage(path string, fileName string, fileParam interface{}) error {
	var (
		filePath = fmt.Sprintf("%s/%s/%s", config.Global.FileSystem.LocalPath, path, fileName)
		content  []byte
		err      error
	)

	// open input file
	if fh, ok := fileParam.(*multipart.FileHeader); ok {
		mf, err := fh.Open()
		if err != nil {
			return err
		}

		// Get file content
		content, err = io.ReadAll(mf)
		if err != nil {
			return fmt.Errorf("[App] FileSystemError: failed to read, %w", err)
		}

		// close fi on exit and check for its returned error
		defer func() {
			if err = mf.Close(); err != nil {
				log.Printf("[App] FileSystemError: %v\n", err)
			}
		}()
	} else if osf, ok := fileParam.(*os.File); ok {
		// Get file content
		content, err = io.ReadAll(osf)
		if err != nil {
			return fmt.Errorf("[App] [App] FileSystemError: failed to read, %w", err)
		}

		// close fi on exit and check for its returned error
		defer func() {
			if err = osf.Close(); err != nil {
				log.Printf("[App] FileSystemError: %v\n", err)
			}
		}()
	} else {
		return fmt.Errorf("[App] FileSystemError: invalid file parameter")
	}

	// Write the content
	if err = os.WriteFile(filePath, content, 0777); err != nil {
		return err
	}

	// File uploaded logging
	log.Printf("[App] FileSystem (local): file uploaded to '%s'", filePath)

	return nil
}

func (f *FileSystem) moveFileInLocalStorage(from string, to string) error {
	var err error

	// Set local file path
	localFromPath := fmt.Sprintf("%s/%s", config.Global.FileSystem.LocalPath, from)
	localToPath := fmt.Sprintf("%s/%s", config.Global.FileSystem.LocalPath, to)

	// Move file in local storage
	if err = os.Rename(localFromPath, localToPath); err != nil {
		return err
	}

	// File moved logging
	log.Printf("FileSystem (local): file moved from '%s' to '%s'", from, to)

	return nil
}

func (f *FileSystem) deleteFileInLocalStorage(path string) error {
	var err error

	// Delete file in local storage
	if err = os.Remove(fmt.Sprintf("%s/%s", config.Global.FileSystem.LocalPath, path)); err != nil {
		return err
	}

	// File removed logging
	log.Printf("FileSystem (local): file removed at '%s'", path)

	return nil
}
