# kubectl-ai

> AI 驱动的 Kubernetes 故障分析工具 — kubectl 插件 + 独立 CLI

快速定位 Pod 故障根因，支持 ImagePullBackOff、CrashLoopBackOff、OOMKilled 等常见场景。

## 快速开始

选择一种 AI 提供商：

**选项 A — DeepSeek（默认）**
```bash
export DEEPSEEK_API_KEY=sk-xxx
```

**选项 B — Anthropic**
```bash
export ANTHROPIC_API_KEY=sk-ant-xxx
```

然后运行：
```bash
# 分析单个故障 Pod（默认用 DeepSeek）
kubectl-ai analyze pod payment-api -n production

# 如果用 Anthropic
kubectl-ai analyze pod payment-api -n production --ai-provider anthropic

# 或批量扫描整个 namespace
kubectl-ai analyze all -n production
```

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

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--ai-provider` | `deepseek` | AI 提供商 (`deepseek`, `anthropic`, `bedrock`) |
| `--ai-model` | `deepseek-chat` | 模型名称 |
| `--ai-api-key` | 环境变量 | 默认读取 `$DEEPSEEK_API_KEY` |
| `-n, --namespace` | `default` | Kubernetes namespace |
| `--kubeconfig` | `~/.kube/config` | kubeconfig 路径 |

### 使用其他 AI 提供商

```bash
# Anthropic
export ANTHROPIC_API_KEY=sk-ant-xxx
kubectl-ai analyze pod xxx --ai-provider anthropic --ai-model claude-sonnet-4-20250514
```

## 系统要求

- Go 1.26+（编译需要）
- 可用的 `~/.kube/config`（运行需要）
- 操作系统：macOS / Linux

## License

MIT
