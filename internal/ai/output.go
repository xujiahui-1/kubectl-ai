package ai

import "fmt"

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
)

func colorize(text string, enabled bool) string {
	if !enabled {
		return text
	}
	return text + colorReset
}

func riskColor(level string) string {
	switch level {
	case "High":
		return colorRed
	case "Medium":
		return colorYellow
	default:
		return colorGreen
	}
}

// ColoredFormat returns a colorized version of Format.
func (r *AnalyzeResult) ColoredFormat() string {
	riskClr := riskColor(r.RiskLevel)

	return fmt.Sprintf(
		"\n%s%s Risk Level: %s%s\n\n"+
			"%sRoot Cause:%s\n"+
			"  %s\n\n"+
			"%sExplanation:%s\n"+
			"  %s\n\n"+
			"%sSuggested Fix:%s\n"+
			"  %s\n",
		riskClr, colorize("■", true), r.RiskLevel, colorReset,
		colorBold+colorBlue, colorReset,
		r.RootCause,
		colorBold+colorBlue, colorReset,
		r.Explanation,
		colorBold+colorBlue, colorReset,
		r.SuggestedFix,
	)
}
