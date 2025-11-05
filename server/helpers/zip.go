package helpers

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Detect if all files are under a common root directory
	var rootDir string
	if len(r.File) > 0 {
		firstPath := r.File[0].Name
		parts := strings.Split(filepath.ToSlash(firstPath), "/")
		if len(parts) > 1 {
			potentialRoot := parts[0] + "/"
			allUnderRoot := true
			for _, f := range r.File {
				if !strings.HasPrefix(filepath.ToSlash(f.Name), potentialRoot) {
					allUnderRoot = false
					break
				}
			}
			if allUnderRoot {
				rootDir = potentialRoot
			}
		}
	}

	for _, f := range r.File {
		// Strip the root directory if detected
		extractPath := f.Name
		if rootDir != "" {
			extractPath = strings.TrimPrefix(filepath.ToSlash(f.Name), rootDir)
			// Skip if it's the root directory itself
			if extractPath == "" {
				continue
			}
		}

		fpath := filepath.Join(dest, extractPath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
