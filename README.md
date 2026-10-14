# GCP Free Tier Monitor (CLI)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](go.mod)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/AlexG695/cloud-cli-starter/pulls)
[![CI Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
[![GCP Cost](https://img.shields.io/badge/GCP_Cost-%240%2Fmonth-success.svg)](#)

English | [Español](README.es.md)

**Prevent surprise bills on Google Cloud Platform by scaffolding and provisioning serverless architectures strictly engineered for the Always Free Tier ($0/month), backed by local emulators and AI-driven Zero-Trust security audits.**

![Demo](docs/assets/demo.gif)

---

## 📌 The Problem vs. The Solution

* **Phantom cloud costs & idle resources:** Developers, students, and startups frequently face unexpected bills at month's end due to abandoned resources, unscaled idle compute, misconfigured network egress, or provisioning in regions ineligible for GCP's Always Free tier.
* **Complex, scattered Always Free rules:** Navigating Google Cloud's free tier limits (e.g., `us-central1` exclusivity, 256MiB RAM limits per Cloud Run service, mandatory scale-to-zero, monthly API quota caps, and regional storage policies) is tedious and error-prone when clicking through the web console.
* **Secure-by-design architecture before charges occur:** This CLI eliminates risk at the root. Through an interactive terminal wizard, it generates validated Terraform infrastructure-as-code locked within $0/month limits, spins up a complete offline Docker Compose emulator suite for local testing, and audits configurations using AI models to enforce Zero-Trust and least-privilege IAM.

---

## 🚀 Key Features

* **Interactive Terminal Wizard:** Guided step-by-step CLI survey (built with `survey`) that captures configuration, microservices, and architectural modules in seconds.
* **Guaranteed $0/Month (Always Free Defaults):** Defaults strictly locked to `us-central1`, 256Mi RAM per container, 1000m (1 vCPU), and scale-to-zero policy (`min_instance_count = 0`).
* **Modular Infrastructure-as-Code (Terraform):** Produces clean, canonical Terraform files (`main.tf`, `compute.tf`, `modules.tf`, `variables.tf`, `outputs.tf`, `providers.tf`) validated with `terraform fmt`.
* **Subcommand `cloud audit` for Cost Prevention:** Native static analysis engine that audits Terraform configuration files (.tf). Detects billing risks (non-`e2-micro` machines, disks larger than 30 GB or non-standard types, containers without scale-to-zero, missing BigQuery expiration) with `--strict` mode to fail CI/CD pipelines.
* **CI/CD Automation & Headless Flags:** Automated pipeline execution via `--config config.json`, preview simulation with `--dry-run`, structured machine output via `--format json`, and custom destination directory with `--output-dir`.
* **Extended Always Free Modules:** Out-of-the-box support for Compute Engine (`e2-micro` VM with private or public IP options and automated idle shutdown via Cloud Scheduler), BigQuery (partitioned datasets capped under 10 GB with 60-day auto-expiration), and Cloud Functions v2 with a 7-day storage zip lifecycle.
* **Zero-Trust Identity & Least Privilege (IAM):** Decouples from default Compute Engine service accounts; provisions dedicated, least-privilege Service Accounts per Cloud Run service (`roles/datastore.user`, `roles/storage.objectUser`, `roles/secretmanager.secretAccessor`).
* **Complete Local Emulator Suite (Docker Compose):** Instant offline environment via Firebase Emulator Suite (`make emulators-up`) emulating Firestore, Storage, Pub/Sub, and Auth with an interactive web UI at `http://localhost:4000`.
* **Multi-Engine AI Security Auditor:** Built-in Zero-Trust audit engine supporting local, offline Ollama (`gemma:2b`), with automatic fallbacks to Google Gemini (`gemini-1.5-flash`), OpenAI (`gpt-4o-mini`), and graceful degradation if no AI backend is detected.
* **Lifecycle Makefile Automation:** One-liner commands to format (`make fmt`), validate (`make check`), plan (`make plan`), deploy (`make apply`), tear down (`make destroy`), and manage local emulators (`make emulators-up` / `make emulators-down`).

---

## 📊 Free Tier Coverage Matrix

| GCP Service | Monitored / Provisioned Metric | Free Tier Limit (Always Free) | MVP Status |
| :--- | :--- | :--- | :---: |
| **Cloud Run** | Memory (256Mi), CPU (1 vCPU), Scale-to-zero (`min_instance_count = 0`) | 2M requests/mo, 360,000 GB-s, 180,000 vCPU-s (us-central1) | ✅ Supported |
| **Cloud Storage** | Regional bucket in `us-central1`, enforced uniform & private access | 5 GB-month standard storage, 5,000 Class A ops, 50,000 Class B ops | ✅ Supported |
| **Cloud Firestore** | NoSQL database in Native Mode (`us-central1`), PITR enabled | 1 GB storage, 50,000 reads, 20,000 writes, 20,000 deletes/day | ✅ Supported |
| **Cloud Pub/Sub** | Topics and Push Subscriptions with dedicated invoker Service Account | 10 GB message data transfer per month | ✅ Supported |
| **Identity Platform / Auth** | Email/Password authentication & provider setup | Up to 50,000 Monthly Active Users (MAU) | ✅ Supported |
| **Secret Manager** | Base application secret with automatic replication | 6 active secret versions, 10,000 access operations/month | ✅ Supported |
| **Compute Engine (VM e2-micro)** | e2-micro instance (1 vCPU, 1 GB RAM, 30 GB standard persistent disk) | 1 non-preemptible e2-micro instance/mo (744 continuous hours in US) | ✅ Supported |
| **BigQuery** | Analytical queries and active table storage | 1 TB SQL queries processed/mo, 10 GB active storage | ✅ Supported |
| **Cloud Functions (2nd Gen)** | Event-driven invocations and serverless compute time | 2M invocations/mo, 400,000 GB-seconds | ✅ Supported |
| **Cloud Monitoring & Budgets** | Real-time quota and cost tracking via GCP Billing API | Automated budget alert rules at 50%, 80%, and 100% thresholds | 🚧 In Roadmap |

---

## ⚡ Quickstart

### Prerequisites
* **Go:** Version 1.21 or later ([download Go](https://go.dev/dl/)).
* **Terraform:** Version >= 1.5.0 ([download Terraform](https://developer.hashicorp.com/terraform/install)).
* **Docker & Docker Compose:** Required to run the local offline emulator suite ([download Docker](https://docs.docker.com/get-docker/)).
* **Google Cloud SDK (`gcloud`):** Required only when deploying generated infrastructure to GCP ([install gcloud](https://cloud.google.com/sdk/docs/install)).

### 1. Clone & Build

```bash
# Clone the repository
git clone https://github.com/AlexG695/cloud-cli-starter.git
cd cloud-cli-starter

# Build the CLI binary (output generated at ./bin/cloud)
make build

# (Optional) Run the unit test suite
make test
```

### 2. Install globally (optional)

```bash
# Copy binary to system PATH
sudo cp bin/cloud /usr/local/bin/cloud

# Or install directly using Go toolchain
go install .
```

---

## 💻 Usage & Examples

### CLI Commands

The primary CLI command is **`cloud`** (with the alias **`init`**).

```bash
# Launch the interactive wizard
./bin/cloud

# Or using the alias
./bin/cloud init

# View command description and flags
./bin/cloud --help

# Dry-run mode: Validate configuration without writing files to disk
./bin/cloud --dry-run

# Headless CI/CD mode with JSON configuration file and clean JSON output
./bin/cloud --config project.json --format json --dry-run

# Specify custom target output directory (default: infra)
./bin/cloud --config project.json --output-dir ./my-infra

# Audit existing Terraform infrastructure against Always Free limits
./bin/cloud audit infra

# Audit with JSON report and strict mode for automated CI/CD gates
./bin/cloud audit infra --format json --strict

# Audit with Zero-Trust AI security recommendations
./bin/cloud audit infra --with-ai
```

### Wizard Workflow

Running the command starts an interactive prompt covering three sequential steps:

1. **Project Identification:**
   * `Project name`: Human-readable project name (e.g., `geo-api`).
   * `GCP Project ID`: Target Google Cloud Console Project ID (e.g., `my-project-prod-123`).

2. **Cloud Run Microservices Loop:**
   * Configure $N$ microservices interactively.
   * `Name`: Service identifier (e.g., `api-gateway`, `auth-service`).
   * `Visibility`:
     * *Public*: Internet-facing via `roles/run.invoker` granted to `allUsers`.
     * *Private*: Attack surface reduced; accessible only through IAM or internal load balancers.
   * `Pub/Sub integration`: Automatically provisions a Cloud Pub/Sub Topic and Push Subscription with authentication tokens.

3. **Optional Architecture Modules:**
   * Multi-select checklist (spacebar to toggle):
     * `IncludeStorage`: Private Cloud Storage bucket with versioning and public access prevention.
     * `IncludeFirestore`: Cloud Firestore NoSQL database in Native Mode.
     * `IncludeAuth`: Identity Platform / Firebase Auth.
     * `IncludeComputeEngine`: Always Free e2-micro VM instance (30GB pd-standard, prompts for private vs public IP and idle shutdown).
     * `IncludeBigQuery`: BigQuery dataset and partitioned event table with 60-day retention to strictly cap storage under 10 GB.
     * `IncludeCloudFunctions`: Cloud Functions v2 with scale-to-zero and 7-day source zip auto-purge on GCS bucket.
     * `IncludeLocalEmulators`: Offline Docker Compose suite with Firebase emulators.

### Generated Output

Once completed, the CLI prints the validated configuration in formatted **JSON** and renders all project files into `./infra`:

```json
==================================================
🎉 Configuration successfully generated (JSON):
==================================================
{
  "project_name": "geo-api",
  "gcp_project_id": "my-project-prod-123",
  "region": "us-central1",
  "services": [
    {
      "name": "api-gateway",
      "memory": "256Mi",
      "cpu": "1000m",
      "is_public": true,
      "enable_pubsub": false
    }
  ],
  "include_storage": true,
  "include_firestore": true,
  "include_auth": true,
  "enable_identity_platform": true,
  "include_secrets": true,
  "include_compute_engine": false,
  "vm_public_ip": false,
  "vm_idle_shutdown": false,
  "include_bigquery": false,
  "include_cloud_functions": false,
  "include_local_emulators": true,
  "enable_gemma_assistant": false
}
```

Generated file structure in `./infra`:

```
infra/
├── Makefile              # Terraform and Docker shortcuts
├── compute.tf            # Cloud Run services, Compute Engine VM, Cloud Functions v2 & Service Accounts
├── main.tf               # Required GCP API enablement
├── modules.tf            # Storage, Firestore, Auth, BigQuery, and Secret Manager
├── outputs.tf            # Service URLs, bucket names, and resource identifiers
├── providers.tf          # Google Cloud provider declarations
├── variables.tf          # Project ID and regional variables
├── docker-compose.yml    # (Optional) Firebase Emulator Suite container definition
├── firebase.json         # (Optional) Emulator ports and rules bindings
├── firestore.rules       # (Optional) Permissive rules for local development
└── storage.rules         # (Optional) Storage rules for local development
```

### Managing Generated Infrastructure

Navigate into `./infra` and leverage the generated **`Makefile`**:

```bash
cd infra

# Display all available targets
make help

# Format and validate Terraform configuration
make check

# Initialize Terraform providers
make init

# Preview changes with zero financial cost
make plan

# Apply infrastructure to Google Cloud
make apply

# Start local emulators with Web UI
make emulators-up

# Stop local emulators
make emulators-down

# Destroy all cloud resources to ensure $0 balance
make destroy
```

> **Local Emulator Dashboard:** Running `make emulators-up` provides a complete emulation environment for Firestore, Storage, Pub/Sub, and Auth accessible at **`http://localhost:4000`** without connecting to GCP or consuming quotas.

---

### GCP Authentication Requirements

To provision cloud resources with `make apply`, authenticate your terminal environment:

#### Option 1: Application Default Credentials (Recommended for developers)
```bash
# 1. Log in with user account
gcloud auth login

# 2. Generate local Application Default Credentials (ADC) for Terraform
gcloud auth application-default login

# 3. Set your active GCP project
gcloud config set project YOUR_PROJECT_ID
```

#### Option 2: Service Account (Recommended for CI/CD pipelines)
```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

---

### AI-Assisted Zero-Trust Security Audit (Ollama / Gemini / OpenAI)

The codebase embeds an intelligent security auditor in `internal/ai` that analyzes Terraform files and provides actionable Zero-Trust recommendations. It automatically selects the best available engine:

1. **Local Ollama (Default - 100% Private & Offline):** If Ollama is running at `http://127.0.0.1:11434`, it queries `gemma:2b`:
   ```bash
   ollama run gemma:2b
   ```
2. **Google Gemini (Cloud Fallback):** If Ollama is offline, it detects `GEMINI_API_KEY`:
   ```bash
   export GEMINI_API_KEY="your-gemini-api-key"
   ```
3. **OpenAI (Cloud Fallback):** If Gemini is not configured, it detects `OPENAI_API_KEY`:
   ```bash
   export OPENAI_API_KEY="your-openai-api-key"
   ```
4. **Graceful Degradation:** If no AI backend is detected, generation completes cleanly without failure.

---

### Static Cost & Compliance Audit (`cloud audit`)

The standalone **`cloud audit`** subcommand performs rule-based static analysis directly over any directory containing Terraform (`.tf`) files to ensure that all infrastructure resources strictly comply with Google Cloud's Always Free tier ($0/month):

#### Strict Rule Enforcement:
* **Compute Engine:** Validates machine type is `e2-micro`, boot disk type is `pd-standard` (`pd-ssd` incurs billing), and boot disk size is $\le$ 30 GB.
* **Cloud Run:** Validates container memory allocation is $\le$ 256Mi and enforces scale-to-zero (`min_instance_count = 0`).
* **Cloud Storage:** Validates public access prevention is `enforced` and uniform bucket-level access is enabled.
* **BigQuery:** Validates the presence of automatic table expiration (`default_table_expiration_ms`) to ensure active data storage remains well under the 10 GB Always Free ceiling.
* **Always Free Eligible Regions:** Validates that resources are deployed within free-tier eligible US regions (`us-central1`, `us-east1`, `us-west1`).
* **Zero-Trust IAM:** Verifies that compute services use dedicated, least-privilege Service Accounts rather than the permissive default Compute Engine SA.

#### Supported Flags:
* `-d, --dir <path>`: Directory containing `.tf` files to audit (default: `infra`).
* `-f, --format text|json`: Terminal-friendly report or raw JSON for programmatic inspection in automated pipelines.
* `--strict`: Exits with code `1` upon detecting any cost violation (`FAIL`), ideal for blocking non-compliant Pull Requests in CI/CD.
* `--with-ai`: Invokes the local or cloud LLM provider to enrich findings with Zero-Trust architectural security recommendations.

#### Audit Usage Examples:

```bash
# Standard local audit
./bin/cloud audit infra

# Strict CI/CD quality gate (GitHub Actions / GitLab CI)
./bin/cloud audit infra --format json --strict

# Combined audit: Always Free cost rules + AI-powered Zero-Trust suggestions
./bin/cloud audit infra --with-ai
```

---

## 🗺️ Roadmap

Status of planned features across release cycles:

- [ ] **Active Real-Time Monitoring Subcommands:**
  - [ ] `cloud status` / `gcp-freetier status`: Live inspection of monthly quota consumption against Always Free thresholds via Cloud Monitoring API and Cloud Billing API.
  - [x] `cloud audit` / `gcp-freetier audit`: Standalone command to run security and cost audits over existing Terraform directories.
- [x] **CLI Output Formatting & Automation Flags:**
  - [x] Support for `--format json` and `--format text` for CI/CD automated gates.
  - [x] Support for `--dry-run` to preview configuration without writing to disk.
  - [x] Headless mode via `-c, --config` and configurable destination via `-o, --output-dir`.
- [x] **Extended Always Free Service Coverage:**
  - [x] Compute Engine module: `e2-micro` VM in `us-central1` (public or private IP) with automated idle shutdown.
  - [x] BigQuery module: Partitioned datasets capped within the 10 GB limit with 60-day auto-retention.
  - [x] Cloud Functions v2 module: Scale-to-zero and 7-day source zip auto-purge.
- [ ] **Proactive Billing Guardrails:**
  - [ ] Webhook alerts for Slack, Discord, and Email triggered when reaching 80% of free limits.
- [ ] **Cross-Platform Distribution:**
  - [ ] Precompiled binaries for Linux, macOS, and Windows via GitHub Releases (GoReleaser).
  - [ ] Homebrew Tap distribution (`brew install alexg695/tap/cloud`).

---

## 🤝 Contributing

Contributions from the open-source community are warmly welcomed! To add new GCP Always Free services, enhance Terraform templates, or expand the AI audit engine:

1. Fork the repository.
2. Create a dedicated feature branch:
   ```bash
   git checkout -b feature/my-new-feature
   ```
3. Make your changes and verify that unit tests pass:
   ```bash
   make test
   ```
4. Commit your changes using [Conventional Commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "feat: add BigQuery Always Free tier template"
   ```
5. Open a Pull Request targeting the `main` branch.

---

## 📄 License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for complete details.
