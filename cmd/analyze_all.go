package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xujiahui-1/kubectl-ai/internal/ai"
	"github.com/xujiahui-1/kubectl-ai/internal/analyzer"
	"github.com/xujiahui-1/kubectl-ai/internal/k8s"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type podStatus struct {
	Name   string
	Status string
}

var analyzeAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Analyze all failing pods in the namespace",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, model, apiKey := resolveAIConfig()
		aiClient, err := ai.NewClient(provider, model, apiKey)
		if err != nil {
			return fmt.Errorf("failed to create AI client: %w", err)
		}

		k := k8s.NewClient(clientset, namespace)

		pods, err := clientset.CoreV1().Pods(namespace).List(cmd.Context(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list pods: %w", err)
		}

		failing := listFailingPods(pods.Items)
		if len(failing) == 0 {
			fmt.Printf("All pods in namespace %q are running normally.\n", namespace)
			return nil
		}

		fmt.Printf("Found %d pod(s) with issues in namespace %q:\n", len(failing), namespace)
		for _, p := range failing {
			fmt.Printf("  - %s (%s)\n", p.Name, p.Status)
		}
		fmt.Println()

		incidentAnalyzer := analyzer.New(k, aiClient)

		for _, p := range failing {
			fmt.Printf("━━━ Analyzing %s ━━━\n", p.Name)

			done := ai.StartSpinner("Analyzing " + p.Name)
			result, err := incidentAnalyzer.AnalyzePod(cmd.Context(), p.Name)
			done()

			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n\n", err)
				continue
			}

			fmt.Print(result.ColoredFormat())
			fmt.Println()
		}

		return nil
	},
}

func listFailingPods(pods []corev1.Pod) []podStatus {
	var failing []podStatus
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodSucceeded {
			continue
		}

		status := summarizeStatus(pod)

		switch {
		case pod.Status.Phase == corev1.PodRunning:
			allReady := true
			for _, c := range pod.Status.ContainerStatuses {
				if !c.Ready {
					allReady = false
					break
				}
			}
			if !allReady {
				failing = append(failing, podStatus{Name: pod.Name, Status: status})
			}
		case pod.Status.Phase == corev1.PodPending:
			failing = append(failing, podStatus{Name: pod.Name, Status: status})
		case pod.Status.Phase == corev1.PodFailed:
			failing = append(failing, podStatus{Name: pod.Name, Status: status})
		case pod.Status.Phase == corev1.PodUnknown:
			failing = append(failing, podStatus{Name: pod.Name, Status: status})
		default:
			// Check for high restart counts even if running
			restarts := 0
			for _, c := range pod.Status.ContainerStatuses {
				restarts += int(c.RestartCount)
			}
			if restarts > 0 {
				failing = append(failing, podStatus{Name: pod.Name, Status: status})
			}
		}
	}
	return failing
}

func summarizeStatus(pod corev1.Pod) string {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			return cs.State.Waiting.Reason
		}
		if cs.State.Terminated != nil {
			return cs.State.Terminated.Reason
		}
	}

	switch pod.Status.Phase {
	case corev1.PodPending:
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
				return "Unschedulable"
			}
		}
		return "Pending"
	case corev1.PodRunning:
		return "NotReady"
	case corev1.PodFailed:
		return "Failed"
	case corev1.PodUnknown:
		return "Unknown"
	}

	return string(pod.Status.Phase)
}

func init() {
	analyzeCmd.AddCommand(analyzeAllCmd)
}
