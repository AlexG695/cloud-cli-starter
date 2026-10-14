package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/tu-usuario/cloud/cmd"
)

func writeCompliantTerraform(t *testing.T, dir string) {
	t.Helper()
	content := `
variable "gcp_region" {
  default = "us-central1"
}

resource "google_service_account" "sa" {
  account_id = "test-sa"
}

resource "google_compute_instance" "vm" {
  name         = "free-vm"
  machine_type = "e2-micro"
  zone         = "us-central1-b"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
      size  = 30
      type  = "pd-standard"
    }
  }

  network_interface {
    network = "default"
  }

  service_account {
    email  = google_service_account.sa.email
    scopes = ["cloud-platform"]
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test terraform: %v", err)
	}
}

func writeBadTerraform(t *testing.T, dir string) {
	t.Helper()
	content := `
resource "google_compute_instance" "costly_vm" {
  name         = "costly-vm"
  machine_type = "c2-standard-16"
  zone         = "europe-west1-b"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
      size  = 200
      type  = "pd-ssd"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write bad terraform: %v", err)
	}
}

func TestAuditCmd_Success(t *testing.T) {
	tempDir := t.TempDir()
	writeCompliantTerraform(t, tempDir)

	rootCmd := cmd.NewInitCmd()
	rootCmd.SetArgs([]string{"audit", tempDir})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected audit to pass, got error: %v", err)
	}
}

func TestAuditCmd_FormatJSON(t *testing.T) {
	tempDir := t.TempDir()
	writeCompliantTerraform(t, tempDir)

	rootCmd := cmd.NewInitCmd()
	rootCmd.SetArgs([]string{"audit", tempDir, "--format", "json"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected audit with --format json to pass, got error: %v", err)
	}
}

func TestAuditCmd_StrictFail(t *testing.T) {
	tempDir := t.TempDir()
	writeBadTerraform(t, tempDir)

	rootCmd := cmd.NewInitCmd()
	rootCmd.SetArgs([]string{"audit", tempDir, "--strict"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected audit to fail under --strict on non-compliant resources, got nil")
	}
}

func TestAuditCmd_NonExistentDir(t *testing.T) {
	rootCmd := cmd.NewInitCmd()
	rootCmd.SetArgs([]string{"audit", "/path/that/does/not/exist"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected error on non-existent directory, got nil")
	}
}
