package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	reComputeInstance = regexp.MustCompile(`(?s)resource\s+"google_compute_instance"\s+"([^"]+)"\s*\{(.*?)\n\}`)
	reCloudRun        = regexp.MustCompile(`(?s)resource\s+"google_cloud_run(?:_v2)?_service"\s+"([^"]+)"\s*\{(.*?)\n\}`)
	reStorageBucket   = regexp.MustCompile(`(?s)resource\s+"google_storage_bucket"\s+"([^"]+)"\s*\{(.*?)\n\}`)
	reBigQueryDataset = regexp.MustCompile(`(?s)resource\s+"google_bigquery_dataset"\s+"([^"]+)"\s*\{(.*?)\n\}`)
	reServiceAccount  = regexp.MustCompile(`resource\s+"google_service_account"`)

	reMachineType = regexp.MustCompile(`machine_type\s*=\s*"([^"]+)"`)
	reDiskType    = regexp.MustCompile(`type\s*=\s*"([^"]+)"`)
	reDiskSize    = regexp.MustCompile(`size\s*=\s*([0-9]+)`)
	reMemoryLimit = regexp.MustCompile(`(?:memory|available_memory)\s*=\s*"([^"]+)"`)
	reMinInstance = regexp.MustCompile(`min_instance_count\s*=\s*([0-9]+)`)
	reRegionVar   = regexp.MustCompile(`variable\s+"gcp_region"\s*\{[^}]*default\s*=\s*"([^"]+)"`)
	rePublicPrev  = regexp.MustCompile(`public_access_prevention\s*=\s*"([^"]+)"`)
	reUniformAcc  = regexp.MustCompile(`uniform_bucket_level_access\s*=\s*(true|false)`)
	reDatasetExp  = regexp.MustCompile(`default_table_expiration_ms\s*=\s*([0-9]+)`)
)

// Allowed Always Free US regions
var allowedFreeRegions = map[string]bool{
	"us-central1": true,
	"us-east1":    true,
	"us-west1":    true,
}

