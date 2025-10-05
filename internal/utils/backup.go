package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// CreateBackup creates a backup of the specified file using SHA256 checksum
// The backup file is named with the format: <original>.sha256[0:32]
// If a backup with the same checksum already exists, no new backup is created (idempotency)
func CreateBackup(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, nothing to backup
		return nil
	}

	// Calculate SHA256 checksum of the file
	checksum, err := calculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("calculate checksum: %w", err)
	}

	// Use first 32 characters of checksum for backup filename
	backupPath := fmt.Sprintf("%s.%s", filePath, checksum[:32])

	// Check if backup already exists (same content, no need to backup again)
	if _, err := os.Stat(backupPath); err == nil {
		return nil // Backup with same checksum already exists
	}

	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("create backup file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	return nil
}

// calculateSHA256 calculates the SHA256 checksum of a file efficiently
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	// Use a buffer to read file in chunks for efficiency
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			hash.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
