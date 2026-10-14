package generator_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tu-usuario/cloud/internal/generator"
)

func TestGenerateProject(t *testing.T) {
	tempDir := t.TempDir()
	infraDir := filepath.Join(tempDir, "infra")

	config := generator.ProjectConfig{
		ProjectName:  "my-test-app",
		GCPProjectID: "test-project-123",
		Region:       "us-central1",
		Services: []generator.CloudRunService{
			{
				Name:         "api",
				Memory:       "256Mi",
				CPU:          "1000m",
				IsPublic:     true,
				EnablePubSub: true,
			},
		},
		IncludeStorage:         true,
		IncludeFirestore:       true,
		IncludeAuth:            false,
		EnableIdentityPlatform: false,
		IncludeSecrets:         true,
		IncludeLocalEmulators:  true,
		EnableGemmaAssistant:   false,
	}

	err := generator.GenerateProject(config, infraDir)
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	expectedFiles := []string{
		"providers.tf",
		"main.tf",
		"compute.tf",
		"modules.tf",
		"variables.tf",
		"outputs.tf",
		"Makefile",
		"docker-compose.yml",
		"firebase.json",
		"firestore.rules",
		"storage.rules",
	}

	for _, relPath := range expectedFiles {
		filePath := filepath.Join(infraDir, relPath)
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Generated file '%s' not found: %v", relPath, err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("Generated file '%s' is empty", relPath)
		}
	}

	// Verify API enablement in main.tf
	mainContent, _ := os.ReadFile(filepath.Join(infraDir, "main.tf"))
	if !strings.Contains(string(mainContent), "google_project_service") {
		t.Errorf("main.tf does not contain google_project_service resource")
	}
	if !strings.Contains(string(mainContent), "cloudresourcemanager.googleapis.com") {
		t.Errorf("main.tf does not enable cloudresourcemanager.googleapis.com")
	}
	if !strings.Contains(string(mainContent), "iam.googleapis.com") {
		t.Errorf("main.tf does not enable iam.googleapis.com")
	}
	if !strings.Contains(string(mainContent), "run.googleapis.com") {
		t.Errorf("main.tf does not enable run.googleapis.com when services exist")
	}
	if !strings.Contains(string(mainContent), "pubsub.googleapis.com") {
		t.Errorf("main.tf does not enable pubsub.googleapis.com when a service requires it")
	}
	if !strings.Contains(string(mainContent), "storage.googleapis.com") {
		t.Errorf("main.tf does not enable storage.googleapis.com")
	}
	if !strings.Contains(string(mainContent), "firestore.googleapis.com") {
		t.Errorf("main.tf does not enable firestore.googleapis.com")
	}
	if !strings.Contains(string(mainContent), "secretmanager.googleapis.com") {
		t.Errorf("main.tf does not enable secretmanager.googleapis.com")
	}
	if strings.Contains(string(mainContent), "identitytoolkit.googleapis.com") {
		t.Errorf("main.tf should not enable identitytoolkit.googleapis.com if Auth is disabled")
	}

	// Verify Zero-Trust and Service Accounts in compute.tf
	computeContent, _ := os.ReadFile(filepath.Join(infraDir, "compute.tf"))
	if !strings.Contains(string(computeContent), "api_sa") {
		t.Errorf("compute.tf does not contain dedicated Service Account api_sa")
	}
	if !strings.Contains(string(computeContent), "service_account = google_service_account.api_sa.email") {
		t.Errorf("compute.tf does not assign dedicated Service Account to Cloud Run runtime")
	}
	if !strings.Contains(string(computeContent), "roles/datastore.user") {
		t.Errorf("compute.tf does not assign least privilege role roles/datastore.user")
	}
	if !strings.Contains(string(computeContent), "INGRESS_TRAFFIC_ALL") {
		t.Errorf("compute.tf does not contain ingress configuration for public service")
	}

	// Verify data protection in modules.tf
	modulesContent, _ := os.ReadFile(filepath.Join(infraDir, "modules.tf"))
	if !strings.Contains(string(modulesContent), "public_access_prevention    = \"enforced\"") {
		t.Errorf("modules.tf does not contain public_access_prevention = enforced")
	}
	if !strings.Contains(string(modulesContent), "versioning") {
		t.Errorf("modules.tf does not contain versioning block for Cloud Storage")
	}

	// Verify outputs.tf
	outputsContent, _ := os.ReadFile(filepath.Join(infraDir, "outputs.tf"))
	if !strings.Contains(string(outputsContent), "cloud_run_api_uri") {
		t.Errorf("outputs.tf does not export cloud_run_api_uri")
	}
	if !strings.Contains(string(outputsContent), "storage_bucket_name") {
		t.Errorf("outputs.tf does not export storage_bucket_name")
	}

	variablesContent, _ := os.ReadFile(filepath.Join(infraDir, "variables.tf"))
	if !strings.Contains(string(variablesContent), "test-project-123") {
		t.Errorf("variables.tf does not contain expected GCP Project ID")
	}

	makefileContent, _ := os.ReadFile(filepath.Join(infraDir, "Makefile"))
	if !strings.Contains(string(makefileContent), "terraform") {
		t.Errorf("Makefile does not contain terraform references")
	}

	dockerContent, _ := os.ReadFile(filepath.Join(infraDir, "docker-compose.yml"))
	if !strings.Contains(string(dockerContent), "andreysenov/firebase-tools:latest") {
		t.Errorf("docker-compose.yml does not contain image andreysenov/firebase-tools:latest")
	}
	if !strings.Contains(string(dockerContent), "4000:4000") {
		t.Errorf("docker-compose.yml does not expose port 4000:4000 for dashboard")
	}
	if !strings.Contains(string(dockerContent), "./firebase.json:/firebase.json:z") {
		t.Errorf("docker-compose.yml does not mount firebase.json with :z")
	}
	if !strings.Contains(string(dockerContent), "FIREBASE_LOG_DIR=/tmp") {
		t.Errorf("docker-compose.yml does not define FIREBASE_LOG_DIR=/tmp")
	}
	if !strings.Contains(string(dockerContent), "firebase emulators:start --config=/firebase.json --project=local-dev") {
		t.Errorf("docker-compose.yml does not contain expected startup command")
	}
	if strings.Contains(string(dockerContent), "9099:9099") {
		t.Errorf("docker-compose.yml should not include port 9099 if Identity Platform is inactive")
	}

	firebaseContent, _ := os.ReadFile(filepath.Join(infraDir, "firebase.json"))
	if !strings.Contains(string(firebaseContent), "firestore.rules") ||
		!strings.Contains(string(firebaseContent), "storage.rules") ||
		!strings.Contains(string(firebaseContent), "singleProjectMode") {
		t.Errorf("firebase.json does not contain rules configuration or singleProjectMode: %s", string(firebaseContent))
	}
	if strings.Contains(string(firebaseContent), "\"auth\"") {
		t.Errorf("firebase.json should not include auth emulator if Identity Platform is inactive")
	}

	firestoreRules, _ := os.ReadFile(filepath.Join(infraDir, "firestore.rules"))
	if !strings.Contains(string(firestoreRules), "allow read, write: if true;") {
		t.Errorf("firestore.rules does not contain development permissive rules")
	}

	storageRules, _ := os.ReadFile(filepath.Join(infraDir, "storage.rules"))
	if !strings.Contains(string(storageRules), "allow read, write: if true;") {
		t.Errorf("storage.rules does not contain development permissive rules")
	}

	// Validate canonical Terraform formatting
	diffCmd := exec.Command("terraform", "fmt", "-diff", infraDir)
	if out, err := diffCmd.CombinedOutput(); err != nil || len(out) > 0 {
		t.Errorf("terraform fmt failed on generated files:\n%s", string(out))
	}
}

