package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type addSkillOptions struct {
	baseDir    string
	binaryName string
	force      bool
}

func newAddSkillCommand() *cobra.Command {
	options := addSkillOptions{
		baseDir:    ".",
		binaryName: "crawldown",
	}

	addSkillCmd := &cobra.Command{
		Use:           "add-skill <name>",
		Short:         "Create an agent skill scaffold for crawldown in the current directory",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddSkill(options, args[0])
		},
	}

	flags := addSkillCmd.Flags()
	flags.StringVar(&options.baseDir, "base-dir", ".", "Base directory where the .agents/skills structure will be created")
	flags.StringVar(&options.binaryName, "binary", "crawldown", "Binary name to use in the generated instructions")
	flags.BoolVar(&options.force, "force", false, "Overwrite an existing SKILL.md file")

	return addSkillCmd
}

func runAddSkill(options addSkillOptions, name string) error {
	skillName := strings.TrimSpace(name)
	if skillName == "" {
		return fmt.Errorf("skill name cannot be empty")
	}

	skillDir := filepath.Join(options.baseDir, ".agents", "skills", skillName)
	skillFile := filepath.Join(skillDir, "SKILL.md")

	if err := os.MkdirAll(skillDir, 0o750); err != nil {
		return fmt.Errorf("create skill directory: %w", err)
	}

	if !options.force {
		if _, err := os.Stat(skillFile); err == nil {
			return fmt.Errorf("skill already exists at %s", skillFile)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check skill file: %w", err)
		}
	}

	content := buildSkillContent(skillName, options.binaryName)
	if err := os.WriteFile(skillFile, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write skill file: %w", err)
	}

	printStdout("Created skill scaffold at %s\n", skillFile)

	return nil
}

func buildSkillContent(skillName, binaryName string) string {
	command := strings.TrimSpace(binaryName)
	if command == "" {
		command = "crawldown"
	}

	return fmt.Sprintf("# %s\n\n"+
		"Use this skill when you need to download a single page or crawl a website into a specific directory with CrawlDown.\n\n"+
		"## Goal\n\n"+
		"Use the CrawlDown CLI to save website content as Markdown files in a user-specified output directory.\n\n"+
		"## Before you start\n\n"+
		"- Confirm the source URL.\n"+
		"- Confirm the output directory.\n"+
		"- Decide whether the user wants a full crawl or a single page download.\n"+
		"- Prefer the installed binary %q. If it is not available but you are inside the repository, fall back to go run ./cmd.\n\n"+
		"## Recommended workflow\n\n"+
		"1. Ask for the URL and the destination directory if they are missing.\n"+
		"2. Create the output directory before running the download command.\n"+
		"3. For a full crawl, use the get command with the positional URL.\n"+
		"4. For a single page, use the --single flag.\n"+
		"5. Add optional flags only when the user asks for them or the site requires them.\n"+
		"6. After the command completes, summarize the generated files or the output directory contents.\n\n"+
		"## Commands\n\n"+
		"### Full crawl\n\n"+
		"    mkdir -p <output-dir>\n"+
		"    %s get -o <output-dir> <url>\n\n"+
		"### Single page download\n\n"+
		"    mkdir -p <output-dir>\n"+
		"    %s get -o <output-dir> --single <page-url>\n\n"+
		"### Crawl with depth and excluded paths\n\n"+
		"    mkdir -p <output-dir>\n"+
		"    %s get -o <output-dir> --depth 3 --exclude /admin --exclude /private <url>\n\n"+
		"### Crawl with timeout, delay, robots, and external links options\n\n"+
		"    mkdir -p <output-dir>\n"+
		"    %s get -o <output-dir> --timeout 30 --delay 2 --ignore-robots-txt --follow-external-links <url>\n\n"+
		"## Repository fallback\n\n"+
		"If the binary is not installed but you are in the repository root, use:\n\n"+
		"    mkdir -p <output-dir>\n"+
		"    go run ./cmd get -o <output-dir> <url>\n\n"+
		"## Notes\n\n"+
		"- The root command behaves like get, so %s -o <output-dir> <url> is also valid.\n"+
		"- Output is written as Markdown files under the chosen directory.\n"+
		"- Use --user-agent if the target site requires a custom user agent.\n"+
		"- Keep the command scoped to the requested site and destination directory.\n",
		skillName,
		command,
		command,
		command,
		command,
		command,
		command,
	)
}
