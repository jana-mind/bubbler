# bubbler

[![CI](https://github.com/jana-mind/bubbler/actions/workflows/ci.yml/badge.svg)](https://github.com/jana-mind/bubbler/actions/workflows/ci.yml) [![Release](https://github.com/jana-mind/bubbler/actions/workflows/release.yml/badge.svg)](https://github.com/jana-mind/bubbler/actions/workflows/release.yml)

Sand **bubbler** crabs are methodical. *They sit at their spot, pick up sand, extract what's useful, disregard the rest and set it aside.* bubbler strives to be like this, but for in-repo kanban boards. Simple, yet powerful: building big projects with simple boards that don't get in your way.

(And also it is using bubbletea for it's highly WIP TUI and bubbler worked quite well with that)

A kanban board for your terminal, living inside your git repo.

Meant for simple projects that like to have a local board state rather than external tools.

## Installation

### go install

Install the latest version directly:

```
go install github.com/jana-mind/bubbler@latest
```

### GitHub Releases

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases page](https://github.com/jana-mind/bubbler/releases). Download the archive for your platform, extract it, and add the binary to your PATH.

## Setup

Initialize bubbler in an existing git repository:

```
cd your-project
bubbler init
```

This creates `.bubble/` containing `default.yaml` and a `default/` directory for issue files.

### Using .bubble as a git submodule

If you want bubbler to automatically commit and push board changes alongside your project, initialize `.bubble/` as a git submodule of your project:

```
bubbler init --submodule
```

This creates `.bubble/` as an empty submodule that bubbler will manage. On every board write, bubbler stages changes, commits with a descriptive message, and pushes the submodule — keeping your board history separate from your project history.

After cloning a repo that has `.bubble` as a submodule:

```
git submodule update --init
```

To switch an existing `.bubble/` directory to submodule mode:

```
git submodule add <your-repo-url> .bubble
```

To update bubbler when using it as a submodule:

```
cd .bubble && git pull && cd .. && git add .bubble && git commit -m "Update bubbler"
```

## Core Commands

### Creating an issue

Interactive mode, prompts for title and opens your editor for description:

```
bubbler issue create
```

With flags to skip the prompts:

```
bubbler issue create --title "Fix login timeout" --tag bug --tag urgent
```

### Listing issues

Show all open issues across columns:

```
bubbler issue list
```

Show completed issues too:

```
bubbler issue list --all
```

Filter by column:

```
bubbler issue list --column in-progress
```

Filter by tag (repeatable, AND logic):

```
bubbler issue list --tag bug --tag urgent
```

### Moving an issue

```
bubbler issue move a1b2c3 in-progress
```

### Commenting on an issue

With a message flag:

```
bubbler issue comment a1b2c3 --message "This is reproduced on Safari 17"
```

Without flags, opens your editor:

```
bubbler issue comment a1b2c3
```

### Full issue detail

```
bubbler issue show a1b2c3
```

### Board management

List columns:

```
bubbler board columns
```

Add or remove columns:

```
bubbler board column add "In Review"
bubbler board column remove "In Review"
```

List, add, and remove tags:

```
bubbler board tags
bubbler board tag add docs
bubbler board tag remove chore
```

## Shell Completion

### bash

```
# Add to ~/.bashrc
source <(bubbler completion bash)
```

Or install system-wide:

```
sudo cp $(bubbler completion bash) /etc/bash_completion.d/bubbler
```

### zsh

```
# Add to ~/.zshrc
source <(bubbler completion zsh)
```

### PowerShell

```
bubbler completion powershell > bubbler.ps1
Import-Module ./bubbler.ps1
```

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
