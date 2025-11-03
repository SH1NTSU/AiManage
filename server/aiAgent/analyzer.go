package aiAgent

import (
	"fmt"
	"strings"
	"time"
)

// PerformanceAnalysis represents AI-generated insights about model performance
type PerformanceAnalysis struct {
	Summary           string                 `json:"summary"`
	Strengths         []string               `json:"strengths"`
	Weaknesses        []string               `json:"weaknesses"`
	Recommendations   []string               `json:"recommendations"`
	OverallAssessment string                 `json:"overall_assessment"`
	RawAnalysis       string                 `json:"raw_analysis"`
	Metrics           map[string]interface{} `json:"metrics"`
}

// AnalyzeTrainingResults analyzes training results using Gemini AI
func (a *Agent) AnalyzeTrainingResults(progress *TrainingProgress) (*PerformanceAnalysis, error) {
	if a.apiKey == "" {
		return nil, fmt.Errorf("Gemini AI analysis requires GEMINI_API_KEY")
	}

	// Prepare the analysis prompt
	prompt := a.buildAnalysisPrompt(progress)

	// Send to Gemini
	response, err := a.client.SendPrompt(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze with Gemini: %w", err)
	}

	// Parse the response (for now, just return the raw analysis)
	// You could add more sophisticated parsing here
	analysis := &PerformanceAnalysis{
		RawAnalysis: response,
		Metrics:     a.extractMetricsSummary(progress),
	}

	// Try to extract structured information from the response
	a.parseAnalysisResponse(response, analysis)

	return analysis, nil
}