func TestGenerateProject_WithIdentityPlatform(t *testing.T) {
	tempDir := t.TempDir()
	infraDir := filepath.Join(tempDir, "infra")

	config := generator.ProjectConfig{
		ProjectName:            "auth-test-app",
		GCPProjectID:           "test-project-456",
		Region:                 "us-central1",
		IncludeStorage:         true,
		IncludeFirestore:       true,
		IncludeAuth:            true,
		EnableIdentityPlatform: true,
		IncludeLocalEmulators:  true,
	}

	err := generator.GenerateProject(config, infraDir)
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	dockerContent, _ := os.ReadFile(filepath.Join(infraDir, "docker-compose.yml"))
	if !strings.Contains(string(dockerContent), "9099:9099") {
		t.Errorf("docker-compose.yml must include port 9099:9099 when Identity Platform is active")
	}

	firebaseContent, _ := os.ReadFile(filepath.Join(infraDir, "firebase.json"))
	if !strings.Contains(string(firebaseContent), "\"auth\"") || !strings.Contains(string(firebaseContent), "9099") {
		t.Errorf("firebase.json must include auth emulator with port 9099 when Identity Platform is active")
	}

	mainContent, _ := os.ReadFile(filepath.Join(infraDir, "main.tf"))
	if !strings.Contains(string(mainContent), "identitytoolkit.googleapis.com") {
		t.Errorf("main.tf must enable identitytoolkit.googleapis.com when Identity Platform is active")
	}
}

