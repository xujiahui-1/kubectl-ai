# Kubernetes AI Incident Analyzer — Project Document

## 1. Product Vision

### Goal

Build an AI-powered Kubernetes incident analysis tool to **reduce MTTR (Mean Time To Recovery)**.

### Target Users

- DevOps / SRE / Platform Engineers
- Kubernetes operators
- Japanese enterprise cloud teams

### Core Positioning

| This is | This is NOT |
|---------|-------------|
| AI-powered Kubernetes Incident Intelligence tool | Generic AI chatbot |
| Kubernetes troubleshooting assistant | Kubernetes Dashboard |
| CLI tool focused on root cause analysis | Full AI Ops platform |

---

## 2. MVP Definition

### Phase 1 Scope

The first version focuses on ONE scenario:

> **CrashLoopBackOff root cause analysis**

#### Inputs

- Pod logs
- `kubectl describe` output
- Events
- Deployment / StatefulSet YAML

#### Outputs

| Field | Description |
|-------|-------------|
| Root Cause | One-line root cause summary |
| Explanation | Detailed cause explanation |
| Risk Level | High / Medium / Low |
| Suggested Fix | Actionable remediation steps |

### Priority Order

| Priority | Failure Type | Complexity | Notes |
|----------|-------------|------------|-------|
| P0 | ImagePullBackOff | ★☆☆ | Deterministic causes (image name / auth / not found), high accuracy, good for first demo |
| P1 | CrashLoopBackOff | ★★★ | Diverse root causes (code bug / config / resource limits), core scenario |
| P2 | Pending Pod | ★★☆ | Insufficient resources / PVC issues / node selector |
| P3 | Probe failures | ★★☆ | Requires understanding app health check logic |
| P4 | OOMKilled | ★☆☆ | Simple resource limit analysis |

### Out of Scope (Phase 1)

Explicitly NOT building:

- SaaS platform / Multi-cluster management / Complex frontend
- Kubernetes Operator / Cluster Agent / Admission Controller
- AI Agent / Auto-remediation
- User system / Billing
- RAG (external knowledge base retrieval)

---

## 3. Product Form

### Phase 1: kubectl plugin / CLI

```bash
kubectl ai analyze pod payment-api
```

**Rationale:**
- DevOps prefer CLI, minimal adoption friction
- No in-cluster deployment needed, fewer security concerns
- Uses the user's existing kubeconfig — no privilege escalation

### Architecture

```text
User Cluster (read-only)
     │  kubectl / client-go
     ▼
kubectl-ai CLI
     │  collect → build prompt → call AI
     ▼
AWS Bedrock ──→ Claude Sonnet
     │
     ▼
Returns: Root Cause / Explanation / Risk / Fix
```

### Key Design Decisions

- AI NEVER holds Kubernetes permissions — **fully reuses user RBAC**
- MVP needs read-only: `get / list / watch` (pods, logs, events, deployments, services, ingresses)
- **Do NOT request cluster-admin** — it blocks enterprise adoption

---

## 4. Tech Stack

| Layer | Choice | Rationale |
|-------|--------|-----------|
| Language | Go | Kubernetes ecosystem standard, CLI-friendly, performant |
| CLI Framework | Cobra | De facto standard for Go CLIs |
| K8s SDK | client-go | Structured data, native K8s API (MVP may use os/exec kubectl for speed) |
| AI Model | Claude Sonnet via AWS Bedrock | Strong YAML/K8s understanding, long context for log analysis |
| AI Client | Abstracted interface (supports Bedrock / Anthropic API) | Anthropic API faster for PoC, Bedrock for enterprise production |

### AI Strategy

**Do NOT train your own model.** Use Claude Sonnet directly.

Why Bedrock (for Japanese enterprise market):
- AWS integration / IAM-based auth / Tokyo Region availability
- Compliance-friendly, no GPU or model hosting required

**Client design principle:** Define an interface. Implement Anthropic Direct API first for rapid PoC, then add Bedrock for enterprise requirements.

---

## 5. Technical Moat

```
The moat is NOT the LLM — it's:
│
├── Context collection strategy — what data to fetch, when, how much
├── Kubernetes failure pattern library — real incident patterns
├── Prompt engineering — getting accurate, actionable analysis from AI
└── Dependency analysis — cross-resource reasoning (Pod/Service/Ingress)
```

### Prompt Strategy: Two-Layer Structure

```text
Layer 1: Collect → Build Incident Context (facts)
  ├── Pod logs (recent N lines)
  ├── Abnormal Events
  ├── Describe output (status snapshot)
  └── Related YAML (resource definitions)

Layer 2: Analyze Context → Output Root Cause (reasoning)
  └── Claude reasons from facts, returns structured analysis
```

---

## 6. Phase 1 Implementation Plan

### Day 1: Project Scaffolding

- `go mod init` + Cobra command skeleton
- AI Client interface + Anthropic API implementation
- End-to-end: CLI receives args → collects data → calls AI → outputs result

### Day 2-3: ImagePullBackOff Scenario

- Pod data collection module
- Scenario-specific prompt template
- Result formatting

### Day 4-5: CrashLoopBackOff Scenario

- CrashLoopBackOff prompt template
- Log truncation strategy (last crash cycle logs)
- Multi-scenario routing

### Day 6-7: Polish & Release

- CLI output formatting
- Error handling (missing kubeconfig, network timeout, AI failure)
- README / GitHub Release

### Success Criteria

- Correct root cause analysis for target scenarios (verified on 3-5 real cases)
- Output is genuinely useful ("your container crashed" is NOT useful)
- CLI is usable within 5 minutes on a fresh environment

---

## 7. Product Roadmap

| Phase | Form Factor | Core Capability |
|-------|------------|-----------------|
| Phase 1 | CLI | 5 failure type analysis |
| Phase 2 | Web UI / SaaS | Visual analysis + history |
| Phase 3 | VSCode Extension | YAML validation / Probe suggestions / Inline AI |
| Phase 4 | AI SRE Agent | Auto-diagnosis / rollback / remediation |

---

## 8. Core Belief

> The future value is NOT "using AI" — it's **understanding Kubernetes incidents better than anyone else.**

The product direction is always **Incident Intelligence**, not a generic AI chatbot.
