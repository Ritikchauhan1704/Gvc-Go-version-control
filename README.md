# GVC — Go Version Control

`GVC` is a minimal, Git-inspired version control system written in Go. It mirrors some of Git's core features, such as storing file contents as Git-style objects, managing trees, and handling repository metadata.

---

## 📥 Installation

### 🔧 Prerequisites
- [Go](https://golang.org/dl/) 1.18 or higher installed

### 🛠️ Build from Source

Clone the repository and build:

For Unix/MacOS
```bash
git clone https://github.com/Ritikchauhan1704/Gvc-Go-version-control.git
cd Gvc-Go-version-control
go build -o gvc ./app
./gvc init   # Unix/macOS
```

For Windows
```bash
git clone https://github.com/Ritikchauhan1704/Gvc-Go-version-control.git
cd Gvc-Go-version-control
go build -o gvc.exe ./app
./gvc.exe init  
```
Without Building
```bash
go run app/main.go init
```

## 📦 Features

### ✅ Implemented

- **`init`**  
  Initializes a new `.gvc` repository structure.

- **`hash-object`**  
  Hashes a file and stores it as a Git-style compressed blob object.

- **`cat-file`**  
  Decompresses and prints the contents of a stored blob object.

- **`ls-tree`**  
  Lists the contents of a tree object (snapshot of the directory structure).

- **`write-tree`**  
  Creates a tree object representing the current working directory.

- **`add`**  
  Adds files to the **index** (staging area) to include in the next commit.

- **`commit`**  
  Records a snapshot of the project state with metadata (author, message, timestamp, etc.).

- **`log`**  
  Displays the commit history from the current branch.
---

## 🔧 Commands & Usage

```bash
# Initialize repository
$ gvc init

# Hash a file and store it
$ gvc hash-object -w file.txt

# View object content by hash
$ gvc cat-file -p <object-sha>

# Write a tree from working directory
$ gvc write-tree

# List the contents of a tree object
$ gvc ls-tree <tree-sha>

# commit the tree object
$ gvc commit-tree <tree-sha> -p <parent-sha> -m "message"

# adds files to the staging area
$ gvc add <file-name>

# commit the files from the staging area
$ gvc commit -m "message"

# show all the commits
$ gvc log"

```

---

## 🧩 Work in Progress (TODO)

- **`add .`**
  '.' is not supported with add as of now

- **`clone`**  
  Initializes a new gvc repository from a remote one (placeholder for now).

---

## 🗃️ Planned (Future Ideas)

- **Branching support**  
  Create and switch between branches.

- **Checkout**  
  Restore a previous version of the repository.

- **Merge**  
  Combine histories of divergent branches.

---

## 📁 Repository Structure

```
.gvc/
├── objects/       # Stores all objects (blobs, trees, commits)
├── refs/          # Stores references to branches
└── HEAD           # Points to the current branch
└── index          # staging area
```

## Built With

- Go
- zlib — For compression
- crypto/sha1 — For content hashing
- os, io, and filepath — For file operations
