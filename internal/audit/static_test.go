package audit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tu-usuario/cloud/internal/audit"
)

func TestAuditDirectory_Compliant(t *testing.T) {
	tempDir := t.TempDir()

	mainTf := `
variable "gcp_region" {
  default = "us-central1"
}

resource "google_service_account" "api_sa" {
  account_id = "api-sa"
}

resource "google_cloud_run_v2_service" "api" {
  name     = "api"
  location = var.gcp_region
  template {
    service_account = google_service_account.api_sa.email
    containers {
      image = "gcr.io/demo/api"
      resources {
        limits = {
          memory = "256Mi"
          cpu    = "1000m"
        }
      }
    }
    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }
  }
}

resource "google_compute_instance" "vm" {
  name         = "vm"
  machine_type = "e2-micro"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
      size  = 30
      type  = "pd-standard"
    }
  }
}

resource "google_storage_bucket" "bucket" {
  name                        = "test-bucket"
  location                    = "us-central1"
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(mainTf), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	report, err := audit.AuditDirectory(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.Passed {
		t.Errorf("expected audit to PASS, but failed with: %+v", report.Findings)
	}
	if report.Summary.Fail != 0 {
		t.Errorf("expected 0 fails, got %d", report.Summary.Fail)
	}
}

func TestAuditDirectory_NonCompliant(t *testing.T) {
	tempDir := t.TempDir()

	badTf := `
variable "gcp_region" {
  default = "europe-west1"
}

resource "google_compute_instance" "costly_vm" {
  name         = "costly-vm"
  machine_type = "n1-standard-4"
  zone         = "europe-west1-b"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
      size  = 100
      type  = "pd-ssd"
    }
  }
}

resource "google_cloud_run_v2_service" "bad_run" {
  name     = "bad-run"
  location = "europe-west1"
  template {
    containers {
      image = "gcr.io/demo/bad"
      resources {
        limits = {
          memory = "2Gi"
        }
      }
    }
    scaling {
      min_instance_count = 2
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "main.tf"), []byte(badTf), 0644); err != nil {
		t.Fatalf("failed to write bad test file: %v", err)
	}

	report, err := audit.AuditDirectory(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Passed {
		t.Errorf("expected audit to FAIL on costly resources")
	}

	if report.Summary.Fail < 5 {
		t.Errorf("expected at least 5 failures, got %d", report.Summary.Fail)
	}
}

func TestAuditDirectory_NoTfFiles(t *testing.T) {
	tempDir := t.TempDir()
	_, err := audit.AuditDirectory(tempDir)
	if err == nil {
		t.Errorf("expected error when no .tf files exist, got nil")
	}
}
