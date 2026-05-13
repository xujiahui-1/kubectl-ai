package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xujiahui-1/kubectl-ai/internal/config"
)

var providerModels = map[string][]string{
	"deepseek":  {"deepseek-chat", "deepseek-reasoner"},
	"openai":    {"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-3.5-turbo"},
	"anthropic": {"claude-sonnet-4-20250514", "claude-3-5-haiku-latest", "claude-opus-4-20250514"},
}

var providerNames = []string{"deepseek", "openai", "anthropic"}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize kubectl-ai configuration",
	Long: `Set up your AI provider, API key, and model interactively.
This creates a configuration file at ~/.kubectl-ai/config.json
so you don't need to specify --ai-provider and --ai-model every time.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if config already exists
		if existingCfg, err := config.Load(); err == nil && existingCfg != nil {
			fmt.Print("Configuration already exists. Overwrite? (y/N): ")
			var answer string
			fmt.Scanf("%s", &answer)
			if strings.ToLower(answer) != "y" {
				fmt.Println("Init cancelled.")
				return nil
			}
		}

		reader := bufio.NewReader(os.Stdin)

		// Step 1: Choose provider
		fmt.Println()
		fmt.Println("Choose your AI provider:")
		for i, name := range providerNames {
			fmt.Printf("  %d) %s\n", i+1, name)
		}
		fmt.Print("Enter number (1-3): ")
		provChoice, _ := reader.ReadString('\n')
		provChoice = strings.TrimSpace(provChoice)
		provIdx := atoiDefault(provChoice, 1) - 1
		if provIdx < 0 || provIdx >= len(providerNames) {
			provIdx = 0
		}
		provider := providerNames[provIdx]
		fmt.Printf("  Selected: %s\n", provider)

		// Step 2: Enter API key
		fmt.Println()
		fmt.Printf("Enter your %s API key: ", provider)
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			fmt.Println("Warning: API key is empty. You can set it via environment variable later.")
		}

		// Step 3: Choose model
		models := providerModels[provider]
		fmt.Println()
		fmt.Println("Choose a model:")
		for i, m := range models {
			fmt.Printf("  %d) %s\n", i+1, m)
		}
		fmt.Printf("Enter number (1-%d): ", len(models))
		modelChoice, _ := reader.ReadString('\n')
		modelChoice = strings.TrimSpace(modelChoice)
		modelIdx := atoiDefault(modelChoice, 1) - 1
		if modelIdx < 0 || modelIdx >= len(models) {
			modelIdx = 0
		}
		model := models[modelIdx]
		fmt.Printf("  Selected: %s\n", model)

		// Step 4: Save
		cfg := &config.Config{
			AIProvider: provider,
			AIModel:    model,
			AIAPIKey:   apiKey,
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		path, _ := config.ConfigPath()
		fmt.Println()
		fmt.Printf("Configuration saved to %s\n", path)
		fmt.Println("You can now run: kubectl-ai analyze pod <name> -n <namespace>")
		return nil
	},
}

func atoiDefault(s string, def int) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func init() {
	rootCmd.AddCommand(initCmd)
}
