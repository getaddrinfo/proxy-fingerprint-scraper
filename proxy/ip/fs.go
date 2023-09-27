package ip

import (
	"bufio"
	"os"

	"go.uber.org/zap"
)

// FileSystemProxySource loads a file as series of ip:port values, where each newline represents a different value
type FileSystemProxySource struct {
	File string
}

func NewFileSystemSource(loc string) *FileSystemProxySource {
	return &FileSystemProxySource{File: loc}
}

func (f *FileSystemProxySource) Load() ([]string, error) {
	file, err := os.Open(f.File)

	// make sure the file is closed
	defer func() {
		if err := file.Close(); err != nil {
			zap.S().Error("failed to close file", err)
		}
	}()

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var ips []string
	for scanner.Scan() {
		ips = append(ips, scanner.Text())
	}

	return ips, nil
}

var _ Source = (*FileSystemProxySource)(nil)
