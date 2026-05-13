package analyzer

import (
	"context"
	"fmt"
	"strings"

	"github.com/xujiahui-1/kubectl-ai/internal/ai"
	"github.com/xujiahui-1/kubectl-ai/internal/k8s"
	"github.com/xujiahui-1/kubectl-ai/internal/prompt"
)

type Analyzer struct {
	k8sClient *k8s.Client
	aiClient  ai.Client
}

func New(k *k8s.Client, a ai.Client) *Analyzer {
	return &Analyzer{
		k8sClient: k,
		aiClient:  a,
	}
}

func (a *Analyzer) AnalyzePod(ctx context.Context, podName string) (*ai.AnalyzeResult, error) {
	pod, err := a.k8sClient.GetPod(ctx, podName)
	if err != nil {
		return nil, fmt.Errorf("pod %s not found: %w", podName, err)
	}

	podSummary, err := a.k8sClient.GetPodYAML(ctx, podName)
	if err != nil {
		podSummary = fmt.Sprintf("Unable to get pod details: %s", err)
	}

	// Collect init container logs if init containers exist
	initLogs := "<no init containers>"
	if len(pod.Spec.InitContainers) > 0 {
		var logs []string
		for _, ic := range pod.Spec.InitContainers {
			if l, err := a.k8sClient.GetInitContainerLogs(ctx, podName, ic.Name); err == nil {
				logs = append(logs, fmt.Sprintf("--- Init Container: %s ---\n%s", ic.Name, l))
			} else {
				logs = append(logs, fmt.Sprintf("--- Init Container: %s ---\n<logs unavailable: %s>", ic.Name, err))
			}
		}
		initLogs = strings.Join(logs, "\n")
	}

	// Main container logs
	podLogs := "<no logs — container never started or logs unavailable>"
	if logs, err := a.k8sClient.GetPodLogs(ctx, podName); err == nil {
		podLogs = logs
	}

	// Events
	podEvents := "<no events>"
	if events, err := a.k8sClient.GetPodEvents(ctx, podName); err == nil {
		podEvents = events
	}

	parentResource := "<no parent resource>"
	if pr, err := a.k8sClient.GetParentResource(ctx, pod); err == nil {
		parentResource = pr
	}

	// Detect scenario using pod state + events
	scenario := detectScenario(pod, podEvents)

	data := prompt.IncidentData{
		PodName:           podName,
		Namespace:         pod.Namespace,
		PodSummary:        podSummary,
		PodLogs:           podLogs,
		InitContainerLogs: initLogs,
		PodEvents:         podEvents,
		ParentResource:    parentResource,
	}

	promptStr := prompt.BuildPrompt(scenario, data)

	return a.aiClient.Analyze(ctx, promptStr)
}
