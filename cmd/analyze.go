package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xujiahui-1/kubectl-ai/internal/ai"
	"github.com/xujiahui-1/kubectl-ai/internal/analyzer"
	"github.com/xujiahui-1/kubectl-ai/internal/config"
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

		provider, model, apiKey := resolveAIConfig()
		aiClient, err := ai.NewClient(provider, model, apiKey)
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

	analyzeCmd.PersistentFlags().StringVar(&aiProvider, "ai-provider", "", "AI provider (deepseek, anthropic, openai)")
	analyzeCmd.PersistentFlags().StringVar(&aiModel, "ai-model", "", "AI model name")
	analyzeCmd.PersistentFlags().StringVar(&aiAPIKey, "ai-api-key", "", "AI API key")
}

func resolveAIConfig() (provider, model, apiKey string) {
	provider = aiProvider
	model = aiModel
	apiKey = aiAPIKey

	cfg, err := config.Load()
	if err == nil && cfg != nil {
		if provider == "" {
			provider = cfg.AIProvider
		}
		if model == "" {
			model = cfg.AIModel
		}
		if apiKey == "" {
			apiKey = cfg.AIAPIKey
		}
	}

	if provider == "" {
		provider = "deepseek"
	}
	if model == "" {
		switch provider {
		case "anthropic":
			model = "claude-sonnet-4-20250514"
		case "openai":
			model = "gpt-4o"
		default:
			model = "deepseek-chat"
		}
	}
	return
}