// AuditDirectory analyzes all .tf files in the given directory and returns an AuditReport.
func AuditDirectory(dir string) (*AuditReport, error) {
	report := &AuditReport{
		Directory: dir,
		Passed:    true,
		Findings:  []Finding{},
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory '%s': %w", dir, err)
	}

	var tfFiles []string
	allContent := ""
	fileContents := make(map[string]string)

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
			fullPath := filepath.Join(dir, entry.Name())
			data, err := os.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("error reading file '%s': %w", fullPath, err)
			}
			content := string(data)
			tfFiles = append(tfFiles, entry.Name())
			fileContents[entry.Name()] = content
			allContent += "\n" + content
		}
	}

	if len(tfFiles) == 0 {
		return nil, fmt.Errorf("no Terraform (.tf) files found in directory '%s'", dir)
	}

	// 1. Audit Region
	if m := reRegionVar.FindStringSubmatch(allContent); len(m) > 1 {
		region := m[1]
		if !allowedFreeRegions[region] && !strings.HasPrefix(region, "{{") {
			report.addFinding(Finding{
				Resource:       "variable.gcp_region",
				Severity:       SeverityFail,
				Message:        fmt.Sprintf("Region '%s' is not eligible for Always Free compute/storage", region),
				Recommendation: "Change region to us-central1, us-east1, or us-west1 to guarantee $0 cost.",
				File:           "variables.tf",
			})
		}
	}

	// 2. Audit Compute Engine Instances
	hasCompute := false
	for file, content := range fileContents {
		instances := reComputeInstance.FindAllStringSubmatch(content, -1)
		for _, inst := range instances {
			hasCompute = true
			name := inst[1]
			body := inst[2]

			// Machine type check
			if mm := reMachineType.FindStringSubmatch(body); len(mm) > 1 {
				mType := mm[1]
				if mType != "e2-micro" && !strings.HasPrefix(mType, "{{") {
					report.addFinding(Finding{
						Resource:       fmt.Sprintf("google_compute_instance.%s", name),
						Severity:       SeverityFail,
						Message:        fmt.Sprintf("Machine type '%s' is not eligible for Always Free (e2-micro required)", mType),
						Recommendation: "Change machine_type to 'e2-micro'.",
						File:           file,
					})
				}
			}

			// Disk checks inside boot_disk block
			bootDiskMatches := regexp.MustCompile(`(?s)boot_disk\s*\{(.*?)\}`).FindStringSubmatch(body)
			if len(bootDiskMatches) > 1 {
				bootDiskBody := bootDiskMatches[1]
				diskTypeMatches := regexp.MustCompile(`\btype\s*=\s*"([^"]+)"`).FindStringSubmatch(bootDiskBody)
				if len(diskTypeMatches) > 1 {
					dType := diskTypeMatches[1]
					if dType != "pd-standard" && !strings.HasPrefix(dType, "{{") {
						report.addFinding(Finding{
							Resource:       fmt.Sprintf("google_compute_instance.%s", name),
							Severity:       SeverityFail,
							Message:        fmt.Sprintf("Disk type '%s' incurs billing (only 'pd-standard' is free)", dType),
							Recommendation: "Change boot_disk type to 'pd-standard'.",
							File:           file,
						})
					}
				}

				diskSizeMatches := regexp.MustCompile(`\bsize\s*=\s*([0-9]+)`).FindStringSubmatch(bootDiskBody)
				if len(diskSizeMatches) > 1 {
					size, _ := strconv.Atoi(diskSizeMatches[1])
					if size > 30 {
						report.addFinding(Finding{
							Resource:       fmt.Sprintf("google_compute_instance.%s", name),
							Severity:       SeverityFail,
							Message:        fmt.Sprintf("Disk size %d GB exceeds the 30 GB Always Free limit", size),
							Recommendation: "Reduce boot disk size to 30 GB or less.",
							File:           file,
						})
					}
				}
			}
		}
	}

	// 3. Audit Cloud Run Services
	for file, content := range fileContents {
		services := reCloudRun.FindAllStringSubmatch(content, -1)
		for _, svc := range services {
			hasCompute = true
			name := svc[1]
			body := svc[2]

			// Memory check
			if mem := reMemoryLimit.FindStringSubmatch(body); len(mem) > 1 {
				val := mem[1]
				if val != "256Mi" && val != "256M" && val != "128Mi" && !strings.HasPrefix(val, "{{") {
					report.addFinding(Finding{
						Resource:       fmt.Sprintf("google_cloud_run_service.%s", name),
						Severity:       SeverityFail,
						Message:        fmt.Sprintf("Memory limit '%s' exceeds Free Tier safe default (256Mi recommended)", val),
						Recommendation: "Set container memory limit to '256Mi'.",
						File:           file,
					})
				}
			}

			// Min instance check (scale to zero)
			if mi := reMinInstance.FindStringSubmatch(body); len(mi) > 1 {
				count, _ := strconv.Atoi(mi[1])
				if count > 0 {
					report.addFinding(Finding{
						Resource:       fmt.Sprintf("google_cloud_run_service.%s", name),
						Severity:       SeverityFail,
						Message:        fmt.Sprintf("min_instance_count = %d prevents service from scaling to zero when idle", count),
						Recommendation: "Set min_instance_count = 0 to avoid idle compute charges.",
						File:           file,
					})
				}
			}
		}
	}

	// 4. Audit Cloud Storage
	for file, content := range fileContents {
		buckets := reStorageBucket.FindAllStringSubmatch(content, -1)
		for _, b := range buckets {
			name := b[1]
			body := b[2]

			if pap := rePublicPrev.FindStringSubmatch(body); len(pap) > 1 {
				if pap[1] != "enforced" {
					report.addFinding(Finding{
						Resource:       fmt.Sprintf("google_storage_bucket.%s", name),
						Severity:       SeverityWarn,
						Message:        "public_access_prevention is not set to 'enforced'",
						Recommendation: "Add public_access_prevention = \"enforced\" to secure bucket data.",
						File:           file,
					})
				}
			} else {
				report.addFinding(Finding{
					Resource:       fmt.Sprintf("google_storage_bucket.%s", name),
					Severity:       SeverityWarn,
					Message:        "Bucket lacks public_access_prevention enforcement",
					Recommendation: "Add public_access_prevention = \"enforced\" for Zero-Trust compliance.",
					File:           file,
				})
			}

			if ubla := reUniformAcc.FindStringSubmatch(body); len(ubla) > 1 {
				if ubla[1] != "true" {
					report.addFinding(Finding{
						Resource:       fmt.Sprintf("google_storage_bucket.%s", name),
						Severity:       SeverityWarn,
						Message:        "uniform_bucket_level_access is not enabled",
						Recommendation: "Set uniform_bucket_level_access = true for IAM least-privilege.",
						File:           file,
					})
				}
			}
		}
	}

	// 5. Audit BigQuery Datasets
	for file, content := range fileContents {
		datasets := reBigQueryDataset.FindAllStringSubmatch(content, -1)
		for _, ds := range datasets {
			name := ds[1]
			body := ds[2]

			if !reDatasetExp.MatchString(body) {
				report.addFinding(Finding{
					Resource:       fmt.Sprintf("google_bigquery_dataset.%s", name),
					Severity:       SeverityWarn,
					Message:        "BigQuery dataset lacks default_table_expiration_ms",
					Recommendation: "Configure table expiration (e.g. 60 days) to keep stored data strictly under 10 GB.",
					File:           file,
				})
			}
		}
	}

	// 6. Zero-Trust Identity Audit
	if hasCompute && !reServiceAccount.MatchString(allContent) {
		report.addFinding(Finding{
			Resource:       "iam.service_account",
			Severity:       SeverityWarn,
			Message:        "Compute resources detected without dedicated Service Accounts",
			Recommendation: "Create least-privilege Service Accounts to avoid relying on default Compute Engine Editor SA.",
		})
	}

	return report, nil
}

func (r *AuditReport) addFinding(f Finding) {
	r.Findings = append(r.Findings, f)
	r.Summary.Total++
	switch f.Severity {
	case SeverityFail:
		r.Summary.Fail++
		r.Passed = false
	case SeverityWarn:
		r.Summary.Warn++
	case SeverityInfo:
		r.Summary.Info++
	}
}
