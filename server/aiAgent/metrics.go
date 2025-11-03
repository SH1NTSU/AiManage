package aiAgent

import (
	"fmt"
	"math"
	"time"
)

// DetailedMetrics provides comprehensive analysis without AI
type DetailedMetrics struct {
	// Overview
	TrainingStatus    string    `json:"training_status"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	TotalDuration     float64   `json:"total_duration_seconds"`
	CompletedEpochs   int       `json:"completed_epochs"`
	TotalEpochs       int       `json:"total_epochs"`
	AverageEpochTime  float64   `json:"average_epoch_time_seconds"`

	// Loss Metrics
	InitialLoss       float64   `json:"initial_loss"`
	FinalLoss         float64   `json:"final_loss"`
	BestLoss          float64   `json:"best_loss"`
	WorstLoss         float64   `json:"worst_loss"`
	AverageLoss       float64   `json:"average_loss"`
	LossImprovement   float64   `json:"loss_improvement_percent"`
	LossStdDev        float64   `json:"loss_std_dev"`

	// Validation Loss Metrics
	InitialValLoss    float64   `json:"initial_val_loss"`
	FinalValLoss      float64   `json:"final_val_loss"`
	BestValLoss       float64   `json:"best_val_loss"`
	AverageValLoss    float64   `json:"average_val_loss"`
	ValLossImprovement float64  `json:"val_loss_improvement_percent"`

	// Accuracy Metrics
	InitialAccuracy   float64   `json:"initial_accuracy"`
	FinalAccuracy     float64   `json:"final_accuracy"`
	BestAccuracy      float64   `json:"best_accuracy"`
	AverageAccuracy   float64   `json:"average_accuracy"`
	AccuracyImprovement float64 `json:"accuracy_improvement_percent"`

	// Validation Accuracy Metrics
	InitialValAccuracy float64  `json:"initial_val_accuracy"`
	FinalValAccuracy   float64  `json:"final_val_accuracy"`
	BestValAccuracy    float64  `json:"best_val_accuracy"`
	AverageValAccuracy float64  `json:"average_val_accuracy"`
	ValAccuracyImprovement float64 `json:"val_accuracy_improvement_percent"`

	// Test Metrics (if available)
	TestAccuracy      float64   `json:"test_accuracy,omitempty"`
	TestLoss          float64   `json:"test_loss,omitempty"`

	// Training Behavior Analysis
	IsConverging      bool      `json:"is_converging"`
	IsOverfitting     bool      `json:"is_overfitting"`
	IsUnderfitting    bool      `json:"is_underfitting"`
	TrainValGap       float64   `json:"train_val_gap"`
	LossVariability   string    `json:"loss_variability"` // "stable", "moderate", "unstable"

	// Performance Assessment
	OverallScore      float64   `json:"overall_score"` // 0-100
	PerformanceLevel  string    `json:"performance_level"` // "excellent", "good", "fair", "poor"

	// Chart Data (ready for React charts)
	EpochData         []EpochMetric `json:"epoch_data"`
	LossHistory       []float64     `json:"loss_history"`
	ValLossHistory    []float64     `json:"val_loss_history"`
	AccuracyHistory   []float64     `json:"accuracy_history"`
	ValAccuracyHistory []float64    `json:"val_accuracy_history"`

	// Insights & Recommendations
	Insights          []string      `json:"insights"`
	Warnings          []string      `json:"warnings"`
	Recommendations   []string      `json:"recommendations"`

	// Model Files
	ModelPath         string        `json:"model_path,omitempty"`
	HasCheckpoint     bool          `json:"has_checkpoint"`
}

// EpochMetric represents metrics for a single epoch (chart-ready)
type EpochMetric struct {
	Epoch         int     `json:"epoch"`
	TrainLoss     float64 `json:"train_loss"`
	ValLoss       float64 `json:"val_loss"`
	TrainAccuracy float64 `json:"train_accuracy"`
	ValAccuracy   float64 `json:"val_accuracy"`
	Duration      float64 `json:"duration_seconds"`
}

// GenerateDetailedMetrics creates comprehensive metrics from training progress
func GenerateDetailedMetrics(progress *TrainingProgress) *DetailedMetrics {
	metrics := &DetailedMetrics{
		TrainingStatus:  string(progress.Status),
		StartTime:       progress.StartTime,
		CompletedEpochs: progress.CurrentEpoch,
		TotalEpochs:     progress.TotalEpochs,
		EpochData:       []EpochMetric{},
		LossHistory:     []float64{},
		ValLossHistory:  []float64{},
		AccuracyHistory: []float64{},
		ValAccuracyHistory: []float64{},
		Insights:        []string{},
		Warnings:        []string{},
		Recommendations: []string{},
		ModelPath:       progress.ModelPath,
	}

	if progress.EndTime != nil {
		metrics.EndTime = *progress.EndTime
		metrics.TotalDuration = progress.EndTime.Sub(progress.StartTime).Seconds()
	}

	// No metrics to analyze
	if len(progress.Metrics) == 0 {
		return metrics
	}

	// Calculate average epoch time
	if metrics.TotalDuration > 0 && progress.CurrentEpoch > 0 {
		metrics.AverageEpochTime = metrics.TotalDuration / float64(progress.CurrentEpoch)
	}

	// Extract metrics for each epoch
	var trainLosses, valLosses, trainAccs, valAccs []float64

	for _, m := range progress.Metrics {
		epochMetric := EpochMetric{
			Epoch:         m.Epoch,
			TrainLoss:     m.TrainLoss,
			ValLoss:       m.ValLoss,
			TrainAccuracy: m.TrainAccuracy * 100, // Convert to percentage
			ValAccuracy:   m.ValAccuracy * 100,
		}
		metrics.EpochData = append(metrics.EpochData, epochMetric)

		if m.TrainLoss > 0 {
			trainLosses = append(trainLosses, m.TrainLoss)
			metrics.LossHistory = append(metrics.LossHistory, m.TrainLoss)
		}
		if m.ValLoss > 0 {
			valLosses = append(valLosses, m.ValLoss)
			metrics.ValLossHistory = append(metrics.ValLossHistory, m.ValLoss)
		}
		if m.TrainAccuracy > 0 {
			trainAccs = append(trainAccs, m.TrainAccuracy * 100)
			metrics.AccuracyHistory = append(metrics.AccuracyHistory, m.TrainAccuracy * 100)
		}
		if m.ValAccuracy > 0 {
			valAccs = append(valAccs, m.ValAccuracy * 100)
			metrics.ValAccuracyHistory = append(metrics.ValAccuracyHistory, m.ValAccuracy * 100)
		}
	}

	// Calculate loss statistics
	if len(trainLosses) > 0 {
		metrics.InitialLoss = trainLosses[0]
		metrics.FinalLoss = trainLosses[len(trainLosses)-1]
		metrics.BestLoss = min(trainLosses)
		metrics.WorstLoss = max(trainLosses)
		metrics.AverageLoss = average(trainLosses)
		metrics.LossStdDev = stdDev(trainLosses)

		if metrics.InitialLoss > 0 {
			metrics.LossImprovement = ((metrics.InitialLoss - metrics.FinalLoss) / metrics.InitialLoss) * 100
		}
	}

	// Calculate validation loss statistics
	if len(valLosses) > 0 {
		metrics.InitialValLoss = valLosses[0]
		metrics.FinalValLoss = valLosses[len(valLosses)-1]
		metrics.BestValLoss = min(valLosses)
		metrics.AverageValLoss = average(valLosses)

		if metrics.InitialValLoss > 0 {
			metrics.ValLossImprovement = ((metrics.InitialValLoss - metrics.FinalValLoss) / metrics.InitialValLoss) * 100
		}
	}

	// Calculate accuracy statistics
	if len(trainAccs) > 0 {
		metrics.InitialAccuracy = trainAccs[0]
		metrics.FinalAccuracy = trainAccs[len(trainAccs)-1]
		metrics.BestAccuracy = max(trainAccs)
		metrics.AverageAccuracy = average(trainAccs)

		if metrics.InitialAccuracy > 0 {
			metrics.AccuracyImprovement = ((metrics.FinalAccuracy - metrics.InitialAccuracy) / metrics.InitialAccuracy) * 100
		}
	}

	// Calculate validation accuracy statistics
	if len(valAccs) > 0 {
		metrics.InitialValAccuracy = valAccs[0]
		metrics.FinalValAccuracy = valAccs[len(valAccs)-1]
		metrics.BestValAccuracy = max(valAccs)
		metrics.AverageValAccuracy = average(valAccs)

		if metrics.InitialValAccuracy > 0 {
			metrics.ValAccuracyImprovement = ((metrics.FinalValAccuracy - metrics.InitialValAccuracy) / metrics.InitialValAccuracy) * 100
		}
	}

	// Add test metrics if available
	if progress.FinalMetrics != nil {
		if progress.FinalMetrics.TestAccuracy > 0 {
			metrics.TestAccuracy = progress.FinalMetrics.TestAccuracy * 100
		}
	}

	// Analyze training behavior
	analyzeTrainingBehavior(metrics, trainLosses, valLosses, trainAccs, valAccs)

	// Generate insights
	generateInsights(metrics, progress)

	// Calculate overall score
	metrics.OverallScore = calculateOverallScore(metrics)
	metrics.PerformanceLevel = getPerformanceLevel(metrics.OverallScore)

	return metrics
}

// analyzeTrainingBehavior determines if model is converging, overfitting, etc.
func analyzeTrainingBehavior(metrics *DetailedMetrics, trainLosses, valLosses, trainAccs, valAccs []float64) {
	// Check convergence
	if len(trainLosses) >= 3 {
		recentLosses := trainLosses[len(trainLosses)-3:]
		if isDecreasing(recentLosses) || isStable(recentLosses) {
			metrics.IsConverging = true
		}
	}

	// Check overfitting (train/val gap)
	if metrics.FinalLoss > 0 && metrics.FinalValLoss > 0 {
		gap := (metrics.FinalValLoss - metrics.FinalLoss) / metrics.FinalLoss
		metrics.TrainValGap = gap * 100

		if gap > 0.3 {
			metrics.IsOverfitting = true
			metrics.Warnings = append(metrics.Warnings, "Model shows signs of overfitting (large train/val gap)")
		}
	}

	// Check underfitting
	if metrics.FinalAccuracy < 60 && metrics.FinalValAccuracy < 60 {
		metrics.IsUnderfitting = true
		metrics.Warnings = append(metrics.Warnings, "Model may be underfitting (low accuracy on both train and validation)")
	}

	// Check loss variability
	if metrics.LossStdDev < metrics.AverageLoss*0.1 {
		metrics.LossVariability = "stable"
	} else if metrics.LossStdDev < metrics.AverageLoss*0.3 {
		metrics.LossVariability = "moderate"
	} else {
		metrics.LossVariability = "unstable"
		metrics.Warnings = append(metrics.Warnings, "Training loss is unstable - consider reducing learning rate")
	}
}

// generateInsights creates actionable insights and recommendations
func generateInsights(metrics *DetailedMetrics, progress *TrainingProgress) {
	// Loss improvement insights
	if metrics.LossImprovement > 50 {
		metrics.Insights = append(metrics.Insights, fmt.Sprintf("Excellent loss reduction: %.1f%% improvement", metrics.LossImprovement))
	} else if metrics.LossImprovement > 20 {
		metrics.Insights = append(metrics.Insights, fmt.Sprintf("Good loss reduction: %.1f%% improvement", metrics.LossImprovement))
	} else if metrics.LossImprovement < 10 {
		metrics.Warnings = append(metrics.Warnings, "Poor loss reduction - model may need more training or learning rate adjustment")
		metrics.Recommendations = append(metrics.Recommendations, "Increase number of epochs or adjust learning rate")
	}

	// Accuracy insights
	if metrics.FinalValAccuracy > 90 {
		metrics.Insights = append(metrics.Insights, fmt.Sprintf("Excellent validation accuracy: %.2f%%", metrics.FinalValAccuracy))
	} else if metrics.FinalValAccuracy > 80 {
		metrics.Insights = append(metrics.Insights, fmt.Sprintf("Good validation accuracy: %.2f%%", metrics.FinalValAccuracy))
	} else if metrics.FinalValAccuracy > 70 {
		metrics.Insights = append(metrics.Insights, fmt.Sprintf("Moderate validation accuracy: %.2f%%", metrics.FinalValAccuracy))
		metrics.Recommendations = append(metrics.Recommendations, "Try data augmentation or model architecture improvements")
	} else {
		metrics.Warnings = append(metrics.Warnings, fmt.Sprintf("Low validation accuracy: %.2f%%", metrics.FinalValAccuracy))
		metrics.Recommendations = append(metrics.Recommendations, "Review data quality and model architecture")
	}

	// Overfitting recommendations
	if metrics.IsOverfitting {
		metrics.Recommendations = append(metrics.Recommendations, "Add dropout layers or L2 regularization")
		metrics.Recommendations = append(metrics.Recommendations, "Increase training data or use data augmentation")
		metrics.Recommendations = append(metrics.Recommendations, "Consider early stopping")
	}

	// Training time insights
	if metrics.AverageEpochTime > 0 {
		estimatedTimeForMore := metrics.AverageEpochTime * 10 // 10 more epochs
		metrics.Insights = append(metrics.Insights, fmt.Sprintf("Average epoch time: %.1f seconds", metrics.AverageEpochTime))

		if !metrics.IsConverging && progress.Status == StatusCompleted {
			metrics.Recommendations = append(metrics.Recommendations,
				fmt.Sprintf("Model may benefit from %d more epochs (~%.1f min)", 10, estimatedTimeForMore/60))
		}
	}

	// Convergence insights
	if metrics.IsConverging {
		metrics.Insights = append(metrics.Insights, "Model is converging well")
	} else if progress.Status == StatusCompleted {
		metrics.Warnings = append(metrics.Warnings, "Model has not fully converged")
		metrics.Recommendations = append(metrics.Recommendations, "Continue training for more epochs")
	}
}

// calculateOverallScore generates a 0-100 score
func calculateOverallScore(metrics *DetailedMetrics) float64 {
	score := 0.0

	// Accuracy contribution (40 points)
	if metrics.FinalValAccuracy > 0 {
		score += metrics.FinalValAccuracy * 0.4
	}

	// Loss improvement contribution (20 points)
	if metrics.LossImprovement > 0 {
		score += math.Min(metrics.LossImprovement, 50) * 0.4
	}

	// Convergence contribution (20 points)
	if metrics.IsConverging {
		score += 20
	}

	// Penalty for overfitting (up to -10 points)
	if metrics.IsOverfitting {
		score -= math.Min(metrics.TrainValGap*0.3, 10)
	}

	// Penalty for underfitting (up to -10 points)
	if metrics.IsUnderfitting {
		score -= 10
	}

	// Stability bonus (10 points)
	if metrics.LossVariability == "stable" {
		score += 10
	} else if metrics.LossVariability == "moderate" {
		score += 5
	}

	return math.Max(0, math.Min(100, score))
}

// getPerformanceLevel converts score to human-readable level
func getPerformanceLevel(score float64) string {
	if score >= 85 {
		return "excellent"
	} else if score >= 70 {
		return "good"
	} else if score >= 50 {
		return "fair"
	}
	return "poor"
}

// Helper functions
func min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values {
		if v < m {
			m = v
		}
	}
	return m
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values {
		if v > m {
			m = v
		}
	}
	return m
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	avg := average(values)
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-avg, 2)
	}
	return math.Sqrt(variance / float64(len(values)))
}

func isDecreasing(values []float64) bool {
	for i := 1; i < len(values); i++ {
		if values[i] >= values[i-1] {
			return false
		}
	}
	return true
}

func isStable(values []float64) bool {
	if len(values) < 2 {
		return true
	}
	sd := stdDev(values)
	avg := average(values)
	return sd < avg*0.1
}
