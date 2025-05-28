# 🧪 gvc — Go Version Control

`gvc` is a minimal, Git-inspired version control system written in Go. It stores file contents as Git-style objects, manages trees, and handles repository metadata.

---

## 📦 Features

### ✅ Implemented

- **`init`**  
  Initializes a new `.gvc` repository structure.

- **`hash-object`**  
  Hashes a file and stores it as a Git-style compressed blob object.

- **`cat-file`**  
  Decompresses and prints the contents of a stored object.

- **`ls-tree`**  
  Lists the contents of a tree object (snapshot of a directory).

- **`write-tree`**  
  Creates a tree object representing the current index state.

- **`commit-tree`**  
  Creates a commit object from a tree SHA, optional parent, and message.

---

## 🔧 Commands & Usage

```bash
# Initialize repository
$ gvc init

# Hash a file and store it as a blob object
$ gvc hash-object -w file.txt

# View object content by hash
$ gvc cat-file -p <object-sha>

# Write a tree from current index
$ gvc write-tree

# List the contents of a tree object
$ gvc ls-tree <tree-sha>

# Create an initial commit (no parent)
$ gvc commit-tree <tree-sha> -p \"\" -m \"Initial commit\"

# Create a commit with a parent
$ gvc commit-tree <new-tree-sha> -p <parent-commit-sha> -m \"Second commit\"

## 🧩 Work in Progress (TODO)

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
└── index          # staging area (stores info of the files using add command)

```

## Built With

- Go
- zlib — For compression
- crypto/sha1 — For content hashing
- os, io, and filepath — For file operations