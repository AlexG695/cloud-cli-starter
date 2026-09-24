package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/tu-usuario/cloud/internal/templates"
)

type templateMapping struct {
	tmplName string
	outName  string
}

// templateFiles defines the mapping between embedded templates and generated output files.
var templateFiles = []templateMapping{
	{tmplName: "providers.tf.tmpl", outName: "providers.tf"},
	{tmplName: "main.tf.tmpl", outName: "main.tf"},
	{tmplName: "compute.tf.tmpl", outName: "compute.tf"},
	{tmplName: "modules.tf.tmpl", outName: "modules.tf"},
	{tmplName: "variables.tf.tmpl", outName: "variables.tf"},
	{tmplName: "outputs.tf.tmpl", outName: "outputs.tf"},
	{tmplName: "Makefile.tmpl", outName: "Makefile"},
	{tmplName: "docker-compose.yml.tmpl", outName: "docker-compose.yml"},
	{tmplName: "firebase.json.tmpl", outName: "firebase.json"},
	{tmplName: "firestore.rules.tmpl", outName: "firestore.rules"},
	{tmplName: "storage.rules.tmpl", outName: "storage.rules"},
}

func isEmulatorFile(tmplName string) bool {
	switch tmplName {
	case "docker-compose.yml.tmpl", "firebase.json.tmpl", "firestore.rules.tmpl", "storage.rules.tmpl":
		return true
	default:
		return false
	}
}

// GenerateProject reads embedded files from internal/templates/embed.FS using text/template,
// renders them with ProjectConfig data, and writes the final files
// into the specified output directory (outputDir).
func GenerateProject(config ProjectConfig, outputDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory '%s': %w", outputDir, err)
	}

	for _, m := range templateFiles {
		// Skip local emulator files if not requested
		if isEmulatorFile(m.tmplName) && !config.IncludeLocalEmulators {
			continue
		}

		content, err := templates.FS.ReadFile(m.tmplName)
		if err != nil {
			return fmt.Errorf("error reading embedded template '%s': %w", m.tmplName, err)
		}

		tmpl, err := template.New(m.tmplName).Parse(string(content))
		if err != nil {
			return fmt.Errorf("error parsing template '%s': %w", m.tmplName, err)
		}

		targetPath := filepath.Join(outputDir, m.outName)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("error creating subdirectory for '%s': %w", targetPath, err)
		}

		outFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("error creating target file '%s': %w", targetPath, err)
		}

		if err := tmpl.Execute(outFile, config); err != nil {
			outFile.Close()
			return fmt.Errorf("error rendering template '%s' into '%s': %w", m.tmplName, targetPath, err)
		}

		if err := outFile.Close(); err != nil {
			return fmt.Errorf("error closing file '%s': %w", targetPath, err)
		}
	}

	return nil
}
