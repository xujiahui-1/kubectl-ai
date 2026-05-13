# kubectl-ai

> AI 驱动的 Kubernetes 故障分析工具 — kubectl 插件 + 独立 CLI

快速定位 Pod 故障根因，支持 ImagePullBackOff、CrashLoopBackOff、OOMKilled 等常见场景。

## 快速开始

```bash
# 1. 交互式初始化（选择 AI 提供商、输入 API Key、选模型）
kubectl-ai init

# 2. 分析故障 Pod
kubectl-ai analyze pod payment-api -n production

# 或批量扫描整个 namespace
kubectl-ai analyze all -n production
```

> 也可以跳过 init，直接使用环境变量：
>
> **DeepSeek（默认）**
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

### 安装

```bash
# 下载二进制
curl -LO https://github.com/xujiahui-1/kubectl-ai/releases/latest/download/kubectl-ai-darwin-arm64
chmod +x kubectl-ai-darwin-arm64
sudo mv kubectl-ai-darwin-arm64 /usr/local/bin/kubectl-ai

# 验证
kubectl-ai --help
# 也支持 kubectl plugin 方式
kubectl ai analyze pod xxx
```

## 支持的故障场景

| 场景 | 说明 |
|------|------|
| ImagePullBackOff | 镜像拉取失败（名称错误/认证/限速） |
| CrashLoopBackOff | 容器反复崩溃（应用错误/配置/依赖） |
| OOMKilled | 内存超限 |
| Pending Pod | 调度失败（资源不足/节点选择器/PVC） |
| Init Container Error | 初始化容器失败 |
| Probe Failure | 健康检查探针失败（端口/路径/超时） |

## 工作原理

```text
kubectl-ai analyze pod xxx
  → 读取 ~/.kube/config（复用你的 RBAC 权限）
  → 收集 Pod 状态 / 日志 / Events / 父级资源
  → 自动识别故障场景
  → 构建场景特化的 Prompt 发送给 AI
  → 返回结构化分析结果（根因 / 解释 / 风险 / 修复建议）
```

- **不高权限** — 使用你现有的 kubeconfig 凭证，不做权限提升
- **不写集群** — 只读操作，不做任何更改
- **不常驻** — 命令执行完即退出，没有后台进程

## 配置

### 交互式初始化（推荐）

运行 `kubectl-ai init` 交互式选择 AI 提供商、输入 API Key 和选择模型。
配置保存到 `~/.kubectl-ai/config.json`，后续使用无需再指定参数。

### 命令行参数

| 参数 | 说明 |
|------|------|
| `--ai-provider` | AI 提供商 (`deepseek`, `anthropic`, `openai`) — 优先使用配置文件，默认 `deepseek` |
| `--ai-model` | 模型名称 — 优先使用配置文件，各提供商有各自默认值 |
| `--ai-api-key` | API Key — 优先级：命令行 > 配置文件 > 环境变量 |
| `-n, --namespace` | Kubernetes namespace（默认 `default`） |
| `--kubeconfig` | kubeconfig 路径（默认 `~/.kube/config`） |

优先级：命令行参数 > 配置文件 (`~/.kubectl-ai/config.json`) > 环境变量 > 内置默认值。

### 手动配置

```bash
# DeepSeek（默认）
export DEEPSEEK_API_KEY=sk-xxx
kubectl-ai analyze pod xxx -n production

# OpenAI
export OPENAI_API_KEY=sk-xxx
kubectl-ai analyze pod xxx -n production --ai-provider openai --ai-model gpt-4o

# Anthropic
export ANTHROPIC_API_KEY=sk-ant-xxx
kubectl-ai analyze pod xxx -n production --ai-provider anthropic --ai-model claude-sonnet-4-20250514
```

## 系统要求

- Go 1.26+（编译需要）
- 可用的 `~/.kube/config`（运行需要）
- 操作系统：macOS / Linux

## License

MIT
