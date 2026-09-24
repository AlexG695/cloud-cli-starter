package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tu-usuario/cloud/internal/generator"
	"github.com/tu-usuario/cloud/internal/interactive"
)

// InitCmd represents the main 'cloud' command of the CLI
var InitCmd = &cobra.Command{
	Use:     "cloud",
	Aliases: []string{"init"},
	Short:   "Initializes the configuration for a new cloud project",
	Long: `Cloud CLI runs an interactive wizard using survey to gather
project requirements (Cloud Run microservices, optional modules)
and generates the ProjectConfig structure, printing it in JSON format for verification
and rendering the infrastructure templates in the specified directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := interactive.RunPrompts()
		if err != nil {
			return fmt.Errorf("interactive wizard error: %w", err)
		}

		jsonData, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("error serializing configuration to JSON: %w", err)
		}

		fmt.Println("\n==================================================")
		fmt.Println("🎉 Configuration successfully generated (JSON):")
		fmt.Println("==================================================")
		fmt.Println(string(jsonData))

		outputDir := "infra"
		if err := generator.GenerateProject(*cfg, outputDir); err != nil {
			return fmt.Errorf("error generating project files: %w", err)
		}

		fmt.Println()
		fmt.Printf("✅ Architecture successfully generated in folder ./%s!\n", outputDir)
		fmt.Printf("👉 To get started: cd %s && make help\n", outputDir)
		if cfg.IncludeLocalEmulators {
			fmt.Println("🌐 Emulators UI Dashboard available at: http://localhost:4000 (run 'make emulators-up')")
		}

		return nil
	},
}

// RootCmd is an alias of InitCmd as the root command
var RootCmd = InitCmd

// NewInitCmd returns the cloud command instance for registration.
func NewInitCmd() *cobra.Command {
	return InitCmd
}

// Execute executes the CLI root command
func Execute() error {
	return InitCmd.Execute()
}
