# 🧪 gvc — Go Version Control

`gvc` is a minimal, Git-inspired version control system written in Go. It mirrors some of Git's core features, such as storing file contents as Git-style objects, managing trees, and handling repository metadata.

---

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
```

---

## 🧩 Work in Progress (TODO)

- **`add`**  
  Adds files to the **index** (staging area) to include in the next commit.

- **`commit`**  
  Records a snapshot of the project state with metadata (author, message, timestamp, etc.).

- **`log`**  
  Displays the commit history from the current branch.

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
```

## Built With

- Go
- zlib — For compression
- crypto/sha1 — For content hashing
- os, io, and filepath — For file operations