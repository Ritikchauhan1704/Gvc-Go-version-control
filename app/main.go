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
	"sort"
	"time"
)

const (
	GvcDir     = ".gvc"
	ObjectsDir = ".gvc/objects"
	RefsDir    = ".gvc/refs"
	HeadFile   = ".gvc/HEAD"
)

// ObjectType represents the type of Git object
type ObjectType string

const (
	BlobObject   ObjectType = "blob"
	TreeObject   ObjectType = "tree"
	CommitObject ObjectType = "commit"
)

// TreeEntry represents an entry in a tree object
type TreeEntry struct {
	Mode string
	Name string
	SHA  string
	Type ObjectType
}

// initializeRepo sets up a new .gvc directory structure if it doesn't already exist.
func initializeRepo() error {
	if _, err := os.Stat(GvcDir); err == nil {
		return errors.New("gvc repository already initialized")
	}

	// Create required subdirectories
	dirs := []string{GvcDir, ObjectsDir, RefsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write the HEAD reference to point to main branch
	headContent := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(HeadFile, headContent, 0644); err != nil {
		return fmt.Errorf("failed to write HEAD file: %w", err)
	}

	fmt.Println("Initialized empty gvc repository")
	return nil
}

// validateSHA checks if the provided SHA is valid
func validateSHA(sha string) error {
	if len(sha) != 40 {
		return fmt.Errorf("invalid SHA length: expected 40, got %d", len(sha))
	}
	if _, err := hex.DecodeString(sha); err != nil {
		return fmt.Errorf("invalid SHA format: %w", err)
	}
	return nil
}

// getObjectPath returns the file path for a given object SHA
func getObjectPath(sha string) string {
	return filepath.Join(ObjectsDir, sha[:2], sha[2:])
}

// readObject reads and decompresses a Git object
func readObject(sha string) (ObjectType, []byte, error) {
	if err := validateSHA(sha); err != nil {
		return "", nil, err
	}

	objPath := getObjectPath(sha)
	data, err := os.ReadFile(objPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read object %s: %w", sha, err)
	}

	// Decompress the object
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", nil, fmt.Errorf("failed to decompress object: %w", err)
	}
	defer zr.Close()

	// Read the object header (e.g., "blob 14")
	var objectType string
	var contentLength int
	if _, err := fmt.Fscanf(zr, "%s %d\x00", &objectType, &contentLength); err != nil {
		return "", nil, fmt.Errorf("failed to parse object header: %w", err)
	}

	// Read and verify the content
	content, err := io.ReadAll(zr)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read decompressed content: %w", err)
	}

	if len(content) != contentLength {
		return "", nil, fmt.Errorf("content length mismatch: expected %d, got %d", contentLength, len(content))
	}

	return ObjectType(objectType), content, nil
}

// writeObject compresses and stores an object, returning its SHA
func writeObject(objectType ObjectType, content []byte) (string, error) {
	// Prepare the object header
	header := fmt.Sprintf("%s %d\x00", objectType, len(content))
	fullContent := append([]byte(header), content...)

	// Generate SHA-1 hash
	hashBytes := sha1.Sum(fullContent)
	sha := hex.EncodeToString(hashBytes[:])

	// Compress the object
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if _, err := w.Write(fullContent); err != nil {
		return "", fmt.Errorf("failed to compress object: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close compressor: %w", err)
	}

	// Store the compressed object
	objDir := filepath.Join(ObjectsDir, sha[:2])
	objPath := filepath.Join(objDir, sha[2:])

	if err := os.MkdirAll(objDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create object directory: %w", err)
	}

	if err := os.WriteFile(objPath, compressed.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write object: %w", err)
	}

	return sha, nil
}

// catFile prints the contents of a gvc object (like Git's cat-file -p)
func catFile(sha string) error {
	objectType, content, err := readObject(sha)
	if err != nil {
		return err
	}

	switch objectType {
	case BlobObject:
		fmt.Print(string(content))
	case TreeObject, CommitObject:
		fmt.Print(string(content))
	default:
		return fmt.Errorf("unknown object type: %s", objectType)
	}

	return nil
}

// hashObject reads a file, creates a blob object, and stores it
func hashObject(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	sha, err := writeObject(BlobObject, data)
	if err != nil {
		return err
	}

	fmt.Println(sha)
	return nil
}

// parseTreeEntries parses tree object content into structured entries
func parseTreeEntries(content []byte) ([]TreeEntry, error) {
	var entries []TreeEntry
	index := 0

	for index < len(content) {
		// Read mode
		modeStart := index
		for index < len(content) && content[index] != ' ' {
			index++
		}
		if index >= len(content) {
			return nil, errors.New("malformed tree: missing space after mode")
		}
		mode := string(content[modeStart:index])
		index++ // skip space

		// Read name
		nameStart := index
		for index < len(content) && content[index] != 0 {
			index++
		}
		if index >= len(content) {
			return nil, errors.New("malformed tree: missing null byte after name")
		}
		name := string(content[nameStart:index])
		index++ // skip null byte

		// Read SHA (20 bytes)
		if index+20 > len(content) {
			return nil, errors.New("malformed tree: incomplete SHA")
		}
		shaBytes := content[index : index+20]
		sha := hex.EncodeToString(shaBytes)
		index += 20

		// Determine object type based on mode
		var objType ObjectType
		switch mode {
		case "40000":
			objType = TreeObject
		default:
			objType = BlobObject
		}

		entries = append(entries, TreeEntry{
			Mode: mode,
			Name: name,
			SHA:  sha,
			Type: objType,
		})
	}

	return entries, nil
}

