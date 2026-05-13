package prompt

import "fmt"

// Scenario represents the type of incident being analyzed.
type Scenario string

const (
	ScenarioImagePullBackOff  Scenario = "ImagePullBackOff"
	ScenarioCrashLoopBackOff  Scenario = "CrashLoopBackOff"
	ScenarioOOMKilled         Scenario = "OOMKilled"
	ScenarioPendingPod        Scenario = "PendingPod"
	ScenarioInitContainerError Scenario = "InitContainerError"
	ScenarioProbeFailure      Scenario = "ProbeFailure"
	ScenarioGeneric           Scenario = "Generic"
)

// IncidentData carries all collected data for prompt building.
type IncidentData struct {
	PodName          string
	Namespace        string
	PodSummary       string
	PodLogs          string
	InitContainerLogs string
	PodEvents        string
	ParentResource   string
}

func BuildPrompt(scenario Scenario, data IncidentData) string {
	switch scenario {
	case ScenarioImagePullBackOff:
		return buildImagePullBackOffPrompt(data)
	case ScenarioCrashLoopBackOff:
		return buildCrashLoopBackOffPrompt(data)
	case ScenarioOOMKilled:
		return buildOOMKilledPrompt(data)
	case ScenarioPendingPod:
		return buildPendingPodPrompt(data)
	case ScenarioInitContainerError:
		return buildInitContainerErrorPrompt(data)
	case ScenarioProbeFailure:
		return buildProbeFailurePrompt(data)
	default:
		return buildGenericPrompt(data)
	}
}

func buildImagePullBackOffPrompt(data IncidentData) string {
	return fmt.Sprintf(`You are a senior Kubernetes SRE specializing in container image issues.

A pod is in ImagePullBackOff state — the container runtime cannot pull the container image.
Analyze the following data and determine the exact root cause.

Pod: %s/%s

=== Pod Status ===
%s

=== Events ===
%s

=== Parent Resource ===
%s

=== Pod Logs ===
%s

Analyze systematically. ImagePullBackOff has a limited set of root causes:
1. IMAGE NAME — Is the image name or tag misspelled? (Common: "lates" vs "latest", wrong registry path)
2. IMAGE EXISTENCE — Does the image actually exist in the registry? Check the error in Events.
3. AUTHENTICATION — Does the registry require authentication? Is imagePullSecrets configured?
4. RATE LIMIT — Is Docker Hub rate limiting being hit? (Anonymous: 100 pulls/6h)
5. IMAGE PULL POLICY — Is imagePullPolicy preventing the pull?

Return ONLY valid JSON (no markdown, no code fences):
{"root_cause": "one line root cause", "explanation": "detailed analysis of what went wrong", "risk_level": "High|Medium|Low", "suggested_fix": "actionable fix command or config change"}`,
		data.Namespace, data.PodName, data.PodSummary, data.PodEvents, data.ParentResource, data.PodLogs)
}

func buildCrashLoopBackOffPrompt(data IncidentData) string {
	return fmt.Sprintf(`You are a senior Kubernetes SRE specializing in application runtime debugging.

A pod is in CrashLoopBackOff — the container starts but keeps crashing.
Analyze the following data and determine the root cause.

Pod: %s/%s

=== Pod Status ===
%s

=== Events ===
%s

=== Pod Logs ===
%s

=== Parent Resource ===
%s

Check each possibility:
1. APPLICATION ERROR — Is there a stack trace or panic in the logs?
2. CONFIGURATION — Missing env vars, config files, or incorrect command args?
3. STARTUP DEPENDENCY — Does the app depend on a database or service that isn't ready?
4. RESOURCE LIMITS — Is the container being OOMKilled or hitting CPU limits?
5. LIVENESS PROBE — Is the liveness probe failing immediately?

Return ONLY valid JSON (no markdown, no code fences):
{"root_cause": "one line root cause", "explanation": "detailed analysis", "risk_level": "High|Medium|Low", "suggested_fix": "actionable fix"}`,
		data.Namespace, data.PodName, data.PodSummary, data.PodEvents, data.PodLogs, data.ParentResource)
}

func buildOOMKilledPrompt(data IncidentData) string {
	return fmt.Sprintf(`You are a senior Kubernetes SRE specializing in resource management.

A container was OOMKilled — it exited because it exceeded its memory limit.
Analyze the following data and determine the root cause.

Pod: %s/%s

=== Pod Status ===
%s

=== Events ===
%s

=== Pod Logs ===
%s

=== Parent Resource ===
%s

Focus on:
1. MEMORY LIMIT — Is the memory limit too low for this workload? Check the ratio between requests and limits.
2. MEMORY LEAK — Does the application have a memory leak? Check logs for repeated allocation patterns.
3. BURST TRAFFIC — Did a traffic spike cause increased memory usage?
4. QoS CLASS — Is the pod Guaranteed, Burstable, or BestEffort? Burstable pods are more likely to be OOMKilled.
5. RESOURCE REQUESTS — Are requests and limits properly configured? Consider increasing memory limit.

Return ONLY valid JSON (no markdown, no code fences):
{"root_cause": "one line root cause", "explanation": "detailed analysis", "risk_level": "High|Medium|Low", "suggested_fix": "actionable fix"}`,
		data.Namespace, data.PodName, data.PodSummary, data.PodEvents, data.PodLogs, data.ParentResource)
}

