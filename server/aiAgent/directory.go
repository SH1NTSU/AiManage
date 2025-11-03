package aiAgent

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DirectoryNavigator handles directory operations
type DirectoryNavigator struct {
	BaseUploadPath string
}

// NewDirectoryNavigator creates a new directory navigator
func NewDirectoryNavigator(baseUploadPath string) *DirectoryNavigator {
	return &DirectoryNavigator{
		BaseUploadPath: baseUploadPath,
	}
}

// OpenDirectory opens and reads a specific directory by name
func (dn *DirectoryNavigator) OpenDirectory(folderName string) (*DirectoryInfo, error) {
	// Construct the full path
	fullPath := filepath.Join(dn.BaseUploadPath, folderName)

	// Check if directory exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory '%s' does not exist", folderName)
		}
		return nil, fmt.Errorf("error accessing directory: %w", err)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return nil, fmt.Errorf("'%s' is not a directory", folderName)
	}

	// Read directory contents
	dirInfo := &DirectoryInfo{
		Name:         folderName,
		Path:         fullPath,
		Files:        []FileInfo{},
		Subdirs:      []string{},
		LastModified: info.ModTime(),
	}

	err = dn.scanDirectory(fullPath, dirInfo)
	if err != nil {
		return nil, fmt.Errorf("error scanning directory: %w", err)
	}

	return dirInfo, nil
}

// scanDirectory recursively scans a directory and populates DirectoryInfo
func (dn *DirectoryNavigator) scanDirectory(path string, dirInfo *DirectoryInfo) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't read
		}

		if entry.IsDir() {
			dirInfo.Subdirs = append(dirInfo.Subdirs, entry.Name())
			// Recursively scan subdirectories
			err = dn.scanDirectory(fullPath, dirInfo)
			if err != nil {
				continue // Skip directories we can't read
			}
		} else {
			fileInfo := FileInfo{
				Name:      entry.Name(),
				Path:      fullPath,
				Size:      info.Size(),
				Extension: strings.TrimPrefix(filepath.Ext(entry.Name()), "."),
				Modified:  info.ModTime(),
			}
			dirInfo.Files = append(dirInfo.Files, fileInfo)
			dirInfo.TotalFiles++
			dirInfo.TotalSize += info.Size()
		}
	}

	return nil
}

// ListDirectories lists all directories in the uploads folder
func (dn *DirectoryNavigator) ListDirectories() ([]string, error) {
	var directories []string

	entries, err := os.ReadDir(dn.BaseUploadPath)
	if err != nil {
		return nil, fmt.Errorf("error reading uploads directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	return directories, nil
}

// GetFileContent reads the content of a specific file
func (dn *DirectoryNavigator) GetFileContent(folderName, fileName string) ([]byte, error) {
	fullPath := filepath.Join(dn.BaseUploadPath, folderName, fileName)

	// Security check: ensure the path is within uploads directory
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error resolving path: %w", err)
	}

	absBasePath, err := filepath.Abs(dn.BaseUploadPath)
	if err != nil {
		return nil, fmt.Errorf("error resolving base path: %w", err)
	}

	if !strings.HasPrefix(absPath, absBasePath) {
		return nil, fmt.Errorf("access denied: path outside uploads directory")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return content, nil
}

// CreateDirectory creates a new directory in uploads
func (dn *DirectoryNavigator) CreateDirectory(folderName string) error {
	fullPath := filepath.Join(dn.BaseUploadPath, folderName)

	// Check if directory already exists
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("directory '%s' already exists", folderName)
	}

	err := os.MkdirAll(fullPath, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	return nil
}

// DirectoryExists checks if a directory exists
func (dn *DirectoryNavigator) DirectoryExists(folderName string) bool {
	fullPath := filepath.Join(dn.BaseUploadPath, folderName)
	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}
