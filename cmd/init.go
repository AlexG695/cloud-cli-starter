package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tu-usuario/cloud/internal/generator"
	"github.com/tu-usuario/cloud/internal/interactive"
)

// NewInitCmd creates and returns an isolated instance of the cloud command.
func NewInitCmd() *cobra.Command {
	var (
		outputFormat string
		dryRun       bool
		configFile   string
		outputDir    string
	)

	cmd := &cobra.Command{
		Use:     "cloud",
		Aliases: []string{"init"},
		Short:   "Initializes the configuration for a new cloud project",
		Long: `Cloud CLI runs an interactive wizard using survey to gather
project requirements (Cloud Run microservices, optional modules)
and generates the ProjectConfig structure, printing it in JSON format for verification
and rendering the infrastructure templates in the specified directory.`,
		RunE: func(c *cobra.Command, args []string) error {
			if outputFormat != "text" && outputFormat != "json" {
				return fmt.Errorf("invalid format '%s': supported formats are 'text' and 'json'", outputFormat)
			}

			var cfg *generator.ProjectConfig
			var err error

			if configFile != "" {
				data, err := os.ReadFile(configFile)
				if err != nil {
					return fmt.Errorf("error reading config file '%s': %w", configFile, err)
				}
				var parsedCfg generator.ProjectConfig
				if err := json.Unmarshal(data, &parsedCfg); err != nil {
					return fmt.Errorf("error parsing JSON config file: %w", err)
				}
				cfg = &parsedCfg
			} else {
				cfg, err = interactive.RunPrompts()
				if err != nil {
					return fmt.Errorf("interactive wizard error: %w", err)
				}
			}

			jsonData, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("error serializing configuration to JSON: %w", err)
			}

			if outputFormat == "json" {
				// Pure JSON to stdout for CI/CD pipelines
				fmt.Println(string(jsonData))
			} else {
				// Human-friendly text mode
				fmt.Println("\n==================================================")
				fmt.Println("🎉 Configuration successfully generated (JSON):")
				fmt.Println("==================================================")
				fmt.Println(string(jsonData))
			}

			if dryRun {
				if outputFormat == "text" {
					fmt.Println("\n🔍 Dry-run mode enabled: No files were written to disk.")
				}
				return nil
			}

			if err := generator.GenerateProject(*cfg, outputDir); err != nil {
				return fmt.Errorf("error generating project files: %w", err)
			}

			if outputFormat == "text" {
				fmt.Println()
				fmt.Printf("✅ Architecture successfully generated in folder ./%s!\n", outputDir)
				fmt.Printf("👉 To get started: cd %s && make help\n", outputDir)
				if cfg.IncludeLocalEmulators {
					fmt.Println("🌐 Emulators UI Dashboard available at: http://localhost:4000 (run 'make emulators-up')")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text|json")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate configuration and validation without writing files to disk")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to JSON configuration file to bypass interactive wizard")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "infra", "Target output directory for generated infrastructure files")

	cmd.AddCommand(NewAuditCmd())

	return cmd
}

// InitCmd represents the main 'cloud' command of the CLI
var InitCmd = NewInitCmd()

// RootCmd is an alias of InitCmd as the root command
var RootCmd = InitCmd

// Execute executes the CLI root command
func Execute() error {
	return InitCmd.Execute()
}
