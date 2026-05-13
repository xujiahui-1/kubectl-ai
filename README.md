
# kubectl-ai

> AI-powered Kubernetes incident analyzer — kubectl plugin + standalone CLI

Quickly identify pod failure root causes: ImagePullBackOff, CrashLoopBackOff, OOMKilled, and more.

## Quick Start

```bash
# 1. Interactive setup (provider, API key, model)
kubectl-ai init

# 2. Analyze a failing pod
kubectl-ai analyze pod payment-api -n production

# Or scan all failing pods in a namespace
kubectl-ai analyze all -n production
```

> Skip the interactive setup by setting environment variables and using flags:
>
> **DeepSeek（default）**
> ```bash
> export DEEPSEEK_API_KEY=sk-xxx
> kubectl-ai analyze pod payment-api -n production
> ```
>
> **OpenAI**
> ```bash
> export OPENAI_API_KEY=sk-xxx
> kubectl-ai analyze pod payment-api -n production --ai-provider openai --ai-model gpt-4o
> ```
>
> **Anthropic**
> ```bash
> export ANTHROPIC_API_KEY=sk-ant-xxx
> kubectl-ai analyze pod payment-api -n production --ai-provider anthropic --ai-model claude-sonnet-4-20250514
> ```

### Installation

```bash
# Download the binary
curl -LO https://github.com/xujiahui-1/kubectl-ai/releases/latest/download/kubectl-ai-darwin-arm64
chmod +x kubectl-ai-darwin-arm64
sudo mv kubectl-ai-darwin-arm64 /usr/local/bin/kubectl-ai

# Verify
kubectl-ai --help
# Also works as a kubectl plugin
kubectl ai analyze pod xxx
```

## Supported Scenarios

| Scenario | Description |
|----------|-------------|
| ImagePullBackOff | Image pull failure (wrong name/auth/rate limit) |
| CrashLoopBackOff | Container keeps crashing (app error/config/dependency) |
| OOMKilled | Memory limit exceeded |
| Pending Pod | Scheduling failure (resources/node selector/PVC) |
| Init Container Error | Init container failed |
| Probe Failure | Health check probe failing (port/path/timeout) |

## How It Works

```text
kubectl-ai analyze pod xxx
  → Reads ~/.kube/config (reuses your RBAC permissions)
  → Collects pod status / logs / events / parent resource
  → Detects failure scenario automatically
  → Builds a scenario-specific prompt and calls the AI
  → Returns structured result (root cause / explanation / risk / fix)
```

- **No privilege escalation** — uses your existing kubeconfig credentials
- **Read-only** — never modifies the cluster
- **No daemon** — exits immediately after analysis

## Configuration

### Interactive Setup (Recommended)

Run `kubectl-ai init` to configure your AI provider, API key, and model interactively.
Settings are saved to `~/.kubectl-ai/config.json`.

### CLI Flags

| Flag | Description |
|------|-------------|
| `--ai-provider` | AI provider (`deepseek`, `anthropic`, `openai`) — uses config file or defaults to `deepseek` |
| `--ai-model` | Model name — uses config file or provider-specific default |
| `--ai-api-key` | AI API key — falls back to config file, then environment variable |
| `-n, --namespace` | Kubernetes namespace (default `default`) |
| `--kubeconfig` | Path to kubeconfig (default `~/.kube/config`) |

Priority: CLI flag > config file (`~/.kubectl-ai/config.json`) > environment variable > built-in default.

### Manual Setup

```bash
# DeepSeek (default)
export DEEPSEEK_API_KEY=sk-xxx
kubectl-ai analyze pod xxx -n production

# OpenAI
export OPENAI_API_KEY=sk-xxx
kubectl-ai analyze pod xxx -n production --ai-provider openai --ai-model gpt-4o

# Anthropic
export ANTHROPIC_API_KEY=sk-ant-xxx
kubectl-ai analyze pod xxx -n production --ai-provider anthropic --ai-model claude-sonnet-4-20250514
```

## Requirements

- Go 1.26+ (to build)
- A valid `~/.kube/config` (to run)
- Platforms: macOS / Linux

## License

MIT