func buildPendingPodPrompt(data IncidentData) string {
	return fmt.Sprintf(`You are a senior Kubernetes SRE specializing in cluster scheduling.

A pod is stuck in Pending state — it has not been scheduled to a node.
Analyze the following data and determine the root cause.

Pod: %s/%s

=== Pod Status ===
%s

=== Events ===
%s

=== Parent Resource ===
%s

Check each possibility:
1. INSUFFICIENT RESOURCES — Does any node have enough CPU, memory, or ephemeral storage?
2. PVC NOT BOUND — Is the pod waiting for a PersistentVolumeClaim to bind?
3. NODE SELECTOR / AFFINITY — Does the nodeSelector or affinity match any available node?
4. TAINTS / TOLERATIONS — Do node taints prevent scheduling? Does the pod have tolerations?
5. LOW NODE PORT — Exhausted node ports if using hostPort or NodePort service?

Return ONLY valid JSON (no markdown, no code fences):
{"root_cause": "one line root cause", "explanation": "detailed analysis", "risk_level": "High|Medium|Low", "suggested_fix": "actionable fix"}`,
		data.Namespace, data.PodName, data.PodSummary, data.PodEvents, data.ParentResource)
}

func buildInitContainerErrorPrompt(data IncidentData) string {
	return fmt.Sprintf(`You are a senior Kubernetes SRE specializing in container initialization.

A pod is failing because an init container exited with an error.
Init containers run before the main application containers — if one fails, the pod cannot start.
Analyze the following data and determine the root cause.

Pod: %s/%s

=== Pod Status ===
%s

=== Events ===
%s

=== Init Container Logs ===
%s

=== Parent Resource ===
%s

Check each possibility:
1. INIT COMMAND — Does the init container command or args have a typo or incorrect syntax?
2. DEPENDENCY UNAVAILABLE — Is the init container trying to reach a service or resource that doesn't exist?
3. DATA VOLUME — Is the init container trying to populate a volume that isn't mounted correctly?
4. BASE IMAGE — Is the init container's image correct? Does it have the required tools?

Return ONLY valid JSON (no markdown, no code fences):
{"root_cause": "one line root cause", "explanation": "detailed analysis", "risk_level": "High|Medium|Low", "suggested_fix": "actionable fix"}`,
		data.Namespace, data.PodName, data.PodSummary, data.PodEvents, data.InitContainerLogs, data.ParentResource)
}

func buildProbeFailurePrompt(data IncidentData) string {
	return fmt.Sprintf(`You are a senior Kubernetes SRE specializing in application health checks.

A pod is failing health check probes (liveness / readiness / startup).
The kubelet has detected that the container is not responding correctly to probes.
Analyze the following data and determine the root cause.

Pod: %s/%s

=== Pod Status ===
%s

=== Events ===
%s

=== Pod Logs ===
%s

=== Parent Resource ===
%s

Check each possibility:
1. PORT MISMATCH — Is the probe trying to connect to the correct port? Check the container's containerPort vs probe port.
2. PROBE PATH — For HTTP probes: is the path correct? (Common: /health vs /healthz, missing context path)
3. TIMEOUT — Is the probe timeout too short for the application startup time?
4. STARTUP PROBE — Is a startup probe needed? Applications that are slow to start need a startupProbe.
5. APPLICATION STATE — Is the app running but returning 5xx? Check pods logs around probe time.

Return ONLY valid JSON (no markdown, no code fences):
{"root_cause": "one line root cause", "explanation": "detailed analysis", "risk_level": "High|Medium|Low", "suggested_fix": "actionable fix"}`,
		data.Namespace, data.PodName, data.PodSummary, data.PodEvents, data.PodLogs, data.ParentResource)
}

func buildGenericPrompt(data IncidentData) string {
	return fmt.Sprintf(`You are a senior Kubernetes SRE. Analyze this Kubernetes incident.

Pod: %s/%s

=== Pod Status ===
%s

=== Pod Logs ===
%s

=== Events ===
%s

=== Parent Resource ===
%s

Return ONLY valid JSON (no markdown, no code fences):
{"root_cause": "one line root cause", "explanation": "detailed analysis", "risk_level": "High|Medium|Low", "suggested_fix": "actionable fix"}`,
		data.Namespace, data.PodName, data.PodSummary, data.PodLogs, data.PodEvents, data.ParentResource)
}
