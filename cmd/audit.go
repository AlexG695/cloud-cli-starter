package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tu-usuario/cloud/internal/ai"
	"github.com/tu-usuario/cloud/internal/audit"
)

// NewAuditCmd creates and returns a new command instance for 'cloud audit'.
func NewAuditCmd() *cobra.Command {
	var (
		auditDir     string
		outputFormat string
		strict       bool
		withAI       bool
	)

	cmd := &cobra.Command{
		Use:     "audit [dir]",
		Aliases: []string{"check-costs", "lint"},
		Short:   "Audits Terraform configuration for Always Free compliance and Zero-Trust",
		Long: `Audits an existing Terraform directory against Google Cloud Always Free Tier limits
(e2-micro VM, 30GB standard disk, Cloud Run 256Mi memory, scale-to-zero)
and evaluates Zero-Trust security postures. Optionally runs an AI-assisted audit.`,
		RunE: func(c *cobra.Command, args []string) error {
			targetDir := auditDir
			if len(args) > 0 && args[0] != "" {
				targetDir = args[0]
			}

			if outputFormat != "text" && outputFormat != "json" {
				return fmt.Errorf("invalid format '%s': supported formats are 'text' and 'json'", outputFormat)
			}

			report, err := audit.AuditDirectory(targetDir)
			if err != nil {
				return err
			}

			// Run optional AI security audit
			if withAI {
				var combinedTF strings.Builder
				entries, _ := os.ReadDir(targetDir)
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
						data, _ := os.ReadFile(filepath.Join(targetDir, entry.Name()))
						combinedTF.Write(data)
						combinedTF.WriteString("\n")
					}
				}
				if combinedTF.Len() > 0 {
					aiRec, err := ai.Audit(combinedTF.String())
					if err == nil {
						report.AIRecommendations = aiRec
					}
				}
			}

			if outputFormat == "json" {
				jsonData, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return fmt.Errorf("error serializing audit report to JSON: %w", err)
				}
				fmt.Println(string(jsonData))
			} else {
				printTextReport(report)
			}

			if strict && !report.Passed {
				return fmt.Errorf("audit failed with %d violation(s) under --strict mode", report.Summary.Fail)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&auditDir, "dir", "d", "infra", "Target directory containing Terraform (.tf) files")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text|json")
	cmd.Flags().BoolVar(&strict, "strict", false, "Exit with code 1 if any FAIL violations are detected")
	cmd.Flags().BoolVar(&withAI, "with-ai", false, "Include Zero-Trust AI security recommendations")

	return cmd
}

func printTextReport(report *audit.AuditReport) {
	fmt.Println("\n==================================================")
	fmt.Println("🛡️  GCP Always Free & Zero-Trust Audit Report")
	fmt.Println("==================================================")
	fmt.Printf("📂 Target Directory: %s\n\n", report.Directory)

	if len(report.Findings) == 0 {
		fmt.Println("✅ No compliance violations found. All resources adhere strictly to Always Free Tier!")
	} else {
		for i, f := range report.Findings {
			badge := "ℹ️  INFO"
			switch f.Severity {
			case audit.SeverityFail:
				badge = "❌ FAIL"
			case audit.SeverityWarn:
				badge = "⚠️  WARN"
			}

			fileInfo := ""
			if f.File != "" {
				fileInfo = fmt.Sprintf(" (%s)", f.File)
			}

			fmt.Printf("%d. [%s] %s%s\n", i+1, badge, f.Resource, fileInfo)
			fmt.Printf("   Issue:          %s\n", f.Message)
			fmt.Printf("   Recommendation: %s\n\n", f.Recommendation)
		}
	}

	fmt.Println("--------------------------------------------------")
	fmt.Printf("📊 Summary: %d Total | %d Fail(s) | %d Warning(s)\n",
		report.Summary.Total, report.Summary.Fail, report.Summary.Warn)
	fmt.Println("--------------------------------------------------")

	if report.AIRecommendations != "" {
		fmt.Println("\n🤖 AI Zero-Trust Security Recommendations:")
		fmt.Println(report.AIRecommendations)
	}

	if report.Passed {
		fmt.Println("\n✅ Status: PASSED - Infrastructure complies with Always Free limits ($0/month).")
	} else {
		fmt.Println("\n❌ Status: FAILED - Potential billing or quota violations detected!")
	}
}
