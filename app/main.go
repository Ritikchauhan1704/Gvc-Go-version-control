package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// initializeRepo sets up a new .gvc directory structure if it doesn't already exist.

func initializeRepo() error {
	if _, err := os.Stat(".gvc"); err == nil {
		return errors.New("gvc repository already initialized")
	}

	// Create required subdirectories
	dirs := []string{".gvc", ".gvc/objects", ".gvc/refs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write the HEAD reference to point to main branch
	headContent := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".gvc/HEAD", headContent, 0644); err != nil {
		return fmt.Errorf("failed to write HEAD file: %w", err)
	}

	fmt.Println("Initialized empty gvc repository")
	return nil
}

// catFile prints the contents of a gvc object (like Git's cat-file -p).
func catFile(hash string) error {
	if len(hash) < 40 {
		return errors.New("invalid hash length")
	}

	// Build path to the object using hash
	objPath := filepath.Join(".gvc", "objects", hash[:2], hash[2:])
	data, err := os.ReadFile(objPath)
	if err != nil {
		return fmt.Errorf("failed to read gvc object: %w", err)
	}

	// Decompress the object
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decompress object: %w", err)
	}
	defer zr.Close()

	// Read the object header (e.g., "blob 14")
	var contentType string
	var contentLength int
	if _, err := fmt.Fscanf(zr, "%s %d\x00", &contentType, &contentLength); err != nil {
		return fmt.Errorf("failed to parse object header: %w", err)
	}

	// Read and verify the content
	content, err := io.ReadAll(zr)
	if err != nil {
		return fmt.Errorf("failed to read decompressed content: %w", err)
	}
	if len(content) != contentLength {
		return fmt.Errorf("content length mismatch: expected %d, got %d", contentLength, len(content))
	}

	// Print the actual blob content
	fmt.Printf("%s", content)
	return nil
}

// hashObject reads a file, wraps it in a Git-style blob, compresses, hashes, and stores it.
func hashObject(fpath string) error {
	// Read file contents
	data, err := os.ReadFile(fpath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", fpath, err)
	}

	// Prepare the blob object header
	header := fmt.Sprintf("blob %d\x00", len(data))
	fullContent := append([]byte(header), data...)

	// Compress the blob
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if _, err := w.Write(fullContent); err != nil {
		return fmt.Errorf("failed to compress object: %w", err)
	}
	w.Close()

	// Generate SHA-1 hash of the blob
	hashBytes := sha1.Sum(fullContent)
	hash := hex.EncodeToString(hashBytes[:])

	// Determine object path based on the hash
	objDir := filepath.Join(".gvc", "objects", hash[:2])
	objPath := filepath.Join(objDir, hash[2:])

	// Create directory and store the compressed object
	if err := os.MkdirAll(objDir, 0755); err != nil {
		return fmt.Errorf("failed to create object directory: %w", err)
	}
	if err := os.WriteFile(objPath, compressed.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write object: %w", err)
	}

	// Output the SHA-1 hash
	fmt.Println(hash)
	return nil
}

// main parses CLI arguments and routes to appropriate command handler
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: mygvc <command> [<args>...]")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		// Initialize .gvc directory
		if err := initializeRepo(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "cat-file":
		// Print contents of a blob object
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: mygvc cat-file -p <hash>")
			os.Exit(1)
		}
		if err := catFile(os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "hash-object":
		// Hash a file and store it as a blob
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: mygvc hash-object -w <file>")
			os.Exit(1)
		}
		if err := hashObject(os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	default:
		// Unknown command
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}