// buildAnalysisPrompt creates a comprehensive prompt for Claude
func (a *Agent) buildAnalysisPrompt(progress *TrainingProgress) string {
	var sb strings.Builder

	sb.WriteString("# Model Training Performance Analysis\n\n")
	sb.WriteString("Please analyze the following machine learning model training results and provide insights.\n\n")

	// Training overview
	sb.WriteString("## Training Overview\n")
	sb.WriteString(fmt.Sprintf("- Status: %s\n", progress.Status))
	sb.WriteString(fmt.Sprintf("- Total Epochs: %d\n", progress.TotalEpochs))
	if progress.EndTime != nil {
		duration := progress.EndTime.Sub(progress.StartTime)
		sb.WriteString(fmt.Sprintf("- Duration: %s\n", duration.Round(time.Second)))
	}
	sb.WriteString(fmt.Sprintf("- Model Path: %s\n", progress.ModelPath))
	sb.WriteString("\n")

	// Metrics progression
	if len(progress.Metrics) > 0 {
		sb.WriteString("## Training Metrics Progression\n\n")

		// Show first, middle, and last epochs
		milestones := []int{0}
		if len(progress.Metrics) > 2 {
			milestones = append(milestones, len(progress.Metrics)/2)
		}
		if len(progress.Metrics) > 1 {
			milestones = append(milestones, len(progress.Metrics)-1)
		}

		for _, idx := range milestones {
			m := progress.Metrics[idx]
			sb.WriteString(fmt.Sprintf("### Epoch %d/%d\n", m.Epoch, m.TotalEpochs))
			if m.TrainLoss > 0 {
				sb.WriteString(fmt.Sprintf("- Training Loss: %.4f\n", m.TrainLoss))
			}
			if m.ValLoss > 0 {
				sb.WriteString(fmt.Sprintf("- Validation Loss: %.4f\n", m.ValLoss))
			}
			if m.TrainAccuracy > 0 {
				sb.WriteString(fmt.Sprintf("- Training Accuracy: %.2f%%\n", m.TrainAccuracy*100))
			}
			if m.ValAccuracy > 0 {
				sb.WriteString(fmt.Sprintf("- Validation Accuracy: %.2f%%\n", m.ValAccuracy*100))
			}
			sb.WriteString("\n")
		}
	}

	// Final metrics
	if progress.FinalMetrics != nil {
		sb.WriteString("## Final Performance\n")
		m := progress.FinalMetrics
		if m.TestAccuracy > 0 {
			sb.WriteString(fmt.Sprintf("- Test Accuracy: %.2f%%\n", m.TestAccuracy*100))
		}
		if m.TrainLoss > 0 {
			sb.WriteString(fmt.Sprintf("- Final Training Loss: %.4f\n", m.TrainLoss))
		}
		if m.ValLoss > 0 {
			sb.WriteString(fmt.Sprintf("- Final Validation Loss: %.4f\n", m.ValLoss))
		}
		sb.WriteString("\n")
	}

	// Recent logs (last 20 lines)
	if len(progress.Logs) > 0 {
		sb.WriteString("## Recent Training Logs\n```\n")
		startIdx := 0
		if len(progress.Logs) > 20 {
			startIdx = len(progress.Logs) - 20
		}
		for i := startIdx; i < len(progress.Logs); i++ {
			sb.WriteString(progress.Logs[i])
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}

	// Error information
	if progress.ErrorMessage != "" {
		sb.WriteString("## Errors\n")
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", progress.ErrorMessage))
	}

	// Analysis request
	sb.WriteString("## Analysis Request\n\n")
	sb.WriteString("Please provide:\n")
	sb.WriteString("1. **Summary**: Brief overview of the training performance\n")
	sb.WriteString("2. **Strengths**: What went well in this training run\n")
	sb.WriteString("3. **Weaknesses**: Areas of concern or poor performance\n")
	sb.WriteString("4. **Recommendations**: Specific suggestions for improvement (hyperparameters, architecture, data, etc.)\n")
	sb.WriteString("5. **Overall Assessment**: Is this model ready for production? Should we retrain?\n")

	return sb.String()
}

// extractMetricsSummary creates a summary of key metrics
func (a *Agent) extractMetricsSummary(progress *TrainingProgress) map[string]interface{} {
	summary := make(map[string]interface{})

	summary["status"] = progress.Status
	summary["total_epochs"] = progress.TotalEpochs
	summary["completed_epochs"] = progress.CurrentEpoch

	if progress.EndTime != nil {
		duration := progress.EndTime.Sub(progress.StartTime)
		summary["training_duration_seconds"] = duration.Seconds()
	}

	// Get final metrics
	if len(progress.Metrics) > 0 {
		lastMetric := progress.Metrics[len(progress.Metrics)-1]
		summary["final_train_loss"] = lastMetric.TrainLoss
		summary["final_val_loss"] = lastMetric.ValLoss
		summary["final_train_accuracy"] = lastMetric.TrainAccuracy
		summary["final_val_accuracy"] = lastMetric.ValAccuracy
	}

	if progress.FinalMetrics != nil {
		summary["test_accuracy"] = progress.FinalMetrics.TestAccuracy
	}

	// Calculate improvement
	if len(progress.Metrics) >= 2 {
		first := progress.Metrics[0]
		last := progress.Metrics[len(progress.Metrics)-1]

		if first.TrainLoss > 0 && last.TrainLoss > 0 {
			improvement := ((first.TrainLoss - last.TrainLoss) / first.TrainLoss) * 100
			summary["loss_improvement_percent"] = improvement
		}

		if first.TrainAccuracy > 0 && last.TrainAccuracy > 0 {
			improvement := ((last.TrainAccuracy - first.TrainAccuracy) / first.TrainAccuracy) * 100
			summary["accuracy_improvement_percent"] = improvement
		}
	}

	return summary
}

// parseAnalysisResponse attempts to extract structured information from Claude's response
func (a *Agent) parseAnalysisResponse(response string, analysis *PerformanceAnalysis) {
	lines := strings.Split(response, "\n")

	var currentSection string
	var summaryLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect sections
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "summary") || strings.Contains(lowerLine, "overview") {
			currentSection = "summary"
			continue
		} else if strings.Contains(lowerLine, "strength") {
			currentSection = "strengths"
			continue
		} else if strings.Contains(lowerLine, "weakness") || strings.Contains(lowerLine, "concern") {
			currentSection = "weaknesses"
			continue
		} else if strings.Contains(lowerLine, "recommendation") || strings.Contains(lowerLine, "suggestion") {
			currentSection = "recommendations"
			continue
		} else if strings.Contains(lowerLine, "overall") || strings.Contains(lowerLine, "assessment") {
			currentSection = "overall"
			continue
		}

		// Parse content based on section
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "•") {
			content := strings.TrimLeft(line, "-*• ")
			switch currentSection {
			case "strengths":
				analysis.Strengths = append(analysis.Strengths, content)
			case "weaknesses":
				analysis.Weaknesses = append(analysis.Weaknesses, content)
			case "recommendations":
				analysis.Recommendations = append(analysis.Recommendations, content)
			}
		} else {
			if currentSection == "summary" {
				summaryLines = append(summaryLines, line)
			} else if currentSection == "overall" {
				if analysis.OverallAssessment == "" {
					analysis.OverallAssessment = line
				} else {
					analysis.OverallAssessment += " " + line
				}
			}
		}
	}

	if len(summaryLines) > 0 {
		analysis.Summary = strings.Join(summaryLines, " ")
	}

	// If we didn't parse anything structured, use the whole response
	if analysis.Summary == "" {
		analysis.Summary = response
	}
}

