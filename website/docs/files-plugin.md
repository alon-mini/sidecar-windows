---
sidebar_position: 4
title: Files Plugin
---

# Files Plugin

A full-featured terminal file browser with syntax-highlighted previews, project-wide search powered by ripgrep, markdown rendering, and complete file operations—no leaving the terminal.

![Files Plugin](/img/sidecar-files.png)

## Key Capabilities

- **Instant search across millions of files**: Fuzzy file finder caches 50,000 files with sub-second response
- **Ripgrep-powered project search**: Find any text across your codebase in milliseconds with regex support
- **Rich content previews**: Syntax highlighting for code, rendered markdown, and terminal graphics for images
- **Live file watching**: Preview updates automatically when files change on disk
- **Full file operations**: Create, rename, move, delete, yank/paste—all with safety confirmations
- **Persistent state**: Your cursor position, expanded folders, and layout survive restarts
- **Two-pane interface**: Resizable tree and preview with vim keybindings throughout

## Quick Start

| Key | Action |
|-----|--------|
| `ctrl+p` | Quick open (fuzzy file search) |
| `ctrl+s` | Search across entire project |
| `/` | Filter visible files in tree |
| `?` | Search within current file |
| `\` | Toggle tree pane |

## Core Concepts

The plugin has two main panes:

- **Tree pane** (left): Navigate your project structure, perform file operations
- **Preview pane** (right): View file contents with syntax highlighting and markdown rendering

Drag the divider between panes to resize. Toggle tree visibility with `\` to maximize preview space.

## Navigation

### Tree Navigation (Left Pane)

Vim-style movement through your project structure.

| Key | Action |
|-----|--------|
| `j`, `↓` | Move down |
| `k`, `↑` | Move up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `ctrl+d` | Page down |
| `ctrl+u` | Page up |
| `l`, `→` | Expand directory or preview file |
| `h`, `←` | Collapse directory or jump to parent |
| `enter` | Toggle directory or preview file |

### Search Features

Four search modes, each optimized for different scenarios:

#### Quick Open (`ctrl+p`)

Fuzzy file finder that caches up to 50,000 files for instant results. Type any part of the filename—no exact matches needed.

```
Example: "mdplug" matches "website/docs/files-plugin.md"
```

#### Project Search (`ctrl+s`)

Full-text search across your entire codebase using ripgrep. Supports regex, case sensitivity toggles, and whole-word matching. Shows up to 1,000 matches with context.

```
Example: Search "TODO" to find all pending tasks
Toggle regex mode for pattern matching
```

#### Tree Filter (`/`)

Filter visible files in the tree by name. Great for quick navigation in the current view.

#### Content Search (`?`)

Search within the currently previewed file. Use `n`/`N` to jump between matches.

## File Preview (Right Pane)

### Scrolling

| Key | Action |
|-----|--------|
| `j`, `↓` | Scroll down |
| `k`, `↑` | Scroll up |
| `ctrl+d` | Half page down |
| `ctrl+u` | Half page up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `h`, `←`, `esc` | Return to tree |

### Rich Content Support

**Syntax Highlighting**
Automatic language detection and highlighting for 100+ programming languages. Uses the Chroma highlighter with theme matching your terminal configuration.

**Markdown Rendering**
Press `m` to toggle between raw markdown and rendered output with styled headings, lists, code blocks, and links. Perfect for viewing README files.

**Image Preview**
Displays images directly in the terminal using graphics protocols (Kitty, iTerm2). Automatically detected for PNG, JPG, GIF, and more.

**Smart File Handling**
- Large files (>500KB): Shows truncated preview with file size warning
- Binary files: Displays metadata instead of corrupted content
- Live reload: Automatically updates when file changes on disk (perfect for watching AI edits)

### Clipboard Operations

| Key | Action |
|-----|--------|
| `y` | Copy entire file contents |
| `c` | Copy relative file path |

Files are copied to your system clipboard for use anywhere.

## File Operations

Complete file management without leaving the terminal. All operations include safety confirmations and validation.

### Creating Files

| Key | Action |
|-----|--------|
| `a` | Create new file |
| `A` | Create new directory |

Type the path (supports nested paths like `src/components/Button.tsx`). Missing parent directories are created automatically with confirmation.

### Modifying Files

| Key | Action |
|-----|--------|
| `r` | Rename file or directory |
| `m` | Move to different location |

The move operation includes path auto-completion showing up to 5 directory suggestions. All paths are validated to prevent moving files outside the project.

### Yank and Paste

| Key | Action |
|-----|--------|
| `y` | Yank (mark for copy) |
| `p` | Paste to current location |

Yank a file, navigate to the destination directory, and paste. Works for both files and entire directories.

### Deleting

| Key | Action |
|-----|--------|
| `D` | Delete with confirmation |

Confirmation modal shows the item being deleted and requires explicit approval.

### File Information

Press `I` for detailed file info modal:

- **Git status**: Tracked, modified, staged, untracked
- **File size**: Human-readable format
- **Modified**: Last modification timestamp
- **Permissions**: Unix permission bits
- **Last commit**: Most recent git commit affecting this file (when available)

## Advanced Features

### Mouse Support

Full mouse integration for faster workflows:

- **Click files/folders**: Select and preview or expand/collapse
- **Drag divider**: Resize panes to your preference
- **Scroll wheel**: Navigate tree or preview content
- **Click and drag in preview**: Multi-line text selection for copying

### Live File Watching

The preview pane watches the current file and automatically reloads when it changes on disk. This is particularly useful for:

- Watching AI agents modify code in real-time
- Monitoring log files
- Previewing generated files during build processes

### State Persistence

Your workspace state survives restarts. These are saved automatically:

- **Expanded directories**: Your tree structure stays intact
- **Cursor position**: Resume exactly where you left off
- **Scroll offsets**: Both tree and preview scroll positions
- **Active pane**: Tree or preview focus is remembered
- **Pane width**: Custom divider position is preserved

State is saved per-project based on working directory.

### System Integration

| Key | Action |
|-----|--------|
| `⌘+o` (or configured) | Open file in $EDITOR |
| `⌘+r` (or configured) | Reveal in Finder/Explorer |

Opens files in your configured editor (respects `$EDITOR` and `$VISUAL` environment variables). Defaults to vim if not set.

Reveal opens the system file manager with the file selected (macOS Finder, Windows Explorer, Linux file manager).

## Keyboard Reference

### Global (Available Anywhere)

| Key | Action |
|-----|--------|
| `ctrl+p` | Quick open (fuzzy file finder) |
| `ctrl+s` | Project search (ripgrep) |
| `\` | Toggle tree pane visibility |
| `tab` | Switch between tree and preview |

### Tree Pane

| Key | Action |
|-----|--------|
| `j/k` or `↓/↑` | Move down/up |
| `g` / `G` | Jump to top/bottom |
| `ctrl+d` / `ctrl+u` | Page down/up |
| `l` or `→` or `enter` | Expand directory or preview file |
| `h` or `←` | Collapse directory or go to parent |
| `/` | Filter tree by filename |
| `a` / `A` | Create new file/directory |
| `r` / `m` | Rename/move file |
| `D` | Delete (with confirmation) |
| `y` / `p` | Yank/paste file |
| `c` | Copy file path |
| `I` | Show file info modal |
| `H` | Toggle hidden/ignored files |

### Preview Pane

| Key | Action |
|-----|--------|
| `j/k` or `↓/↑` | Scroll down/up |
| `g` / `G` | Jump to top/bottom |
| `ctrl+d` / `ctrl+u` | Page down/up |
| `h` or `←` or `esc` | Return to tree |
| `?` | Search within file |
| `n` / `N` | Next/previous search match |
| `m` | Toggle markdown rendering |
| `y` | Copy file contents |
| `c` | Copy file path |

### Quick Open Modal

| Key | Action |
|-----|--------|
| type | Filter by filename (fuzzy) |
| `j/k` or `↓/↑` | Navigate results |
| `enter` | Open selected file |
| `esc` | Cancel |

### Project Search Modal

| Key | Action |
|-----|--------|
| type | Search query |
| `j/k` or `↓/↑` | Navigate results |
| `enter` | Open file at match line |
| `space` | Toggle file expansion |
| `esc` | Close search |

Supports regex mode, case sensitivity, and whole-word toggles (see hints in modal).

## Performance

The files plugin is built for speed, even on large codebases:

- **Quick open**: Caches 50,000 files in memory with 2-second scan timeout—handles massive monorepos
- **Project search**: Uses ripgrep (one of the fastest code search tools) with 30-second timeout
- **Preview rendering**: Syntax highlighting is cached until file changes
- **File watching**: Efficient fsnotify-based watching only for the previewed file
- **Lazy loading**: Tree nodes expand on demand, keeping memory usage low

**Limits:**
- Preview files truncated at 500KB (with warning shown)
- Max 10,000 lines displayed per file
- Max 1,000 search results shown per project search
- Max 50 quick open results displayed

## Common Workflows

### Finding and Editing a File

1. Press `ctrl+p` for quick open
2. Type part of the filename (fuzzy matching)
3. Press `enter` to preview
4. Press `⌘+o` (or configured) to open in your editor

### Searching Across the Project

1. Press `ctrl+s` to open project search
2. Type your search query (supports regex)
3. Navigate results with `j/k`
4. Press `enter` to jump to the match

### Refactoring Files

1. Navigate to file in tree with `j/k`
2. Press `r` to rename or `m` to move
3. Type the new name/path (auto-completion for move)
4. Press `enter` to confirm

### Copying File Structure

1. Navigate to file/directory
2. Press `y` to yank
3. Navigate to destination with `j/k` and `l/h`
4. Press `p` to paste

## Tips and Tricks

- **Use quick open for everything**: `ctrl+p` is faster than navigating the tree manually
- **Markdown previews**: Press `m` in any `.md` file to see rendered output
- **Watch AI changes**: Preview a file before starting an AI agent—watch it update in real-time
- **Multi-line copy**: Click and drag in the preview to select specific lines to copy
- **Regex search**: In project search, toggle regex mode to find patterns like `TODO|FIXME|HACK`
- **Path completion**: When moving files, start typing a path to see directory suggestions
- **Tree filtering**: Use `/` to quickly filter visible files without changing your tree expansion state

## Troubleshooting

**Quick open shows no results**
- Check the timeout wasn't exceeded (look for timeout message)
- Some files may be ignored by git patterns
- Try project search (`ctrl+s`) instead—it searches all files

**Preview shows "Binary file"**
- File contains null bytes in the first 512 bytes
- Use `⌘+o` to open in an external editor that supports binary files

**Syntax highlighting looks wrong**
- Highlighting is based on file extension
- Rename the file to use the correct extension
- Check that your terminal supports 256 colors

**File watching not working**
- File watching only works for the currently previewed file
- Some filesystems (network drives, some Docker volumes) don't support fsnotify
- Large files may take a moment to reload after changes

**Tree pane disappeared**
- Press `\` to toggle tree visibility
- You may have accidentally hidden it

## Integration with Other Plugins

The files plugin communicates with other plugins through messages:

- **Git plugin**: Can navigate to files in the file browser using `NavigateToFileMsg`
- **Editor integration**: Opens files at specific line numbers from search results
- **Focus switching**: Use `app.FocusPlugin("file-browser")` to switch to files plugin

See the plugin communication guide for details on integrating with the files plugin.
