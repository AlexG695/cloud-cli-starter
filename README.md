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
| **Cloudflare DNS / SSL** | DNS records integration and SSL proxy for custom domains | Free Tier plan for custom domain management | ✅ Supported |
| **Compute Engine (VM e2-micro)** | e2-micro instance (1 vCPU, 1 GB RAM, 30 GB standard persistent disk) | 1 non-preemptible e2-micro instance/mo (744 continuous hours in US) | 🚧 In Roadmap |
| **BigQuery** | Analytical queries and active table storage | 1 TB SQL queries processed/mo, 10 GB active storage | 🚧 In Roadmap |
| **Cloud Functions (2nd Gen)** | Event-driven invocations and serverless compute time | 2M invocations/mo, 400,000 GB-seconds | 🚧 In Roadmap |
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
     * `IncludeCloudflare`: Cloudflare DNS & SSL proxy (prompts for custom domain).
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
  "include_cloudflare": false,
  "include_local_emulators": true,
  "enable_gemma_assistant": false
}
```

Generated file structure in `./infra`:

```
infra/
├── Makefile              # Terraform and Docker shortcuts
├── compute.tf            # Cloud Run services & dedicated Service Accounts
├── main.tf               # Required GCP API enablement
├── modules.tf            # Storage, Firestore, Auth, and Secret Manager
├── outputs.tf            # Service URLs, bucket names, and resource identifiers
├── providers.tf          # Google Cloud and Cloudflare provider declarations
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

## 🗺️ Roadmap

Planned features for upcoming releases:

- [ ] **Active Real-Time Monitoring Subcommands:**
  - `cloud status` / `gcp-freetier status`: Live inspection of monthly quota consumption against Always Free thresholds via Cloud Monitoring API and Cloud Billing API.
  - `cloud audit` / `gcp-freetier audit`: Standalone command to run security and cost audits over existing Terraform directories.
- [ ] **CLI Output Formatting Flags:**
  - Support for `--format json` and `--format text` for CI/CD automated gates.
  - Support for `--dry-run` to preview configuration without writing to disk.
- [ ] **Extended Always Free Service Coverage:**
  - Compute Engine module: `e2-micro` VM in `us-central1` with automated idle shutdown.
  - BigQuery module: Partitioned datasets capped within the 10 GB limit.
  - Cloud Functions v2 module.
- [ ] **Proactive Billing Guardrails:**
  - Webhook alerts for Slack, Discord, and Email triggered when reaching 80% of free limits.
- [ ] **Cross-Platform Distribution:**
  - Precompiled binaries for Linux, macOS, and Windows via GitHub Releases (GoReleaser).
  - Homebrew Tap distribution (`brew install alexg695/tap/cloud`).

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
