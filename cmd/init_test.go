package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tu-usuario/cloud/cmd"
	"github.com/tu-usuario/cloud/internal/generator"
)

func writeSampleConfig(t *testing.T, dir string) string {
	t.Helper()
	cfg := generator.ProjectConfig{
		ProjectName:          "ci-test-app",
		GCPProjectID:         "ci-project-999",
		Region:               "us-central1",
		IncludeStorage:       true,
		IncludeComputeEngine: false,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(dir, "test-config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	return configPath
}

func TestInitCmd_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	configPath := writeSampleConfig(t, tempDir)
	outDir := filepath.Join(tempDir, "should-not-exist")

	rootCmd := cmd.NewInitCmd()
	rootCmd.SetArgs([]string{
		"--config", configPath,
		"--output-dir", outDir,
		"--dry-run",
	})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error with --dry-run, got: %v", err)
	}

	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		t.Errorf("expected directory %s to NOT exist after --dry-run", outDir)
	}
}

func TestInitCmd_CustomOutputDir(t *testing.T) {
	tempDir := t.TempDir()
	configPath := writeSampleConfig(t, tempDir)
	outDir := filepath.Join(tempDir, "custom-infra")

	rootCmd := cmd.NewInitCmd()
	rootCmd.SetArgs([]string{
		"--config", configPath,
		"--output-dir", outDir,
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected successful generation, got: %v", err)
	}

	mainTf := filepath.Join(outDir, "main.tf")
	if _, err := os.Stat(mainTf); os.IsNotExist(err) {
		t.Errorf("expected generated file %s to exist", mainTf)
	}
}

func TestInitCmd_FormatJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := writeSampleConfig(t, tempDir)

	rootCmd := cmd.NewInitCmd()
	rootCmd.SetArgs([]string{
		"--config", configPath,
		"--format", "json",
		"--dry-run",
	})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error with --format json --dry-run, got: %v", err)
	}
}

func TestInitCmd_InvalidFormat(t *testing.T) {
	tempDir := t.TempDir()
	configPath := writeSampleConfig(t, tempDir)

	rootCmd := cmd.NewInitCmd()
	rootCmd.SetArgs([]string{
		"--config", configPath,
		"--format", "yaml",
		"--dry-run",
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected error with unsupported format, got nil")
	}
}

func TestInitCmd_MissingConfigFile(t *testing.T) {
	rootCmd := cmd.NewInitCmd()
	rootCmd.SetArgs([]string{
		"--config", "/non/existent/path.json",
		"--dry-run",
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected error with non-existent config file, got nil")
	}
}
