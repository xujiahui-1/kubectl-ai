package analyzer

import (
	"strings"

	"github.com/xujiahui-1/kubectl-ai/internal/prompt"
	corev1 "k8s.io/api/core/v1"
)

// detectScenario identifies the failure scenario from a Pod's container statuses and events.
func detectScenario(pod *corev1.Pod, events string) prompt.Scenario {
	// 1. Init Container Error
	for _, ic := range pod.Status.InitContainerStatuses {
		if ic.State.Terminated != nil && ic.State.Terminated.ExitCode != 0 {
			return prompt.ScenarioInitContainerError
		}
		if ic.State.Waiting != nil && ic.State.Waiting.Reason == "CrashLoopBackOff" {
			return prompt.ScenarioInitContainerError
		}
	}

	// 2. Image pull issues
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			switch cs.State.Waiting.Reason {
			case "ImagePullBackOff", "ErrImagePull":
				return prompt.ScenarioImagePullBackOff
			case "CrashLoopBackOff":
				return prompt.ScenarioCrashLoopBackOff
			}
		}

		// 3. OOMKilled
		if cs.State.Terminated != nil && cs.State.Terminated.Reason == "OOMKilled" {
			return prompt.ScenarioOOMKilled
		}
		if cs.LastTerminationState.Terminated != nil && cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
			return prompt.ScenarioOOMKilled
		}

		// 4. Implicit crash loop (Terminated:Error with multiple restarts)
		if cs.State.Terminated != nil && cs.State.Terminated.Reason == "Error" {
			if cs.RestartCount > 2 {
				return prompt.ScenarioCrashLoopBackOff
			}
		}
	}

	// 5. Pending pods with scheduling issues
	if pod.Status.Phase == corev1.PodPending {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
				if strings.Contains(cond.Reason, "Unschedulable") {
					return prompt.ScenarioPendingPod
				}
			}
		}
	}

	// 6. Probe failure (detected via events)
	if strings.Contains(events, "Unhealthy") {
		return prompt.ScenarioProbeFailure
	}

	return prompt.ScenarioGeneric
}
