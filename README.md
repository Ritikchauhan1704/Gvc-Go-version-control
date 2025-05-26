# mygvc

`mygvc` is a simplified version control system written in Go, inspired by Git. It supports basic functionality like initializing a repository, hashing file contents as Git-style blobs, and reading stored objects.

---

## 📦 Features

- `init`: Initializes a new `.gvc` repository structure.
- `hash-object`: Hashes a file and stores it as a Git-style blob object.
- `cat-file`: Decompresses and prints the contents of a stored blob object.

---

## 📁 TODO

- `ls-tree`: Lists the contents of a tree object (files and directories tracked in a snapshot).
- `write-tree`: Writes the current index (staging area) as a tree object.
- `add`: Adds files to the staging area (index) to be included in the next commit.
- `commit`: Records a snapshot of the project with a message and metadata (author, timestamp, etc.).
- `log`: Displays the commit history starting from the current branch.

