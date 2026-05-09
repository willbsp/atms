# atms

![Preview](preview.gif)

Another tmux sessioniser - built specifically to support workflows. A terminal UI for browsing and jumping between Git repositories open in tmux sessions.

## What it does

`atms` scans your configured directories for Git repositories, shows you branch/commit info and status at a glance, and opens (or switches to) a tmux session for the one you pick.

## Features

- Fuzzy search across all discovered repositories
- Previews branch, latest commit, status, remotes, and recent branches
- Git worktree support — worktrees appear nested under their parent repo
- Preview pane with repository / worktree details.

## Requirements

- `git`
- `tmux`

## Install

Build from source:

```sh
git clone https://github.com/willspooner/atms
cd atms
go build -o ~/.local/bin/atms
```

## Usage

```sh
atms
```

## Configuration

Config lives at `~/.config/atms/config.json` (or `$XDG_CONFIG_HOME/atms/config.json`). Created automatically on first run with defaults.

```json
{
  "search_dirs": [
    "~/repos",
    "~/Developer"
  ]
}
```

Add any directories you want scanned. `atms` searches one level deep for `.git` folders.

## Related projects
- [tmux-sessionizer](https://github.com/jrmoulton/tmux-sessionizer) (jrmoulton)
- [tmux-sessionizer](https://github.com/ThePrimeagen/tmux-sessionizer) (ThePrimeagen)