func TestGenerateProject_WithComputeEngine_Private(t *testing.T) {
	tempDir := t.TempDir()
	infraDir := filepath.Join(tempDir, "infra")

	config := generator.ProjectConfig{
		ProjectName:          "vm-private-app",
		GCPProjectID:         "test-vm-123",
		Region:               "us-central1",
		IncludeComputeEngine: true,
		VMPublicIP:           false,
		IncludeStorage:       true,
		IncludeFirestore:     true,
		IncludeSecrets:       true,
	}

	err := generator.GenerateProject(config, infraDir)
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	mainContent, _ := os.ReadFile(filepath.Join(infraDir, "main.tf"))
	if !strings.Contains(string(mainContent), "compute.googleapis.com") {
		t.Errorf("main.tf must enable compute.googleapis.com when Compute Engine is active")
	}

	variablesContent, _ := os.ReadFile(filepath.Join(infraDir, "variables.tf"))
	if !strings.Contains(string(variablesContent), "variable \"gcp_zone\"") {
		t.Errorf("variables.tf must define gcp_zone when Compute Engine is active")
	}

	computeContent, _ := os.ReadFile(filepath.Join(infraDir, "compute.tf"))
	if !strings.Contains(string(computeContent), "resource \"google_compute_instance\" \"vm_instance\"") {
		t.Errorf("compute.tf must define vm_instance")
	}
	if !strings.Contains(string(computeContent), "machine_type = \"e2-micro\"") {
		t.Errorf("compute.tf must enforce machine_type = e2-micro for Free Tier")
	}
	if !strings.Contains(string(computeContent), "type  = \"pd-standard\"") {
		t.Errorf("compute.tf must enforce pd-standard disk type for Free Tier")
	}
	if !strings.Contains(string(computeContent), "size  = 30") {
		t.Errorf("compute.tf must enforce disk size = 30 for Free Tier")
	}
	if !strings.Contains(string(computeContent), "image = \"debian-cloud/debian-12\"") {
		t.Errorf("compute.tf must use standard Linux image debian-12")
	}
	if !strings.Contains(string(computeContent), "resource \"google_service_account\" \"vm_sa\"") {
		t.Errorf("compute.tf must define dedicated Service Account vm_sa")
	}
	if strings.Contains(string(computeContent), "access_config") {
		t.Errorf("compute.tf must NOT contain access_config when VMPublicIP is false")
	}

	outputsContent, _ := os.ReadFile(filepath.Join(infraDir, "outputs.tf"))
	if !strings.Contains(string(outputsContent), "output \"vm_instance_name\"") {
		t.Errorf("outputs.tf must export vm_instance_name")
	}
	if !strings.Contains(string(outputsContent), "output \"vm_private_ip\"") {
		t.Errorf("outputs.tf must export vm_private_ip when VMPublicIP is false")
	}
	if strings.Contains(string(outputsContent), "output \"vm_public_ip\"") {
		t.Errorf("outputs.tf must NOT export vm_public_ip when VMPublicIP is false")
	}

	// Validate canonical Terraform formatting
	diffCmd := exec.Command("terraform", "fmt", "-diff", infraDir)
	if out, err := diffCmd.CombinedOutput(); err != nil || len(out) > 0 {
		t.Errorf("terraform fmt failed on private VM generated files:\n%s", string(out))
	}
}