// lsTree lists the contents of a tree object
func lsTree(treeSHA string, nameOnly bool) error {
	objectType, content, err := readObject(treeSHA)
	if err != nil {
		return err
	}

	if objectType != TreeObject {
		return fmt.Errorf("expected tree object, got %s", objectType)
	}

	entries, err := parseTreeEntries(content)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if nameOnly {
			fmt.Println(entry.Name)
		} else {
			fmt.Printf("%s %s %s\t%s\n", entry.Mode, entry.Type, entry.SHA, entry.Name)
		}
	}

	return nil
}

// writeTree recursively creates tree objects for a directory
func writeTree(basePath string) (string, error) {
	dirEntries, err := os.ReadDir(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	var treeEntries []TreeEntry

	for _, entry := range dirEntries {
		name := entry.Name()

		// Skip .gvc directory
		if name == GvcDir {
			continue
		}

		fullPath := filepath.Join(basePath, name)
		var entrySHA string
		var entryMode string
		var entryType ObjectType

		if entry.IsDir() {
			entrySHA, err = writeTree(fullPath)
			if err != nil {
				return "", err
			}
			entryMode = "40000"
			entryType = TreeObject
		} else {
			// Read file and create blob
			data, err := os.ReadFile(fullPath)
			if err != nil {
				return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
			}

			entrySHA, err = writeObject(BlobObject, data)
			if err != nil {
				return "", err
			}

			entryMode = "100644"
			entryType = BlobObject
		}

		treeEntries = append(treeEntries, TreeEntry{
			Mode: entryMode,
			Name: name,
			SHA:  entrySHA,
			Type: entryType,
		})
	}

	// Sort entries by name (Git requirement)
	sort.Slice(treeEntries, func(i, j int) bool {
		return treeEntries[i].Name < treeEntries[j].Name
	})

	// Build tree content
	var treeContent bytes.Buffer
	for _, entry := range treeEntries {
		// Format: <mode> <name>\0<20-byte SHA>
		treeContent.WriteString(fmt.Sprintf("%s %s", entry.Mode, entry.Name))
		treeContent.WriteByte(0)
		shaBytes, _ := hex.DecodeString(entry.SHA)
		treeContent.Write(shaBytes)
	}

	return writeObject(TreeObject, treeContent.Bytes())
}

// commitTree creates a commit object
func commitTree(treeSHA, parentSHA, message string) (string, error) {
	if err := validateSHA(treeSHA); err != nil {
		return "", fmt.Errorf("invalid tree SHA: %w", err)
	}

	if parentSHA != "" {
		if err := validateSHA(parentSHA); err != nil {
			return "", fmt.Errorf("invalid parent SHA: %w", err)
		}
	}

	author := "gvc <Ritik Chauhan> <critik1704@gmail.com>"
	timestamp := fmt.Sprintf("%d +0000", time.Now().Unix())

	var commitContent bytes.Buffer
	commitContent.WriteString(fmt.Sprintf("tree %s\n", treeSHA))
	if parentSHA != "" {
		commitContent.WriteString(fmt.Sprintf("parent %s\n", parentSHA))
	}
	commitContent.WriteString(fmt.Sprintf("author %s %s\ncommitter %s %s\n\n%s\n",
		author, timestamp, author, timestamp, message))

	return writeObject(CommitObject, commitContent.Bytes())
}

// Command handlers
func handleInit() error {
	return initializeRepo()
}

func handleCatFile(args []string) error {
	if len(args) < 2 || args[0] != "-p" {
		return errors.New("usage: gvc cat-file -p <hash>")
	}
	return catFile(args[1])
}

func handleHashObject(args []string) error {
	if len(args) < 2 || args[0] != "-w" {
		return errors.New("usage: gvc hash-object -w <file>")
	}
	return hashObject(args[1])
}

func handleLsTree(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: gvc ls-tree [--name-only] <tree-sha>")
	}

	var nameOnly bool
	var treeSHA string

	if len(args) == 1 {
		treeSHA = args[0]
	} else if len(args) == 2 && args[0] == "--name-only" {
		nameOnly = true
		treeSHA = args[1]
	} else {
		return errors.New("usage: gvc ls-tree [--name-only] <tree-sha>")
	}

	return lsTree(treeSHA, nameOnly)
}

func handleWriteTree(args []string) error {
	if len(args) > 0 {
		return errors.New("usage: gvc write-tree")
	}

	treeSHA, err := writeTree(".")
	if err != nil {
		return err
	}

	fmt.Println(treeSHA)
	return nil
}

func handleCommitTree(args []string) error {
	if len(args) < 5 || args[1] != "-p" || args[3] != "-m" {
		return errors.New("usage: gvc commit-tree <tree_sha> -p <parent_sha> -m <commit_message>")
	}

	treeSHA := args[0]
	parentSHA := args[2]
	message := args[4]

	commitSHA, err := commitTree(treeSHA, parentSHA, message)
	if err != nil {
		return err
	}

	fmt.Print(commitSHA)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gvc <command> [<args>...]")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "init":
		err = handleInit()
	case "cat-file":
		err = handleCatFile(args)
	case "hash-object":
		err = handleHashObject(args)
	case "ls-tree":
		err = handleLsTree(args)
	case "write-tree":
		err = handleWriteTree(args)
	case "commit-tree":
		err = handleCommitTree(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}