package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// version is injected at build time via -ldflags "-X main.version=<version>"
var version = "dev"

func runGitCmd(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func grepRepo(repoPath, pattern string, includeGlobs, excludeGlobs []string) ([]string, error) {
	var cmd *exec.Cmd
	if commandExists("rg") {
		args := []string{"-n", "-uu", "--pcre2"}
		for _, g := range includeGlobs {
			if strings.TrimSpace(g) == "" {
				continue
			}
			args = append(args, "--glob", g)
		}
		for _, g := range excludeGlobs {
			if strings.TrimSpace(g) == "" {
				continue
			}
			args = append(args, "--glob", "!"+g)
		}
		args = append(args, pattern)
		cmd = exec.Command("rg", args...)
	} else {
		cmd = exec.Command("grep", "-rnE", pattern, ".")
	}

	cmd.Dir = repoPath
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	if err != nil && out.Len() == 0 {
		// Both rg and grep return non-zero if no matches are found
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(out.String()), "\n"), nil
}

func main() {
	app := &cli.App{
		Name:    "git-regex-search",
		Usage:   "Search for regex matches across branches in a git repository",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "repo",
				Usage:    "Path to the git repository",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "regex",
				Usage:    "Regular expression to search for",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "branches",
				Usage: "Comma-separated list of branches to search (optional)",
			},
			&cli.StringSliceFlag{
				Name:  "include-glob",
				Usage: "Include only files/dirs matching glob (ripgrep only). Repeatable.",
			},
			&cli.StringSliceFlag{
				Name:  "exclude-glob",
				Usage: "Exclude files/dirs matching glob (ripgrep only). Repeatable.",
			},
		},
		Action: func(c *cli.Context) error {
			repoPath, err := filepath.Abs(c.String("repo"))
			if err != nil {
				return fmt.Errorf("invalid repo path: %v", err)
			}

			regex := c.String("regex")
			branchesArg := c.String("branches")
			includeGlobs := c.StringSlice("include-glob")
			excludeGlobs := c.StringSlice("exclude-glob")

			// Ensure repo exists
			if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
				return fmt.Errorf("not a git repository: %s", repoPath)
			}

			// Get current branch
			currentBranch, err := runGitCmd(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
			if err != nil {
				return fmt.Errorf("failed to get current branch: %v", err)
			}

			fmt.Printf("📁 Repository: %s\n", repoPath)
			fmt.Printf("🔍 Search pattern: %s\n", regex)
			fmt.Printf("🌿 Current branch: %s\n", currentBranch)
			fmt.Println()

			// Warn if include/exclude globs provided but ripgrep is not available
			if (len(includeGlobs) > 0 || len(excludeGlobs) > 0) && !commandExists("rg") {
				fmt.Println("⚠️  Warning: include/exclude glob options require 'rg' (ripgrep). Options will be ignored because 'rg' was not found in PATH.")
			}

			// Stash changes
			fmt.Println("💾 Stashing uncommitted changes...")
			_, _ = runGitCmd(repoPath, "stash", "push", "-u", "-m", "git-regex-search-temp-stash")

			// Pull remote branches list
			fmt.Println("🌐 Fetching remote branches...")
			_, _ = runGitCmd(repoPath, "fetch", "--all", "--quiet")

			var branches []string
			if branchesArg != "" {
				branches = strings.Split(branchesArg, ",")
			} else {
				branchList, err := runGitCmd(repoPath, "branch", "-r")
				if err != nil {
					return fmt.Errorf("failed to list remote branches: %v", err)
				}
				for _, b := range strings.Split(branchList, "\n") {
					b = strings.TrimSpace(b)
					if b != "" && !strings.Contains(b, "->") {
						branches = append(branches, strings.TrimPrefix(b, "origin/"))
					}
				}
			}

			// Regex validation
			if _, err := regexp.Compile(regex); err != nil {
				return fmt.Errorf("invalid regex: %v", err)
			}

			branchColor := color.New(color.FgGreen).SprintFunc()
			lineNumColor := color.New(color.FgYellow).SprintFunc()

			fmt.Printf("Searching across %d branches...\n\n", len(branches))

			for _, branch := range branches {
				branch = strings.TrimSpace(branch)
				if branch == "" {
					continue
				}

				fmt.Printf("\n🔍 Searching branch: %s\n", branchColor(branch))
				_, _ = runGitCmd(repoPath, "checkout", branch)
				fmt.Printf("📥 Pulling latest changes for %s...\n", branchColor(branch))
				_, _ = runGitCmd(repoPath, "pull", "origin", branch)
				matches, err := grepRepo(repoPath, regex, includeGlobs, excludeGlobs)
				if err != nil {
					return fmt.Errorf("search failed on branch %s: %v", branch, err)
				}

				if len(matches) > 0 {
					fmt.Printf("✅ Found %d matches in %s\n", len(matches), branchColor(branch))
				} else {
					fmt.Printf("❌ No matches found in %s\n", branchColor(branch))
				}

				for _, match := range matches {
					parts := strings.SplitN(match, ":", 3)
					if len(parts) < 3 {
						continue
					}
					file := parts[0]
					lineNum := parts[1]
					lineText := parts[2]

					fmt.Printf("%s:%s%s %s\n",
						branchColor(branch),
						file, lineNumColor(":"+lineNum),
						lineText,
					)
				}
			}

			fmt.Println()
			// Restore original branch and stash
			fmt.Printf("🔄 Restoring original branch: %s\n", currentBranch)
			_, _ = runGitCmd(repoPath, "checkout", currentBranch)
			fmt.Println("📤 Restoring stashed changes...")
			_, _ = runGitCmd(repoPath, "stash", "pop")

			fmt.Println("✨ Search completed!")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