func TestGenerateProject_WithComputeEngine_Public(t *testing.T) {
	tempDir := t.TempDir()
	infraDir := filepath.Join(tempDir, "infra")

	config := generator.ProjectConfig{
		ProjectName:          "vm-public-app",
		GCPProjectID:         "test-vm-456",
		Region:               "us-central1",
		IncludeComputeEngine: true,
		VMPublicIP:           true,
	}

	err := generator.GenerateProject(config, infraDir)
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	computeContent, _ := os.ReadFile(filepath.Join(infraDir, "compute.tf"))
	if !strings.Contains(string(computeContent), "access_config") {
		t.Errorf("compute.tf must contain access_config when VMPublicIP is true")
	}

	outputsContent, _ := os.ReadFile(filepath.Join(infraDir, "outputs.tf"))
	if !strings.Contains(string(outputsContent), "output \"vm_public_ip\"") {
		t.Errorf("outputs.tf must export vm_public_ip when VMPublicIP is true")
	}
	if strings.Contains(string(outputsContent), "output \"vm_private_ip\"") {
		t.Errorf("outputs.tf must NOT export vm_private_ip when VMPublicIP is true")
	}

	// Validate canonical Terraform formatting
	diffCmd := exec.Command("terraform", "fmt", "-diff", infraDir)
	if out, err := diffCmd.CombinedOutput(); err != nil || len(out) > 0 {
		t.Errorf("terraform fmt failed on public VM generated files:\n%s", string(out))
	}
}

func TestGenerateProject_WithBigQuery(t *testing.T) {
	tempDir := t.TempDir()
	infraDir := filepath.Join(tempDir, "infra")

	config := generator.ProjectConfig{
		ProjectName:     "bq-test-app",
		GCPProjectID:    "test-bq-123",
		Region:          "us-central1",
		IncludeBigQuery: true,
	}

	err := generator.GenerateProject(config, infraDir)
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	mainContent, _ := os.ReadFile(filepath.Join(infraDir, "main.tf"))
	if !strings.Contains(string(mainContent), "bigquery.googleapis.com") {
		t.Errorf("main.tf must enable bigquery.googleapis.com when BigQuery is active")
	}

	modulesContent, _ := os.ReadFile(filepath.Join(infraDir, "modules.tf"))
	if !strings.Contains(string(modulesContent), "resource \"google_bigquery_dataset\" \"analytics_dataset\"") {
		t.Errorf("modules.tf must define analytics_dataset")
	}
	if !strings.Contains(string(modulesContent), "default_table_expiration_ms = 5184000000") {
		t.Errorf("modules.tf must cap table expiration to 60 days to prevent exceeding 10GB")
	}
	if !strings.Contains(string(modulesContent), "time_partitioning") {
		t.Errorf("modules.tf must configure time_partitioning for BigQuery events table")
	}

	outputsContent, _ := os.ReadFile(filepath.Join(infraDir, "outputs.tf"))
	if !strings.Contains(string(outputsContent), "output \"bigquery_dataset_id\"") {
		t.Errorf("outputs.tf must export bigquery_dataset_id")
	}
	if !strings.Contains(string(outputsContent), "output \"bigquery_events_table_id\"") {
		t.Errorf("outputs.tf must export bigquery_events_table_id")
	}

	// Validate canonical Terraform formatting
	diffCmd := exec.Command("terraform", "fmt", "-diff", infraDir)
	if out, err := diffCmd.CombinedOutput(); err != nil || len(out) > 0 {
		t.Errorf("terraform fmt failed on BigQuery generated files:\n%s", string(out))
	}
}

