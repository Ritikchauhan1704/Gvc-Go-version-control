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
	// "reflect"
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

func lsTree(treeSha string, flag string) error {
	objPath := filepath.Join(".gvc", "objects", treeSha[:2], treeSha[2:])
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

	var contentType string
	var contentLength int
	if _, err := fmt.Fscanf(zr, "%s %d\x00", &contentType, &contentLength); err != nil {
		return fmt.Errorf("failed to parse object header: %w", err)
	}
	if contentType != "tree" {
		return fmt.Errorf("expected tree object, got %s", contentType)
	}
	// Read and verify the content
	content, err := io.ReadAll(zr)
	if err != nil {
		return fmt.Errorf("failed to read tree: %w", err)
	}
	ind := 0
	for ind < len(content) {
		modeStart := ind
		for content[ind] != ' ' {
			ind++
		}
		mode := string(content[modeStart : ind])
		ind++ // skip space
		
		nameStart := ind
		for content[ind] != 0 {
			ind++
		}
		name := string(content[nameStart:ind])
		ind++ // skip null byte

		// Read 20 bytes of SHA-1 (raw binary)
		shaBytes := content[ind : ind + 20]
		sha := hex.EncodeToString(shaBytes)
		ind += 20

		if flag == "--name-only" {
			fmt.Println(name)
		} else {
			objType := "blob"
			if mode == "40000" {
				objType = "tree"
			}
			fmt.Printf("%s %s %s\t%s\n", mode, objType, sha, name)
		}
	}
	return nil
}

func writeTree(basePath string) (string, error) {
	var treeEntries []byte

	dirEntries, err := os.ReadDir(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	for _, entry := range dirEntries {
		name := entry.Name()

		// Skip .gvc directory
		if name == ".gvc" {
			continue
		}

		fullPath := filepath.Join(basePath, name)

		var entrySha string
		var entryMode string

		if entry.IsDir() {
			entrySha, err = writeTree(fullPath)
			if err != nil {
				return "", err
			}
			entryMode = "40000"
		} else {
			// Read file and create blob
			data, err := os.ReadFile(fullPath)
			if err != nil {
				return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
			}

			header := fmt.Sprintf("blob %d\x00", len(data))
			fullContent := append([]byte(header), data...)

			// Compress the blob
			var compressed bytes.Buffer
			w := zlib.NewWriter(&compressed)
			if _, err := w.Write(fullContent); err != nil {
				return "", fmt.Errorf("failed to compress blob: %w", err)
			}
			w.Close()

			// Hash it
			hash := sha1.Sum(fullContent)
			entrySha = hex.EncodeToString(hash[:])

			// Save it
			objDir := filepath.Join(".gvc", "objects", entrySha[:2])
			objPath := filepath.Join(objDir, entrySha[2:])
			if err := os.MkdirAll(objDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create blob directory: %w", err)
			}
			if err := os.WriteFile(objPath, compressed.Bytes(), 0644); err != nil {
				return "", fmt.Errorf("failed to write blob: %w", err)
			}

			entryMode = "100644"
		}

		// Build tree entry: <mode> <name>\0<20-byte SHA>
		treeEntry := fmt.Sprintf("%s %s", entryMode, name)
		treeEntryBytes := append([]byte(treeEntry), 0)
		shaBytes, _ := hex.DecodeString(entrySha)
		treeEntryBytes = append(treeEntryBytes, shaBytes...)

		treeEntries = append(treeEntries, treeEntryBytes...)
	}

	// Now make the tree object
	treeHeader := fmt.Sprintf("tree %d\x00", len(treeEntries))
	treeObject := append([]byte(treeHeader), treeEntries...)

	// Compress tree
	var compressedTree bytes.Buffer
	zw := zlib.NewWriter(&compressedTree)
	if _, err := zw.Write(treeObject); err != nil {
		return "", fmt.Errorf("failed to compress tree: %w", err)
	}
	zw.Close()

	treeHash := sha1.Sum(treeObject)
	treeSha := hex.EncodeToString(treeHash[:])

	objDir := filepath.Join(".gvc", "objects", treeSha[:2])
	objPath := filepath.Join(objDir, treeSha[2:])
	if err := os.MkdirAll(objDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tree object dir: %w", err)
	}
	if err := os.WriteFile(objPath, compressedTree.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write tree object: %w", err)
	}

	// fmt.Println(treeSha)
	return treeSha, nil
}


func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gvc <command> [<args>...]")
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
			fmt.Fprintln(os.Stderr, "usage: gvc cat-file -p <hash>")
			os.Exit(1)
		}
		if err := catFile(os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "hash-object":
		// Hash a file and store it as a blob
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: gvc hash-object -w <file>")
			os.Exit(1)
		}
		if err := hashObject(os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "ls-tree":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: gvc ls-tree [--name-only] <tree-sha>")
			os.Exit(1)
		}
		var flag, treesha string
		if len(os.Args) == 3 {
			treesha = os.Args[2]
		} else if len(os.Args) == 4 {
			flag = os.Args[2]
			treesha = os.Args[3]
			if flag != "--name-only" {
				fmt.Fprintln(os.Stderr, "usage: gvc ls-tree [--name-only] <tree-sha>")
				os.Exit(1)
			}
		}
		if err := lsTree(treesha, flag); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "write-tree":
		if len(os.Args) > 2 {
			fmt.Fprintln(os.Stderr, "usage: gvc write-tree")
			os.Exit(1)
		}
		treeSha, err := writeTree(".");
		if  err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println(treeSha)
	default:
		// Unknown command
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}
