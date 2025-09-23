# git-regex-search

A powerful CLI tool to search for regex patterns across multiple Git branches in a repository. This tool automatically handles branch switching, stashing changes, and provides colorized output with detailed progress information.

## Features

- 🔍 **Cross-branch regex search**: Search for patterns across all remote branches or specific branches
- 🚀 **Smart tool detection**: Uses `ripgrep` (rg) for faster searches when available, falls back to `grep`
- 💾 **Safe operations**: Automatically stashes uncommitted changes and restores them after search
- 🎨 **Colorized output**: Branch names and line numbers are highlighted for better readability
- 📊 **Verbose progress**: Real-time updates showing which branch is being searched and match counts
- 🌐 **Remote branch support**: Automatically fetches and searches all remote branches
- ⚡ **Selective branch search**: Option to search only specific branches

## Installation

### Prerequisites

- Go 1.24.4 or later
- Git
- `ripgrep` (optional, for faster searches) - install via:
  - macOS: `brew install ripgrep`
  - Ubuntu/Debian: `apt install ripgrep`
  - Other systems: See [ripgrep installation guide](https://github.com/BurntSushi/ripgrep#installation)

### Build from source

```bash
git clone <repository-url>
cd git-regex-search
go mod tidy
go build -o git-regex-search
```

## Usage

### Basic Usage

```bash
./git-regex-search --repo /path/to/git/repo --regex "your-pattern"
```

### Search specific branches

```bash
./git-regex-search --repo /path/to/git/repo --regex "your-pattern" --branches "main,develop,feature-branch"
```

### Command Line Options

| Flag | Description | Required |
|------|-------------|----------|
| `--repo` | Path to the git repository | ✅ Yes |
| `--regex` | Regular expression pattern to search for | ✅ Yes |
| `--branches` | Comma-separated list of branches to search | ❌ No (defaults to all remote branches) |

## Examples

### Search for a function across all branches
```bash
./git-regex-search --repo ~/my-project --regex "func.*handleRequest"
```

### Search for TODO comments in specific branches
```bash
./git-regex-search --repo ~/my-project --regex "TODO|FIXME" --branches "main,develop"
```

### Search for API endpoints
```bash
./git-regex-search --repo ~/my-project --regex "\/api\/v[0-9]+\/"
```

## Output Format

The tool provides verbose output showing:

1. **Setup information**: Repository path, search pattern, current branch
2. **Progress updates**: Which branch is currently being searched
3. **Match results**: Colorized output in the format:
   ```
   branch-name:path/to/file:line-number content-of-matching-line
   ```
4. **Summary**: Number of matches found per branch
5. **Cleanup**: Restoration of original branch and stashed changes

### Example Output

```
📁 Repository: /home/user/my-project
🔍 Search pattern: handleRequest
🌿 Current branch: main

💾 Stashing uncommitted changes...
🌐 Fetching remote branches...
Searching across 3 branches...

🔍 Searching branch: main
✅ Found 2 matches in main
main:src/handlers/api.go:42 func handleRequest(w http.ResponseWriter, r *http.Request) {
main:src/handlers/auth.go:15 func handleRequestAuth(token string) bool {

🔍 Searching branch: develop
❌ No matches found in develop

🔍 Searching branch: feature-api
✅ Found 1 matches in feature-api
feature-api:src/new_handler.go:23 func handleRequestV2(ctx context.Context) error {

🔄 Restoring original branch: main
📤 Restoring stashed changes...
✨ Search completed!
```

## How It Works

1. **Validation**: Checks if the provided path is a valid Git repository
2. **Safety**: Stashes any uncommitted changes to prevent data loss
3. **Preparation**: Fetches all remote branches to ensure up-to-date search
4. **Search**: For each branch:
   - Checks out the branch
   - Runs regex search using `rg` (preferred) or `grep` (fallback)
   - Collects and displays matches with colored output
5. **Cleanup**: Restores the original branch and pops stashed changes

## Dependencies

The project uses the following Go modules:

- `github.com/urfave/cli/v2` - CLI framework
- `github.com/fatih/color` - Terminal color output

## Performance

- **With ripgrep**: Significantly faster searches, especially on large repositories
- **With grep**: Reliable fallback that works on all Unix-like systems
- **Smart detection**: Automatically chooses the best available tool

## Safety Features

- ✅ Stashes uncommitted changes before starting
- ✅ Restores original branch after completion
- ✅ Validates regex patterns before execution
- ✅ Handles interrupted operations gracefully

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## Troubleshooting

### Common Issues

**"not a git repository" error**
- Ensure the path points to a directory containing a `.git` folder

**"invalid regex" error**
- Test your regex pattern with a Go regex tester or simpler pattern

**No matches found**
- Verify the pattern exists in the repository
- Try a simpler regex pattern
- Check if you're searching the correct branches

**Permission denied**
- Ensure you have read access to the repository
- Check if the repository requires authentication for remote operations