func TestGenerateProject_WithCloudFunctions(t *testing.T) {
	tempDir := t.TempDir()
	infraDir := filepath.Join(tempDir, "infra")

	config := generator.ProjectConfig{
		ProjectName:           "fn-test-app",
		GCPProjectID:          "test-fn-123",
		Region:                "us-central1",
		IncludeCloudFunctions: true,
	}

	err := generator.GenerateProject(config, infraDir)
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	mainContent, _ := os.ReadFile(filepath.Join(infraDir, "main.tf"))
	if !strings.Contains(string(mainContent), "cloudfunctions.googleapis.com") {
		t.Errorf("main.tf must enable cloudfunctions.googleapis.com")
	}
	if !strings.Contains(string(mainContent), "cloudbuild.googleapis.com") {
		t.Errorf("main.tf must enable cloudbuild.googleapis.com")
	}

	providersContent, _ := os.ReadFile(filepath.Join(infraDir, "providers.tf"))
	if !strings.Contains(string(providersContent), "archive = {") {
		t.Errorf("providers.tf must declare archive provider when Cloud Functions is active")
	}

	computeContent, _ := os.ReadFile(filepath.Join(infraDir, "compute.tf"))
	if !strings.Contains(string(computeContent), "resource \"google_cloudfunctions2_function\" \"app_function\"") {
		t.Errorf("compute.tf must define app_function")
	}
	if !strings.Contains(string(computeContent), "min_instance_count = 0") {
		t.Errorf("compute.tf must enforce min_instance_count = 0 (scale to zero)")
	}
	if !strings.Contains(string(computeContent), "available_memory   = \"256M\"") {
		t.Errorf("compute.tf must configure 256M available memory for Free Tier")
	}
	if !strings.Contains(string(computeContent), "age = 7") {
		t.Errorf("compute.tf must configure 7-day lifecycle purge rule on deployment bucket")
	}

	outputsContent, _ := os.ReadFile(filepath.Join(infraDir, "outputs.tf"))
	if !strings.Contains(string(outputsContent), "output \"cloud_function_name\"") {
		t.Errorf("outputs.tf must export cloud_function_name")
	}
	if !strings.Contains(string(outputsContent), "output \"cloud_function_uri\"") {
		t.Errorf("outputs.tf must export cloud_function_uri")
	}

	// Validate canonical Terraform formatting
	diffCmd := exec.Command("terraform", "fmt", "-diff", infraDir)
	if out, err := diffCmd.CombinedOutput(); err != nil || len(out) > 0 {
		t.Errorf("terraform fmt failed on Cloud Functions generated files:\n%s", string(out))
	}
}

func TestGenerateProject_WithVMIdleShutdown(t *testing.T) {
	tempDir := t.TempDir()
	infraDir := filepath.Join(tempDir, "infra")

	config := generator.ProjectConfig{
		ProjectName:          "vm-idle-app",
		GCPProjectID:         "test-idle-123",
		Region:               "us-central1",
		IncludeComputeEngine: true,
		VMIdleShutdown:       true,
	}

	err := generator.GenerateProject(config, infraDir)
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	computeContent, _ := os.ReadFile(filepath.Join(infraDir, "compute.tf"))
	if !strings.Contains(string(computeContent), "metadata_startup_script") {
		t.Errorf("compute.tf must inject metadata_startup_script when VMIdleShutdown is true")
	}
	if !strings.Contains(string(computeContent), "idle-shutdown.sh") {
		t.Errorf("compute.tf must include idle-shutdown.sh watchdog script")
	}

	// Validate canonical Terraform formatting
	diffCmd := exec.Command("terraform", "fmt", "-diff", infraDir)
	if out, err := diffCmd.CombinedOutput(); err != nil || len(out) > 0 {
		t.Errorf("terraform fmt failed on VM with Idle Shutdown generated files:\n%s", string(out))
	}
}
