# Kubernetes AI Incident Analyzer — 项目文档

## 1. 项目定位

### 目标

构建一个 AI 驱动的 Kubernetes 故障分析工具，核心指标：**缩短 MTTR（故障恢复时间）**。

### 目标用户

- DevOps / SRE / Platform Engineer
- Kubernetes 运维人员
- 日本企业云团队

### 核心原则

| 这个产品是 | 这个产品不是 |
|-----------|------------|
| AI 驱动的 Kubernetes Incident Intelligence 工具 | 通用 AI 聊天机器人 |
| Kubernetes 故障排查助手 | Kubernetes Dashboard |
| 聚焦根因分析的 CLI 工具 | 全功能 AI Ops 平台 |

---

## 2. MVP 定义

### 第一阶段范围

第一版只做一个场景：

> **CrashLoopBackOff 根因分析**

#### 输入

- Pod logs
- `kubectl describe` 输出
- Events
- Deployment / StatefulSet YAML

#### 输出

| 字段 | 说明 |
|------|------|
| Root Cause | 根因一句话总结 |
| Explanation | 详细原因解释 |
| Risk Level | 风险等级（High / Medium / Low） |
| Suggested Fix | 可操作的修复建议 |

### 优先级排序

| 优先级 | 故障类型 | 复杂度 | 说明 |
|--------|---------|--------|------|
| P0 | ImagePullBackOff | ★☆☆ | 原因确定（镜像名/认证/不存在），分析准确率高，适合首轮 Demo |
| P1 | CrashLoopBackOff | ★★★ | 根因多样（代码 bug / 配置 / 资源限制），核心场景 |
| P2 | Pending Pod | ★★☆ | 资源不足 / PVC 问题 / 节点选择器 |
| P3 | Readiness/Liveness 探针失败 | ★★☆ | 需要理解应用健康检查逻辑 |
| P4 | OOMKilled | ★☆☆ | 相对简单，资源限制分析 |

### 不做清单

第一阶段明确不做：

- SaaS 平台 / 多集群管理 / 复杂前端
- Kubernetes Operator / Cluster Agent / Admission Controller
- AI Agent / 自动修复
- 用户系统 / Billing
- RAG（外部知识库检索）

---

## 3. 产品形态

### 第一阶段：kubectl 插件 / CLI

```bash
kubectl ai analyze pod payment-api
```

**选择理由：**
- DevOps 习惯 CLI，接入成本低
- 企业无需集群内部署，安全顾虑少
- 使用用户现有 kubeconfig 权限，不做权限提升

### 架构

```text
用户集群（只读权限）
     │  kubectl / client-go
     ▼
kubectl-ai CLI
     │  收集 → 构建 Prompt → 调用 AI
     ▼
AWS Bedrock ──→ Claude Sonnet
     │
     ▼
返回: Root Cause / Explanation / Risk / Fix
```

### 关键设计决策

- AI 本身不持有 Kubernetes 权限，**完全复用用户 RBAC**
- MVP 仅需只读权限：`get / list / watch`（pods, logs, events, deployments, services, ingresses）
- **不要请求 cluster-admin**，这会成为企业采纳的阻力

---

## 4. 技术栈

| 层次 | 选型 | 理由 |
|------|------|------|
| 语言 | Go | Kubernetes 生态标准，CLI 友好，性能好 |
| CLI 框架 | Cobra | Go CLI 事实标准 |
| K8s SDK | client-go | 结构化数据，原生 K8s API（MVP 阶段可选 os/exec kubectl 加速） |
| AI 模型 | Claude Sonnet via AWS Bedrock | YAML/K8s 理解强，长上下文适合日志分析 |
| AI 客户端 | 接口抽象（支持 Bedrock / Anthropic API 切换） | PoC 阶段 Anthropic API 直连更快，Bedrock 留作企业生产方案 |

### AI 方案说明

**不训练自己的模型。** 直接使用 Claude Sonnet。

选择 Bedrock 的原因（面向日本企业市场）：
- AWS 集成 / IAM 权限 / 东京 Region 可用
- 合规友好，不需要 GPU 和模型托管

**AI 客户端设计：** 定义接口，先实现 Anthropic Direct API 快速跑通 POC，再实现 Bedrock 满足企业诉求。

---

## 5. 真正的技术壁垒

```
壁垒不在 LLM，而在：
│
├── 上下文收集策略 — 什么时候取什么数据，取多少
├── Kubernetes 故障模式库 — 积累真实故障模式
├── Prompt 工程 — 如何让 AI 输出准确、可操作的根因分析
└── 依赖关系分析 — 跨资源（Pod/Service/Ingress）的关联推理
```

### Prompt 策略：两层结构

```text
第一层：收集数据 → 构建 Incident Context（事实层）
  ├── Pod logs（最近 N 行）
  ├── Events（异常 Event）
  ├── Describe output（状态快照）
  └── 相关 YAML（资源定义）

第二层：分析 Context → 输出 Root Cause（推理层）
  └── Claude 基于事实推理，返回结构化分析
```

---

## 6. 第一阶段实现计划

### Day 1：项目初始化

- `go mod init` + Cobra 命令骨架
- AI Client 接口定义 + Anthropic API 实现
- 端到端打通：CLI 接收参数 → 收集数据 → 调用 AI → 输出结果

### Day 2-3：ImagePullBackOff 场景

- Pod 数据收集模块
- 场景特化的 Prompt 模板
- 结果输出格式化

### Day 4-5：CrashLoopBackOff 场景

- CrashLoopBackOff Prompt 模板
- 日志截取策略（最近崩溃周期的日志）
- 多场景路由

### Day 6-7：完善与发布

- CLI 输出美化
- 错误处理（kubeconfig 未找到、网络超时、AI 调用失败）
- README / GitHub Release

### 成功标准

- 能正确分析该场景的根因（验证 3-5 个真实案例）
- 分析结果对用户有实际帮助（不是"你的容器崩溃了"这种废话）
- CLI 在全新环境下 5 分钟内可用

---

## 7. 产品路线图

| 阶段 | 形态 | 核心能力 |
|------|------|---------|
| 第一阶段 | CLI | CrashLoopBackOff 等 5 种故障分析 |
| 第二阶段 | Web UI / SaaS | 可视化分析 + 历史记录 |
| 第三阶段 | VSCode 插件 | YAML 检查 / Probe 建议 / Inline AI 提示 |
| 第四阶段 | AI SRE Agent | 自动诊断 / 回滚 / 修复执行 |

---

## 8. 核心信念

> 未来真正的价值不是"用了 AI"，而是**比别人更懂 Kubernetes 故障。**

产品方向始终围绕 Incident Intelligence（故障智能分析），而不是做一个通用 AI 聊天机器人。
