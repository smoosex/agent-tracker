package logging

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
)

var filePath string

func Init(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	filePath = path
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	return nil
}

func Path() string {
	return filePath
}

func ReadRecentLines(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 200
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make([]string, 0, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(lines) == limit {
			copy(lines, lines[1:])
			lines[len(lines)-1] = scanner.Text()
			continue
		}
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
