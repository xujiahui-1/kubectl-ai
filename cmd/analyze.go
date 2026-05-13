package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xujiahui-1/kubectl-ai/internal/ai"
	"github.com/xujiahui-1/kubectl-ai/internal/analyzer"
	"github.com/xujiahui-1/kubectl-ai/internal/k8s"
)

var (
	aiProvider string
	aiModel    string
	aiAPIKey   string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a Kubernetes incident",
}

var analyzePodCmd = &cobra.Command{
	Use:   "pod [name]",
	Short: "Analyze a specific pod's incident",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		podName := args[0]

		aiClient, err := ai.NewClient(aiProvider, aiModel, aiAPIKey)
		if err != nil {
			return fmt.Errorf("failed to create AI client: %w", err)
		}

		k := k8s.NewClient(clientset, namespace)
		incidentAnalyzer := analyzer.New(k, aiClient)

		done := ai.StartSpinner("Analyzing " + podName)
		result, err := incidentAnalyzer.AnalyzePod(cmd.Context(), podName)
		done()

		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		fmt.Print(result.ColoredFormat())
		return nil
	},
}

func init() {
	analyzeCmd.AddCommand(analyzePodCmd)
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.PersistentFlags().StringVar(&aiProvider, "ai-provider", "deepseek", "AI provider (deepseek, anthropic, openai, bedrock)")
	analyzeCmd.PersistentFlags().StringVar(&aiModel, "ai-model", "deepseek-chat", "AI model name")
	analyzeCmd.PersistentFlags().StringVar(&aiAPIKey, "ai-api-key", "", "AI API key (default $DEEPSEEK_API_KEY or $ANTHROPIC_API_KEY)")
}
