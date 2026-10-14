package generator

// CloudRunService defines the configuration of an individual Cloud Run microservice.
type CloudRunService struct {
	Name         string `json:"name"`
	Memory       string `json:"memory"`        // e.g.: "256Mi" (Guarantees Free Tier)
	CPU          string `json:"cpu"`           // e.g.: "1000m" (1 vCPU)
	IsPublic     bool   `json:"is_public"`     // true = internet-facing | false = dedicated Service Account
	EnablePubSub bool   `json:"enable_pubsub"` // true = creates Topic and Push Subscription
}

// ProjectConfig is the main structure storing all user choices.
type ProjectConfig struct {
	// Project identification
	ProjectName  string `json:"project_name"`   // e.g.: "geo-api"
	GCPProjectID string `json:"gcp_project_id"` // e.g.: "geo-engine-prod-123"
	Region       string `json:"region"`         // e.g.: "us-central1" (Always Free Tier)

	// Dynamic microservices (Cloud Run loop)
	Services []CloudRunService `json:"services"`

	IncludeStorage         bool `json:"include_storage"`          // Private Cloud Storage bucket
	IncludeFirestore       bool `json:"include_firestore"`        // Cloud Firestore NoSQL database
	IncludeAuth            bool `json:"include_auth"`             // Identity Platform / Firebase Auth
	EnableIdentityPlatform bool `json:"enable_identity_platform"` // Enables emulator and Identity Platform resources
	IncludeSecrets         bool `json:"include_secrets"`          // Secret Manager
	IncludeComputeEngine   bool `json:"include_compute_engine"`   // Free Tier Compute Engine VM (e2-micro, 30GB pd-standard)
	VMPublicIP             bool `json:"vm_public_ip"`             // true = ephemeral public IP | false = internal only (Zero-Trust/IAP)
	VMIdleShutdown         bool `json:"vm_idle_shutdown"`         // Shuts down VM after 30 mins of idle CPU (<1%)
	IncludeBigQuery        bool `json:"include_bigquery"`         // BigQuery dataset capped with 60-day expiration (10GB limit)
	IncludeCloudFunctions  bool `json:"include_cloud_functions"`  // Cloud Functions v2 (2M invocations/mo, scale to zero)

	// Local Development & AI environment
	IncludeLocalEmulators bool `json:"include_local_emulators"` // Generates docker-compose.yml
	EnableGemmaAssistant  bool `json:"enable_gemma_assistant"`  // Enables Zero-Trust audit with Gemma 2B
}

// NewDefaultConfig creates a base configuration with safe values within the Free Tier.
func NewDefaultConfig(projectName, projectID string) ProjectConfig {
	return ProjectConfig{
		ProjectName:           projectName,
		GCPProjectID:          projectID,
		Region:                "us-central1", // Default region for Always Free
		IncludeStorage:        true,          // Recommended default
		IncludeSecrets:        true,          // Recommended default
		IncludeComputeEngine:  false,
		VMPublicIP:            false,
		VMIdleShutdown:        false,
		IncludeBigQuery:       false,
		IncludeCloudFunctions: false,
		IncludeLocalEmulators: true, // Generates docker-compose by default
		EnableGemmaAssistant:  false,
		Services:              []CloudRunService{},
	}
}

// HasServices returns true if at least one Cloud Run service is configured.
func (c ProjectConfig) HasServices() bool {
	return len(c.Services) > 0
}

// HasPubSub returns true if at least one service has Cloud Pub/Sub integration enabled.
func (c ProjectConfig) HasPubSub() bool {
	for _, s := range c.Services {
		if s.EnablePubSub {
			return true
		}
	}
	return false
}

