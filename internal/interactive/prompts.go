package interactive

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/tu-usuario/cloud/internal/generator"
)

// Optional architectural modules checklist options
const (
	optStorage        = "IncludeStorage (Private Cloud Storage bucket)"
	optFirestore      = "IncludeFirestore (Cloud Firestore NoSQL database)"
	optAuth           = "IncludeAuth (Identity Platform / Firebase Auth)"
	optCloudflare     = "IncludeCloudflare (Cloudflare DNS, SSL & Proxy)"
	optLocalEmulators = "IncludeLocalEmulators (Local emulators with docker-compose.yml)"
)

// RunPrompts guides the user step-by-step through interactive questions
// to build the complete ProjectConfig.
func RunPrompts() (*generator.ProjectConfig, error) {
	fmt.Println("🚀 Starting interactive wizard for Cloud CLI")
	fmt.Println("--------------------------------------------------")

	// 1. Basic project settings
	var projectName string
	if err := survey.AskOne(&survey.Input{
		Message: "Project name:",
		Default: "cloud-app",
		Help:    "Human-readable project / repository name",
	}, &projectName, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	var gcpProjectID string
	if err := survey.AskOne(&survey.Input{
		Message: "GCP Project ID:",
		Help:    "Google Cloud Console Project ID (e.g., my-project-prod-123)",
	}, &gcpProjectID, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	// Initialize with default Free Tier settings
	cfg := generator.NewDefaultConfig(projectName, gcpProjectID)

	// 2. Loop to add N Cloud Run microservices
	fmt.Println("\n📦 Cloud Run microservices configuration")
	var services []generator.CloudRunService

	for {
		addService := false
		promptMessage := "Do you want to add a Cloud Run microservice?"
		if len(services) > 0 {
			promptMessage = "Do you want to add another Cloud Run microservice?"
		}

		if err := survey.AskOne(&survey.Confirm{
			Message: promptMessage,
			Default: len(services) == 0,
		}, &addService); err != nil {
			return nil, err
		}

		if !addService {
			break
		}

		var svcName string
		if err := survey.AskOne(&survey.Input{
			Message: fmt.Sprintf("Microservice #%d name:", len(services)+1),
			Help:    "Cloud Run service identifier (e.g., api-gateway, auth-service)",
		}, &svcName, survey.WithValidator(survey.Required)); err != nil {
			return nil, err
		}

		// Visibility: public or private
		var accessChoice string
		if err := survey.AskOne(&survey.Select{
			Message: fmt.Sprintf("Visibility for '%s':", svcName),
			Options: []string{
				"Public (internet access without prior authentication)",
				"Private (requires IAM / Service Account authentication)",
			},
			Default: "Public (internet access without prior authentication)",
		}, &accessChoice); err != nil {
			return nil, err
		}
		isPublic := strings.HasPrefix(accessChoice, "Public")

		// Pub/Sub integration
		enablePubSub := false
		if err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Will '%s' integrate with Cloud Pub/Sub (Topic + Push Subscription)?", svcName),
			Default: false,
		}, &enablePubSub); err != nil {
			return nil, err
		}

		services = append(services, generator.CloudRunService{
			Name:         svcName,
			Memory:       "256Mi",
			CPU:          "1000m",
			IsPublic:     isPublic,
			EnablePubSub: enablePubSub,
		})
	}
	cfg.Services = services

	// 3. Optional architecture modules checklist
	fmt.Println("\n⚙️  Optional architecture modules")
	var selectedModules []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select optional modules to include:",
		Options: []string{
			optStorage,
			optFirestore,
			optAuth,
			optCloudflare,
			optLocalEmulators,
		},
		Default: []string{
			optStorage,
			optLocalEmulators,
		},
		Help: "Use arrow keys to navigate, space to select/deselect, and Enter to confirm",
	}, &selectedModules); err != nil {
		return nil, err
	}

	selectedMap := make(map[string]bool)
	for _, m := range selectedModules {
		selectedMap[m] = true
	}

	cfg.IncludeStorage = selectedMap[optStorage]
	cfg.IncludeFirestore = selectedMap[optFirestore]
	cfg.IncludeAuth = selectedMap[optAuth]
	cfg.EnableIdentityPlatform = cfg.IncludeAuth
	cfg.IncludeCloudflare = selectedMap[optCloudflare]
	cfg.IncludeLocalEmulators = selectedMap[optLocalEmulators]

	// If Cloudflare was selected, optionally ask for domain
	if cfg.IncludeCloudflare {
		var domain string
		_ = survey.AskOne(&survey.Input{
			Message: "Domain for Cloudflare (optional, e.g., api.mydomain.com):",
		}, &domain)
		cfg.DomainName = strings.TrimSpace(domain)
	}

	// AI integration paused for now
	cfg.EnableGemmaAssistant = false

	return &cfg, nil
}
