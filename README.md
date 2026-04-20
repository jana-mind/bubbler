# bubbler

A kanban board for your terminal, living inside your git repo.

## How it works

Your board lives in a `.bubble` directory at the repo root — two kinds of files: one for the overall board state (issues, titles, tags, which column everything is in), and one per issue for everything else: description, comments, column changes, the full history. Nothing ever gets overwritten. Every change is appended, and attributed to whoever made it using the local git identity (`user.name` + `user.email`).

Run `bubbler` from anywhere in your repo and it'll find its way.

## Keeping local project state and board state independent

If `.bubble` is a git submodule, bubbler manages it for you — pulling before reads, committing and pushing after writes. Your board history stays completely separate from your main project history, so everyone on the team sees the current board state no matter which branch they're on. You can disable this behaviour by creating a `.bubble-manual` file inside the `.bubble` directory.

If `.bubble` is just a regular directory, bubbler doesn't touch git at all. You're in control.

The one-file-per-issue layout also means merge conflicts are rare in practice — two people would have to edit the exact same issue at the same time for one to occur.

## What's in v1

- Single board with customizable columns (defaults: `waiting`, `in-progress`, `completed`)
- Issues with title, tags, description, comments.
- Full per-issue history log
- Shell completion for bash, zsh, and PowerShell
- `go install` or pre-built binaries via GitHub Releases

Built with [cobra](https://github.com/spf13/cobra) and [go-git](https://github.com/go-git/go-git).
