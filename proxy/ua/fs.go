package ua

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"

	"go.uber.org/zap"
)

type FileSystemSource struct {
	File string

	items []string
}

func NewFileSystemSource(path string) *FileSystemSource {
	return &FileSystemSource{
		File:  path,
		items: nil,
	}
}

func (f *FileSystemSource) List() ([]string, error) {
	if f.items != nil {
		return f.items, nil
	}

	file, err := os.Open(f.File)

	// make sure the file is closed
	defer func() {
		if err := file.Close(); err != nil {
			zap.S().Error("failed to close file: %s", err)
		}
	}()

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var userAgents []string
	for scanner.Scan() {
		userAgents = append(userAgents, scanner.Text())
	}

	f.items = userAgents

	return userAgents, nil
}

func (fs *FileSystemSource) Random() (string, error) {
	if fs.items == nil {
		if _, err := fs.List(); err != nil {
			return "", err
		}
	}

	if len(fs.items) == 0 {
		return "", fmt.Errorf("no user agents present")
	}

	return fs.items[rand.Intn(len(fs.items))], nil
}

var _ Source = (*FileSystemSource)(nil)
