package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

const defaultLogLines = 100

type Client struct {
	cs        kubernetes.Interface
	namespace string
}

func NewClient(cs kubernetes.Interface, namespace string) *Client {
	return &Client{cs: cs, namespace: namespace}
}

func (c *Client) GetPod(ctx context.Context, name string) (*corev1.Pod, error) {
	return c.cs.CoreV1().Pods(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) GetPodLogs(ctx context.Context, podName string) (string, error) {
	req := c.cs.CoreV1().Pods(c.namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: int64Ptr(defaultLogLines),
	})
	data, err := req.Do(ctx).Raw()
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod %s: %w", podName, err)
	}
	return string(data), nil
}

func (c *Client) GetInitContainerLogs(ctx context.Context, podName, containerName string) (string, error) {
	req := c.cs.CoreV1().Pods(c.namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: int64Ptr(defaultLogLines),
		Container: containerName,
	})
	data, err := req.Do(ctx).Raw()
	if err != nil {
		return "", fmt.Errorf("failed to get init container logs for %s/%s: %w", podName, containerName, err)
	}
	return string(data), nil
}

func (c *Client) GetPodEvents(ctx context.Context, podName string) (string, error) {
	events, err := c.cs.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fields.Set{
			"involvedObject.name": podName,
			"involvedObject.kind": "Pod",
		}.AsSelector().String(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get events for pod %s: %w", podName, err)
	}

	if len(events.Items) == 0 {
		return "No events found.", nil
	}

	var b strings.Builder
	for _, e := range events.Items {
		b.WriteString(fmt.Sprintf("[%s] %s: %s (x%d)\n",
			e.LastTimestamp.Format("2006-01-02 15:04:05"),
			e.Reason,
			e.Message,
			e.Count,
		))
	}
	return b.String(), nil
}

func (c *Client) GetPodYAML(ctx context.Context, podName string) (string, error) {
	pod, err := c.GetPod(ctx, podName)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Name: %s\n", pod.Name))
	b.WriteString(fmt.Sprintf("Namespace: %s\n", pod.Namespace))
	b.WriteString(fmt.Sprintf("Node: %s\n", pod.Spec.NodeName))
	b.WriteString(fmt.Sprintf("Status: %s\n", pod.Status.Phase))
	b.WriteString(fmt.Sprintf("HostIP: %s\n", pod.Status.HostIP))
	b.WriteString(fmt.Sprintf("PodIP: %s\n", pod.Status.PodIP))
	b.WriteString(fmt.Sprintf("QoS Class: %s\n", pod.Status.QOSClass))

	// Init containers
	for _, ic := range pod.Status.InitContainerStatuses {
		b.WriteString(fmt.Sprintf("\nInit Container: %s\n", ic.Name))
		b.WriteString(fmt.Sprintf("  Image: %s\n", ic.Image))
		b.WriteString(fmt.Sprintf("  RestartCount: %d\n", ic.RestartCount))
		if ic.State.Waiting != nil {
			b.WriteString(fmt.Sprintf("  State: Waiting (%s) — %s\n", ic.State.Waiting.Reason, ic.State.Waiting.Message))
		}
		if ic.State.Terminated != nil {
			b.WriteString(fmt.Sprintf("  State: Terminated (%s — exit %d)\n", ic.State.Terminated.Reason, ic.State.Terminated.ExitCode))
		}
	}

	// Main containers
	for _, c := range pod.Status.ContainerStatuses {
		b.WriteString(fmt.Sprintf("\nContainer: %s\n", c.Name))
		b.WriteString(fmt.Sprintf("  Image: %s\n", c.Image))
		b.WriteString(fmt.Sprintf("  Ready: %v\n", c.Ready))
		b.WriteString(fmt.Sprintf("  RestartCount: %d\n", c.RestartCount))
		if c.State.Waiting != nil {
			b.WriteString(fmt.Sprintf("  State: Waiting (%s) — %s\n", c.State.Waiting.Reason, c.State.Waiting.Message))
		}
		if c.State.Running != nil {
			b.WriteString("  State: Running\n")
		}
		if c.State.Terminated != nil {
			b.WriteString(fmt.Sprintf("  State: Terminated (%s) — %s\n", c.State.Terminated.Reason, c.State.Terminated.Message))
		}
		if c.LastTerminationState.Terminated != nil {
			lt := c.LastTerminationState.Terminated
			b.WriteString(fmt.Sprintf("  LastTermination: %s — %s (exit %d)\n", lt.Reason, lt.Message, lt.ExitCode))
		}
		// Include probe info if configured
		if c.Ready && c.State.Running != nil {
			// Check if probes are configured by looking at the spec
		}
	}

	for _, c := range pod.Status.Conditions {
		b.WriteString(fmt.Sprintf("\nCondition: %s = %s (%s)\n", c.Type, c.Status, c.Message))
	}

	return b.String(), nil
}

func (c *Client) GetParentResource(ctx context.Context, pod *corev1.Pod) (string, error) {
	for _, ref := range pod.OwnerReferences {
		if ref.Kind == "ReplicaSet" {
			rs, err := c.cs.AppsV1().ReplicaSets(c.namespace).Get(ctx, ref.Name, metav1.GetOptions{})
			if err != nil {
				return "", fmt.Errorf("failed to get ReplicaSet %s: %w", ref.Name, err)
			}
			for _, owner := range rs.OwnerReferences {
				if owner.Kind == "Deployment" {
					deploy, err := c.cs.AppsV1().Deployments(c.namespace).Get(ctx, owner.Name, metav1.GetOptions{})
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("Deployment: %s\nReplicas: %d\nStrategy: %s\n",
						deploy.Name, deploy.Status.Replicas, deploy.Spec.Strategy.Type), nil
				}
			}
		}
	}
	return "No parent Deployment found (standalone Pod).", nil
}

func int64Ptr(i int64) *int64 { return &i }
