# GCP Free Tier Monitor (CLI)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](go.mod)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/AlexG695/cloud-cli-starter/pulls)
[![CI Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
[![GCP Cost](https://img.shields.io/badge/GCP_Cost-%240%2Fmonth-success.svg)](#)

[English](README.md) | Español

**Previene facturas sorpresa en Google Cloud Platform desplegando arquitecturas serverless preconfiguradas para permanecer estrictamente dentro del Always Free Tier ($0/mes), con emulación local y auditoría de seguridad asistida por IA.**

![Demo](docs/assets/demo.gif)

---

## 📌 Por qué existe (The Problem vs. The Solution)

- **Costos fantasma e instancias en reposo:** Muchos desarrolladores, estudiantes y startups reciben facturas inesperadas a fin de mes por recursos "olvidados", cuotas mínimas de cómputo mal dimensionadas o servicios desplegados en regiones fuera de la cobertura de la capa gratuita de GCP.
- **Complejidad y dispersión de las reglas Always Free:** Comprender los límites exactos de Google Cloud (ej. `us-central1`, 256MiB de RAM por microservicio en Cloud Run, escala a cero obligatoria, cuotas mensuales de peticiones y buckets regionales con versionado) requiere horas de análisis y es susceptible a fallas humanas en la consola web.
- **Arquitectura segura por diseño antes del cobro:** Esta herramienta resuelve el problema de raíz: mediante un wizard interactivo en terminal, genera infraestructura como código (Terraform) precalculada bajo los límites de costo $0, levanta un entorno de emulación local en Docker para pruebas sin conexión a la nube y audita configuraciones con modelos de IA para aplicar Zero-Trust y mínimo privilegio.

---

## 🚀 Características Principales (Key Features)

- **Terminal Wizard Interactivo:** Asistente interactivo guiado por consola (construido con `survey`) que solicita variables clave, configura múltiples microservicios Cloud Run y selecciona módulos opcionales en segundos.
- **Garantía $0/Mes (Always Free Defaults):** Ajuste estricto por defecto en la región elegible `us-central1`, límite de memoria por contenedor de `256Mi`, CPU `1000m` (1 vCPU) y política de escalado a cero (`min_instance_count = 0`).
- **Infraestructura como Código (IaC) Modular:** Genera templates limpios y validados de Terraform (`main.tf`, `compute.tf`, `modules.tf`, `variables.tf`, `outputs.tf`, `providers.tf`) listos para versionar y desplegar.
- **Identidad y Mínimo Privilegio (Zero-Trust IAM):** Desacopla la Service Account por defecto de Compute Engine y genera Service Accounts dedicadas por cada microservicio con permisos granulares específicos (`roles/datastore.user`, `roles/storage.objectUser`, `roles/secretmanager.secretAccessor`).
- **Suite de Emulación Local (Docker Compose):** Aprovisionamiento automático de `docker-compose.yml`, `firebase.json` y reglas de seguridad para levantar emuladores de Cloud Firestore, Cloud Storage, Cloud Pub/Sub y Auth/Identity Platform con un panel web interactivo en `http://localhost:4000`.
- **Motor de Auditoría de Seguridad con IA:** Módulo interno de auditoría Zero-Trust con soporte local nativo para Ollama (`gemma:2b`), fallback automático a Google Gemini (`gemini-1.5-flash`), OpenAI (`gpt-4o-mini`) y degradación suave (_graceful degradation_) si no hay proveedor disponible.
- **Automatización con Makefile:** Automatiza el ciclo de vida del proyecto generado: formateo (`make fmt`), validación de sintaxis (`make check`), despliegue (`make apply`), destrucción (`make destroy`) y control de emuladores (`make emulators-up` / `make emulators-down`).

---

## 📊 Matriz de Servicios Soportados (Free Tier Coverage)

| Servicio GCP                     | Métrica Monitoreada / Aprovisionada                                     | Límite Free Tier (Always Free)                                                    | Estado en MVP |
| :------------------------------- | :---------------------------------------------------------------------- | :-------------------------------------------------------------------------------- | :-----------: |
| **Cloud Run**                    | Memoria (256Mi), CPU (1 vCPU), Escala a 0 (`min_instance_count = 0`)    | 2M peticiones/mes, 360,000 GB-s, 180,000 vCPU-s (us-central1)                     | ✅ Soportado  |
| **Cloud Storage**                | Bucket regional en `us-central1`, acceso uniforme y privado forzado     | 5 GB-mes almacenamiento estándar, 5,000 op. Clase A, 50,000 Clase B               | ✅ Soportado  |
| **Cloud Firestore**              | Base de datos NoSQL en Modo Nativo (`us-central1`), PITR activo         | 1 GB almacenamiento, 50,000 lecturas, 20,000 escrituras, 20,000 eliminaciones/día | ✅ Soportado  |
| **Cloud Pub/Sub**                | Tópicos y Push Subscriptions con Service Account invocadora             | 10 GB de datos transferidos en mensajes por mes                                   | ✅ Soportado  |
| **Identity Platform / Auth**     | Autenticación con email/password y configuración de proveedor           | Hasta 50,000 usuarios activos mensuales (MAU)                                     | ✅ Soportado  |
| **Secret Manager**               | Secreto base de aplicación con replicación automática                   | 6 versiones activas de secretos, 10,000 operaciones de acceso/mes                 | ✅ Soportado  |
| **Cloudflare DNS / SSL**         | Integración de registros DNS y proxy SSL para dominios personalizados   | Nivel gratuito de Cloudflare DNS                                                  | ✅ Soportado  |
| **Compute Engine (VM e2-micro)** | Instancia e2-micro (1 vCPU, 1 GB RAM, 30 GB disco persistente estándar) | 1 instancia e2-micro no interrumpible/mes (744 horas continuas en US)             | 🚧 En Roadmap |
| **BigQuery**                     | Consultas analíticas y almacenamiento de tablas activas                 | 1 TB de consultas SQL/mes, 10 GB de almacenamiento activo                         | 🚧 En Roadmap |
| **Cloud Functions (2nd Gen)**    | Invocaciones por evento y tiempo de cómputo serverless                  | 2M de invocaciones/mes, 400,000 GB-segundos                                       | 🚧 En Roadmap |
| **Cloud Monitoring & Budgets**   | Monitoreo en tiempo real de cuotas contra GCP Billing API               | Alertas de presupuesto automáticas a los umbrales del 50%, 80% y 100%             | 🚧 En Roadmap |

---

## ⚡ Instalación Rápida (Quickstart)

### Prerrequisitos

- **Go:** Versión 1.25 o superior ([descargar Go](https://go.dev/dl/)).
- **Terraform:** Versión >= 1.5.0 ([descargar Terraform](https://developer.hashicorp.com/terraform/install)).
- **Docker y Docker Compose:** Requerido para ejecutar la suite de emuladores locales ([descargar Docker](https://docs.docker.com/get-docker/)).
- **Google Cloud SDK (`gcloud`):** Requerido únicamente para desplegar la infraestructura generada en GCP ([descargar gcloud](https://cloud.google.com/sdk/docs/install)).

### 1. Clonar y compilar

```bash
# Clonar el repositorio
git clone https://github.com/AlexG695/cloud-cli-starter.git
cd cloud-cli-starter

# Compilar el binario de la CLI (se genera en ./bin/cloud)
make build

# (Opcional) Ejecutar suite de pruebas unitarias
make test
```

### 2. Instalar globalmente en tu sistema (opcional)

```bash
# Copiar al PATH del sistema
sudo cp bin/cloud /usr/local/bin/cloud

# O compilar directamente con el toolchain de Go
go install .
```

---

## 💻 Guía de Uso (Usage & Examples)

### Comandos de la CLI

El comando principal de la CLI es **`cloud`** (con el alias **`init`**).

```bash
# Ejecutar el asistente interactivo
./bin/cloud

# O usando el alias
./bin/cloud init

# Ver la ayuda y descripción de la herramienta
./bin/cloud --help
```

### Flujo de Ejecución del Asistente

Al ejecutar la CLI, se inicia el wizard interactivo en consola con las siguientes etapas:

1. **Identificación del Proyecto:**
   - `Project name`: Nombre legible del proyecto (ej. `geo-api`).
   - `GCP Project ID`: ID del proyecto en Google Cloud Console (ej. `mi-proyecto-prod-123`).

2. **Microservicios Cloud Run:**
   - Configuración de $N$ servicios en bucle.
   - `Name`: Identificador del microservicio (ej. `api-gateway`, `auth-service`).
   - `Visibility`:
     - _Public_: Acceso directo por internet vía `roles/run.invoker` para `allUsers`.
     - _Private_: Tráfico restringido vía IAM con Service Account dedicada.
   - `Pub/Sub integration`: Pregunta si el servicio consumirá eventos mediante un Topic y Push Subscription dedicado.

3. **Módulos de Arquitectura Opcionales:**
   - Selector múltiple (con barra espaciadora):
     - `IncludeStorage`: Bucket privado en Cloud Storage con prevención de acceso público.
     - `IncludeFirestore`: Base de datos NoSQL Cloud Firestore en Modo Nativo.
     - `IncludeAuth`: Identity Platform / Firebase Auth.
     - `IncludeCloudflare`: Configuración de DNS/SSL (solicita dominio opcional).
     - `IncludeLocalEmulators`: Generación de entorno Docker Compose local.

### Salida Generada

Tras completar el asistente, la CLI imprime en consola la configuración serializada en formato **JSON** para auditoría y crea el directorio `./infra`:

```json
==================================================
🎉 Configuration successfully generated (JSON):
==================================================
{
  "project_name": "geo-api",
  "gcp_project_id": "mi-proyecto-prod-123",
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

Estructura de archivos generada en `./infra`:

```
infra/
├── Makefile              # Atajos de comandos Terraform y Docker
├── compute.tf            # Definición de Cloud Run y Service Accounts
├── main.tf               # Activación de APIs requeridas de GCP
├── modules.tf            # Storage, Firestore, Auth y Secret Manager
├── outputs.tf            # URLs de servicios, nombres de buckets y recursos
├── providers.tf          # Configuración de proveedores Google y Cloudflare
├── variables.tf          # Variables de proyecto y región
├── docker-compose.yml    # (Opcional) Contenedor de Firebase Emulator Suite
├── firebase.json         # (Opcional) Configuración de puertos de emuladores
├── firestore.rules       # (Opcional) Reglas permisivas para entorno local
└── storage.rules         # (Opcional) Reglas de storage para entorno local
```

### Gestión de la Infraestructura Generada

Desplázate al directorio `./infra` para gestionar tu entorno con el **`Makefile`**:

```bash
cd infra

# Ver todos los comandos disponibles
make help

# Validar y formatear la sintaxis de Terraform
make check

# Inicializar proveedores de Terraform
make init

# Previsualizar cambios sin costo
make plan

# Aplicar los cambios en Google Cloud Platform
make apply

# Levantar emuladores locales con interfaz web
make emulators-up

# Detener los emuladores locales
make emulators-down

# Destruir toda la infraestructura en la nube para asegurar $0
make destroy
```

> **Dashboard de Emuladores Locales:** Al ejecutar `make emulators-up`, puedes interactuar con Firestore, Storage, Pub/Sub y Auth desde tu navegador en **`http://localhost:4000`** sin conectarte a GCP ni gastar cuotas.

---

### Requisitos de Autenticación de GCP

Para aplicar los recursos en Google Cloud con `make apply`, asegúrate de autenticar tu entorno local:

#### Opción 1: Credenciales por defecto de aplicación (Recomendado para desarrolladores)

```bash
# 1. Autenticar sesión de usuario
gcloud auth login

# 2. Generar credenciales locales para Terraform (ADC)
gcloud auth application-default login

# 3. Establecer el proyecto activo
gcloud config set project TU_PROJECT_ID
```

#### Opción 2: Cuenta de Servicio (CI/CD o entornos automatizados)

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/ruta/a/service-account-key.json"
```

---

### Auditoría de Seguridad Zero-Trust con IA (Ollama / Gemini / OpenAI)

El proyecto incluye un motor inteligente de análisis de seguridad en `internal/ai` que evalúa las configuraciones generadas y sugiere mejoras de aislamiento y Least Privilege. Detecta automáticamente el proveedor disponible:

1. **Ollama local (Predeterminado - Privacidad 100%):** Si Ollama está ejecutándose en `http://127.0.0.1:11434`, utiliza el modelo liviano `gemma:2b`:
   ```bash
   ollama run gemma:2b
   ```
2. **Google Gemini (Fallback Cloud):** Si Ollama no está activo, detecta automáticamente la variable de entorno:
   ```bash
   export GEMINI_API_KEY="tu-api-key-de-gemini"
   ```
3. **OpenAI (Fallback Cloud):** Si Gemini no está configurado, utiliza:
   ```bash
   export OPENAI_API_KEY="tu-api-key-de-openai"
   ```
4. **Degradación Suave (_Graceful Degradation_):** Si ningún motor de IA está activo, el sistema continúa sin interrumpir la generación de archivos.

---

## 🗺️ Hoja de Ruta (Roadmap)

Próximas funcionalidades en desarrollo planificadas para las siguientes versiones:

- [ ] **Comandos de Monitoreo Dinámico en Tiempo Real:**
  - `cloud status` / `gcp-freetier status`: Consulta en tiempo real del consumo acumulado en el mes contra las cuotas Always Free vía Cloud Monitoring API y Cloud Billing API.
  - `cloud audit` / `gcp-freetier audit`: Subcomando CLI dedicado para ejecutar auditorías estáticas de costo y seguridad sobre carpetas de Terraform existentes.
- [ ] **Soporte de Flags CLI de Salida:**
  - Flags `--format json` y `--format text` para integración en pipelines de CI/CD.
  - Flag `--dry-run` para previsualizar la configuración sin escribir en disco.
- [ ] **Ampliación de Cobertura de Servicios:**
  - Módulo para Compute Engine: Instancia `e2-micro` en `us-central1` con apagado programado.
  - Módulo para BigQuery: Conjunto de datos y tablas particionadas con límite de 10 GB.
  - Módulo para Cloud Functions v2.
- [ ] **Alertas Preventivas de Facturación (Cost Guardrails):**
  - Webhooks de alerta para Slack, Discord o Email cuando el consumo alcance el 80% del límite Always Free.
- [ ] **Distribución y Empaquetado:**
  - Publicación de binarios multi-plataforma (Linux, macOS, Windows) mediante GitHub Releases y GoReleaser.
  - Instalación mediante Homebrew Tap (`brew install alexg695/tap/cloud`).

---

## 🤝 Contribución

¡Las contribuciones de la comunidad son bienvenidas! Si deseas agregar soporte para nuevos servicios Always Free de GCP, optimizar los templates de Terraform o mejorar el motor de auditoría con IA:

1. Haz un Fork de este repositorio.
2. Crea tu rama con un nombre descriptivo:
   ```bash
   git checkout -b feature/nueva-funcionalidad
   ```
3. Realiza tus cambios y verifica que las pruebas unitarias y de formato pasen satisfactoriamente:
   ```bash
   make test
   ```
4. Haz commit de tus cambios siguiendo la convención de [Conventional Commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "feat: add BigQuery Always Free tier template"
   ```
5. Envía tu Pull Request a la rama `main`.

---

## 📄 Licencia

Este proyecto está bajo la Licencia **MIT**. Consulta el archivo [LICENSE](LICENSE) para más información.
