
# kubectl-ai

> AI-powered Kubernetes incident analyzer — kubectl plugin + standalone CLI

Quickly identify pod failure root causes: ImagePullBackOff, CrashLoopBackOff, OOMKilled, and more.

## Quick Start

```bash
# 1. Set your AI API key
export DEEPSEEK_API_KEY=sk-xxx

# 2. Analyze a failing pod
kubectl-ai analyze pod payment-api -n production

# 3. Or scan all failing pods in a namespace
kubectl-ai analyze all -n production
```

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

| Flag | Default | Description |
|------|---------|-------------|
| `--ai-provider` | `deepseek` | AI provider (`deepseek`, `anthropic`, `bedrock`) |
| `--ai-model` | `deepseek-chat` | Model name |
| `--ai-api-key` | env var | Defaults to `$DEEPSEEK_API_KEY` |
| `-n, --namespace` | `default` | Kubernetes namespace |
| `--kubeconfig` | `~/.kube/config` | Path to kubeconfig |

### Using Other AI Providers

```bash
# Anthropic
export ANTHROPIC_API_KEY=sk-ant-xxx
kubectl-ai analyze pod xxx --ai-provider anthropic --ai-model claude-sonnet-4-20250514
```

## Requirements

- Go 1.26+ (to build)
- A valid `~/.kube/config` (to run)
- Platforms: macOS / Linux

## License

MIT