// QuickAnalysis provides a quick analysis without Claude AI
func (a *Agent) QuickAnalysis(progress *TrainingProgress) *PerformanceAnalysis {
	analysis := &PerformanceAnalysis{
		Metrics:     a.extractMetricsSummary(progress),
		Strengths:   []string{},
		Weaknesses:  []string{},
		Recommendations: []string{},
	}

	// Basic heuristics
	if len(progress.Metrics) >= 2 {
		first := progress.Metrics[0]
		last := progress.Metrics[len(progress.Metrics)-1]

		// Check loss improvement
		if last.TrainLoss < first.TrainLoss*0.5 {
			analysis.Strengths = append(analysis.Strengths, "Training loss decreased significantly")
		} else if last.TrainLoss > first.TrainLoss*0.9 {
			analysis.Weaknesses = append(analysis.Weaknesses, "Training loss did not decrease much")
			analysis.Recommendations = append(analysis.Recommendations, "Consider adjusting learning rate")
		}

		// Check overfitting
		if last.ValLoss > 0 && last.TrainLoss > 0 {
			gap := (last.ValLoss - last.TrainLoss) / last.TrainLoss
			if gap > 0.3 {
				analysis.Weaknesses = append(analysis.Weaknesses, "Significant gap between training and validation loss (possible overfitting)")
				analysis.Recommendations = append(analysis.Recommendations, "Add regularization or dropout")
				analysis.Recommendations = append(analysis.Recommendations, "Increase training data or use data augmentation")
			}
		}

		// Check accuracy
		if last.ValAccuracy > 0.9 {
			analysis.Strengths = append(analysis.Strengths, fmt.Sprintf("High validation accuracy: %.2f%%", last.ValAccuracy*100))
		} else if last.ValAccuracy < 0.6 {
			analysis.Weaknesses = append(analysis.Weaknesses, "Low validation accuracy")
			analysis.Recommendations = append(analysis.Recommendations, "Review model architecture and data quality")
		}
	}

	// Overall assessment
	if progress.Status == StatusCompleted {
		if len(analysis.Weaknesses) == 0 {
			analysis.OverallAssessment = "Training completed successfully with good performance metrics"
		} else if len(analysis.Weaknesses) > len(analysis.Strengths) {
			analysis.OverallAssessment = "Training completed but with performance concerns - recommend retraining with adjustments"
		} else {
			analysis.OverallAssessment = "Training completed with acceptable performance - minor improvements possible"
		}
	} else if progress.Status == StatusFailed {
		analysis.OverallAssessment = "Training failed - review error logs"
	}

	return analysis
}
