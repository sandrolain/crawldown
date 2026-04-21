package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAddSkillCreatesSkillScaffold(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	options := addSkillOptions{
		baseDir:    baseDir,
		binaryName: "crawldown",
	}

	if err := runAddSkill(options, "site-fetch"); err != nil {
		t.Fatalf("runAddSkill returned error: %v", err)
	}

	skillPath := filepath.Join(baseDir, ".agents", "skills", "site-fetch", "SKILL.md")
	//nolint:gosec // The path is created under t.TempDir and controlled by the test.
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("reading generated skill: %v", err)
	}

	skillContent := string(content)
	if !strings.Contains(skillContent, "crawldown get -o <output-dir> <url>") {
		t.Fatalf("generated skill does not contain crawl command: %s", skillContent)
	}

	if !strings.Contains(skillContent, "go run ./cmd get -o <output-dir> <url>") {
		t.Fatalf("generated skill does not contain repository fallback: %s", skillContent)
	}
}

func TestRunAddSkillFailsWhenSkillExistsWithoutForce(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	options := addSkillOptions{baseDir: baseDir}

	if err := runAddSkill(options, "site-fetch"); err != nil {
		t.Fatalf("initial runAddSkill returned error: %v", err)
	}

	err := runAddSkill(options, "site-fetch")
	if err == nil {
		t.Fatal("expected an error when skill already exists")
	}
}

func TestRunAddSkillOverwritesWhenForceIsEnabled(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	options := addSkillOptions{baseDir: baseDir}

	if err := runAddSkill(options, "site-fetch"); err != nil {
		t.Fatalf("initial runAddSkill returned error: %v", err)
	}

	forceOptions := addSkillOptions{
		baseDir:    baseDir,
		binaryName: "crawler",
		force:      true,
	}

	if err := runAddSkill(forceOptions, "site-fetch"); err != nil {
		t.Fatalf("force runAddSkill returned error: %v", err)
	}

	skillPath := filepath.Join(baseDir, ".agents", "skills", "site-fetch", "SKILL.md")
	//nolint:gosec // The path is created under t.TempDir and controlled by the test.
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("reading overwritten skill: %v", err)
	}

	if !strings.Contains(string(content), "crawler get -o <output-dir> <url>") {
		t.Fatalf("expected overwritten skill to use custom binary name, got: %s", string(content))
	}
}
