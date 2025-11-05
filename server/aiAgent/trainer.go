package aiAgent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"server/internal/repository"
)

// BroadcastCallback is a function type for broadcasting training updates
type BroadcastCallback func(trainingID string, updateType string, data interface{})

var broadcastCallback BroadcastCallback

// SetBroadcastCallback sets the callback function for broadcasting updates
func SetBroadcastCallback(callback BroadcastCallback) {
	broadcastCallback = callback
}

// TrainingStatus represents the current state of training
type TrainingStatus string

const (
	StatusPending   TrainingStatus = "pending"
	StatusRunning   TrainingStatus = "running"
	StatusCompleted TrainingStatus = "completed"
	StatusFailed    TrainingStatus = "failed"
)

// TrainingMetrics holds training performance metrics
type TrainingMetrics struct {
	Epoch          int                    `json:"epoch"`
	TotalEpochs    int                    `json:"total_epochs"`
	TrainLoss      float64                `json:"train_loss,omitempty"`
	ValLoss        float64                `json:"val_loss,omitempty"`
	TrainAccuracy  float64                `json:"train_accuracy,omitempty"`
	ValAccuracy    float64                `json:"val_accuracy,omitempty"`
	TestAccuracy   float64                `json:"test_accuracy,omitempty"`
	Duration       time.Duration          `json:"duration"`
	CustomMetrics  map[string]interface{} `json:"custom_metrics,omitempty"`
}

// TrainingProgress tracks the progress of a training session
type TrainingProgress struct {
	Status        TrainingStatus    `json:"status"`
	CurrentEpoch  int               `json:"current_epoch"`
	TotalEpochs   int               `json:"total_epochs"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       *time.Time        `json:"end_time,omitempty"`
	Logs          []string          `json:"logs"`
	Metrics       []TrainingMetrics `json:"metrics"`
	FinalMetrics  *TrainingMetrics  `json:"final_metrics,omitempty"`
	ErrorMessage  string            `json:"error_message,omitempty"`
	ModelPath     string            `json:"model_path,omitempty"`
	mu            sync.RWMutex
}

// TrainingRequest represents a request to train a model
type TrainingRequest struct {
	FolderName    string            `json:"folder_name"`
	ScriptName    string            `json:"script_name"`    // e.g., "train.py"
	PythonCommand string            `json:"python_command"` // e.g., "python3" or "python"
	Args          []string          `json:"args,omitempty"` // Additional arguments
	Env           map[string]string `json:"env,omitempty"`  // Environment variables
}

// Trainer handles model training execution
type Trainer struct {
	navigator      *DirectoryNavigator
	activeTraining map[string]*TrainingProgress
	mu             sync.RWMutex
}

// NewTrainer creates a new trainer instance
func NewTrainer(navigator *DirectoryNavigator) *Trainer {
	return &Trainer{
		navigator:      navigator,
		activeTraining: make(map[string]*TrainingProgress),
	}
}

// StartTraining starts a training job
func (t *Trainer) StartTraining(ctx context.Context, req TrainingRequest) (*TrainingProgress, error) {
	println("📂 [TRAINER] Validating folder:", req.FolderName)

	// Validate folder exists
	if !t.navigator.DirectoryExists(req.FolderName) {
		println("❌ [TRAINER] Folder does not exist:", req.FolderName)
		return nil, fmt.Errorf("folder '%s' does not exist", req.FolderName)
	}
	println("✅ [TRAINER] Folder exists")

	// Get full path to script
	scriptPath := filepath.Join(t.navigator.BaseUploadPath, req.FolderName, req.ScriptName)
	println("📄 [TRAINER] Looking for script at:", scriptPath)

	if _, err := os.Stat(scriptPath); err != nil {
		println("❌ [TRAINER] Script not found:", scriptPath)
		return nil, fmt.Errorf("training script '%s' not found: %w", req.ScriptName, err)
	}
	println("✅ [TRAINER] Script found")

	// Create progress tracker
	progress := &TrainingProgress{
		Status:      StatusPending,
		StartTime:   time.Now(),
		Logs:        []string{},
		Metrics:     []TrainingMetrics{},
		TotalEpochs: 0,
	}

	// Store in active trainings
	trainingID := fmt.Sprintf("%s_%d", req.FolderName, time.Now().Unix())
	println("🆔 [TRAINER] Training ID:", trainingID)

	t.mu.Lock()
	t.activeTraining[trainingID] = progress
	t.mu.Unlock()

	println("📊 [TRAINER] Active trainings count:", len(t.activeTraining))

	// Start training in background
	println("🚀 [TRAINER] Starting training in background goroutine")
	go t.executeTraining(ctx, trainingID, req, progress)

	return progress, nil
}

// executeTraining runs the actual training script
func (t *Trainer) executeTraining(ctx context.Context, trainingID string, req TrainingRequest, progress *TrainingProgress) {
	println("\n═══════════════════════════════════════")
	println("⚙️  [EXECUTE] Training execution started")
	println("   Training ID:", trainingID)
	println("═══════════════════════════════════════\n")

	// Capture file snapshot BEFORE training
	folderPath := filepath.Join(t.navigator.BaseUploadPath, req.FolderName)
	beforeSnapshot, err := t.captureFileSnapshot(folderPath)
	if err != nil {
		println("⚠️  [EXECUTE] Failed to capture before snapshot:", err.Error())
		beforeSnapshot = nil // Continue anyway, just won't detect models
	}

	defer func() {
		endTime := time.Now()
		progress.mu.Lock()
		progress.EndTime = &endTime
		if progress.Status == StatusCompleted {
			progress.mu.Unlock() // Unlock before file I/O
			println("✅ [EXECUTE] Training completed successfully - detecting models")

			// Capture file snapshot AFTER training and detect new models
			if beforeSnapshot != nil {
				afterSnapshot, err := t.captureFileSnapshot(folderPath)
				if err == nil {
					changedModels := t.detectNewOrModifiedModels(beforeSnapshot, afterSnapshot)
					if len(changedModels) > 0 {
						println("🔍 [EXECUTE] Found", len(changedModels), "new/modified model files")
						bestModel := t.selectBestModel(changedModels)
						if bestModel != "" {
							// Convert to relative path from base upload directory
							relPath, err := filepath.Rel(t.navigator.BaseUploadPath, bestModel)
							if err != nil {
								relPath = bestModel // Fallback to absolute path
							}
							progress.mu.Lock()
							progress.ModelPath = relPath
							progress.mu.Unlock()
							println("💾 [EXECUTE] Saved trained model path:", relPath)

							// Update database with trained model path
							dbCtx := context.Background()
							if err := repository.UpdateTrainedModelPath(dbCtx, req.FolderName, relPath); err != nil {
								println("⚠️  [EXECUTE] Failed to update database with model path:", err.Error())
							} else {
								println("✅ [EXECUTE] Database updated with trained model path")
							}
						}
					} else {
						println("ℹ️  [EXECUTE] No new model files detected")
					}
				} else {
					println("⚠️  [EXECUTE] Failed to capture after snapshot:", err.Error())
				}
			}

			// Broadcast completion with model path
			progress.mu.Lock()
			if broadcastCallback != nil {
				broadcastCallback(trainingID, "status", map[string]interface{}{
					"status":        StatusCompleted,
					"error_message": "",
					"model_path":    progress.ModelPath,
				})
			}
		}
		progress.mu.Unlock()
		println("\n═══════════════════════════════════════")
		println("🏁 [EXECUTE] Training execution finished")
		println("═══════════════════════════════════════\n")
	}()

	// Update status
	progress.mu.Lock()
	progress.Status = StatusRunning
	progress.mu.Unlock()
	println("▶️  [EXECUTE] Status changed to RUNNING")

	// Broadcast status change
	if broadcastCallback != nil {
		broadcastCallback(trainingID, "status", map[string]interface{}{
			"status":        StatusRunning,
			"error_message": "",
		})
	}

	// Prepare command
	workingDir := filepath.Join(t.navigator.BaseUploadPath, req.FolderName)
	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		t.setError(progress, trainingID, fmt.Errorf("failed to resolve working directory: %w", err))
		return
	}

	// Always use direct python execution (skip wrapper scripts to avoid package compilation)
	pythonCmd := req.PythonCommand
	if pythonCmd == "" {
		pythonCmd = "python3"
	}

	scriptPath := filepath.Join(absWorkingDir, req.ScriptName)

	println("📍 [EXECUTE] Working directory:", absWorkingDir)
	println("🐍 [EXECUTE] Python command:", pythonCmd)
	println("📜 [EXECUTE] Script path:", scriptPath)

	// Use only the script name since we're setting the working directory
	args := append([]string{req.ScriptName}, req.Args...)
	println("🔧 [EXECUTE] Full command:", pythonCmd, args)

	cmd := exec.CommandContext(ctx, pythonCmd, args...)
	cmd.Dir = absWorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	// Force Python unbuffered output for real-time logs
	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")
	// Optional hints for standardized model saving (users can use or ignore)
	cmd.Env = append(cmd.Env, fmt.Sprintf("MODEL_OUTPUT_DIR=%s", filepath.Join(absWorkingDir, "saved_models")))
	cmd.Env = append(cmd.Env, fmt.Sprintf("MODEL_NAME=%s", req.FolderName))
	for key, val := range req.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}

	// Create pipes for stdout and stderr
	println("📡 [EXECUTE] Creating output pipes...")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		println("❌ [EXECUTE] Failed to create stdout pipe:", err.Error())
		t.setError(progress, trainingID, fmt.Errorf("failed to create stdout pipe: %w", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		println("❌ [EXECUTE] Failed to create stderr pipe:", err.Error())
		t.setError(progress, trainingID, fmt.Errorf("failed to create stderr pipe: %w", err))
		return
	}

	// Start command
	println("🚀 [EXECUTE] Starting Python process...")
	if err := cmd.Start(); err != nil {
		println("❌ [EXECUTE] Failed to start process:", err.Error())
		t.setError(progress, trainingID, fmt.Errorf("failed to start training: %w", err))
		return
	}
	println("✅ [EXECUTE] Python process started successfully!")

	// Read output in goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	println("👀 [EXECUTE] Starting output readers...")
	go func() {
		defer wg.Done()
		t.readOutput(stdout, progress, trainingID, false)
	}()

	go func() {
		defer wg.Done()
		t.readOutput(stderr, progress, trainingID, true)
	}()

	wg.Wait()
	println("📖 [EXECUTE] Finished reading output")

	// Wait for command to finish
	println("⏳ [EXECUTE] Waiting for process to complete...")
	if err := cmd.Wait(); err != nil {
		println("❌ [EXECUTE] Process failed:", err.Error())
		t.setError(progress, trainingID, fmt.Errorf("training failed: %w", err))
		return
	}

	// Training completed successfully
	progress.mu.Lock()
	progress.Status = StatusCompleted
	progress.mu.Unlock()
}

// readOutput reads and processes output from the training script
func (t *Trainer) readOutput(reader io.Reader, progress *TrainingProgress, trainingID string, isError bool) {
	streamType := "stdout"
	if isError {
		streamType = "stderr"
	}
	println("📡 [OUTPUT] Starting", streamType, "reader")

	lineCount := 0
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Print the output line (prefix based on stream)
		if isError {
			println("🔴 [stderr]", line)
		} else {
			println("🟢 [stdout]", line)
		}

		// Add to logs
		progress.mu.Lock()
		progress.Logs = append(progress.Logs, line)
		progress.mu.Unlock()

		// Broadcast log line
		if broadcastCallback != nil {
			broadcastCallback(trainingID, "log", map[string]interface{}{
				"message":  line,
				"is_error": isError,
			})
		}

		// Try to parse metrics from the line
		if metrics := t.parseMetrics(line); metrics != nil {
			println("📊 [METRICS] Parsed:", fmt.Sprintf("Epoch %d/%d, Loss: %.4f, Acc: %.2f%%",
				metrics.Epoch, metrics.TotalEpochs, metrics.TrainLoss, metrics.TrainAccuracy*100))

			progress.mu.Lock()
			progress.Metrics = append(progress.Metrics, *metrics)
			progress.CurrentEpoch = metrics.Epoch
			if metrics.TotalEpochs > progress.TotalEpochs {
				progress.TotalEpochs = metrics.TotalEpochs
			}
			progress.mu.Unlock()

			// Broadcast metrics update
			if broadcastCallback != nil {
				broadcastCallback(trainingID, "metrics", metrics)
			}

			// Broadcast progress update
			if broadcastCallback != nil {
				progress.mu.RLock()
				broadcastCallback(trainingID, "progress", map[string]interface{}{
					"status":        progress.Status,
					"current_epoch": progress.CurrentEpoch,
					"total_epochs":  progress.TotalEpochs,
				})
				progress.mu.RUnlock()
			}
		}
	}

	println("📡 [OUTPUT]", streamType, "reader finished. Total lines:", lineCount)
}

// parseMetrics attempts to extract metrics from a log line
func (t *Trainer) parseMetrics(line string) *TrainingMetrics {
	metrics := &TrainingMetrics{
		CustomMetrics: make(map[string]interface{}),
	}

	// Pattern: Epoch 1/10, Train Loss: 0.5432
	epochPattern := regexp.MustCompile(`Epoch\s+(\d+)[/:](\d+)`)
	if matches := epochPattern.FindStringSubmatch(line); len(matches) == 3 {
		epoch, _ := strconv.Atoi(matches[1])
		total, _ := strconv.Atoi(matches[2])
		metrics.Epoch = epoch
		metrics.TotalEpochs = total
	}

	// Pattern: Train Loss: 0.5432 or loss: 0.5432
	lossPattern := regexp.MustCompile(`(?i)(train\s*)?loss[:\s]+([0-9.]+)`)
	if matches := lossPattern.FindStringSubmatch(line); len(matches) == 3 {
		loss, _ := strconv.ParseFloat(matches[2], 64)
		metrics.TrainLoss = loss
	}

	// Pattern: Val Loss: 0.4321 or validation loss: 0.4321
	valLossPattern := regexp.MustCompile(`(?i)(val|validation)\s*loss[:\s]+([0-9.]+)`)
	if matches := valLossPattern.FindStringSubmatch(line); len(matches) == 3 {
		valLoss, _ := strconv.ParseFloat(matches[2], 64)
		metrics.ValLoss = valLoss
	}

	// Pattern: Accuracy: 0.95 or Train Accuracy: 95%
	accPattern := regexp.MustCompile(`(?i)(train\s*)?acc(?:uracy)?[:\s]+([0-9.]+)%?`)
	if matches := accPattern.FindStringSubmatch(line); len(matches) == 3 {
		acc, _ := strconv.ParseFloat(matches[2], 64)
		// Convert to 0-1 range if it's a percentage
		if acc > 1 {
			acc = acc / 100
		}
		metrics.TrainAccuracy = acc
	}

	// Pattern: Val Accuracy: 0.93
	valAccPattern := regexp.MustCompile(`(?i)(val|validation)\s*acc(?:uracy)?[:\s]+([0-9.]+)%?`)
	if matches := valAccPattern.FindStringSubmatch(line); len(matches) == 3 {
		valAcc, _ := strconv.ParseFloat(matches[2], 64)
		if valAcc > 1 {
			valAcc = valAcc / 100
		}
		metrics.ValAccuracy = valAcc
	}

	// Only return metrics if we found something useful
	if metrics.Epoch > 0 || metrics.TrainLoss > 0 || metrics.TrainAccuracy > 0 {
		return metrics
	}

	return nil
}

// setError sets an error on the progress
func (t *Trainer) setError(progress *TrainingProgress, trainingID string, err error) {
	progress.mu.Lock()
	defer progress.mu.Unlock()
	progress.Status = StatusFailed
	progress.ErrorMessage = err.Error()
	endTime := time.Now()
	progress.EndTime = &endTime

	// Broadcast error
	if broadcastCallback != nil {
		broadcastCallback(trainingID, "status", map[string]interface{}{
			"status":        StatusFailed,
			"error_message": err.Error(),
		})
	}
}

// GetProgress returns the current progress of a training job
func (t *Trainer) GetProgress(trainingID string) (*TrainingProgress, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	progress, exists := t.activeTraining[trainingID]
	if !exists {
		return nil, fmt.Errorf("training job '%s' not found", trainingID)
	}

	return progress, nil
}

// GetAllTrainings returns all training jobs
func (t *Trainer) GetAllTrainings() map[string]*TrainingProgress {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]*TrainingProgress)
	for k, v := range t.activeTraining {
		result[k] = v
	}
	return result
}

// CleanupOldTrainings removes completed training jobs older than the specified duration
func (t *Trainer) CleanupOldTrainings(olderThan time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for id, progress := range t.activeTraining {
		if progress.EndTime != nil && now.Sub(*progress.EndTime) > olderThan {
			delete(t.activeTraining, id)
		}
	}
}

// ClearModelTrainings removes all training progress for a specific model
func (t *Trainer) ClearModelTrainings(modelName string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	count := 0
	for id := range t.activeTraining {
		// Training IDs are formatted as "{modelName}_{timestamp}"
		if strings.HasPrefix(id, modelName+"_") {
			delete(t.activeTraining, id)
			count++
		}
	}

	if count > 0 {
		log.Printf("🗑️  Cleared %d training progress entries for model '%s'", count, modelName)
	}

	return count
}

// FileSnapshot represents a snapshot of a file at a point in time
type FileSnapshot struct {
	Path    string
	ModTime time.Time
	Size    int64
}

// captureFileSnapshot records all files in directory and subdirectories
func (t *Trainer) captureFileSnapshot(folderPath string) (map[string]FileSnapshot, error) {
	snapshot := make(map[string]FileSnapshot)

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		snapshot[path] = FileSnapshot{
			Path:    path,
			ModTime: info.ModTime(),
			Size:    info.Size(),
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to capture snapshot: %w", err)
	}

	println("📸 [SNAPSHOT] Captured", len(snapshot), "files in", folderPath)
	return snapshot, nil
}

// detectNewOrModifiedModels compares before/after snapshots and returns changed model files
func (t *Trainer) detectNewOrModifiedModels(before, after map[string]FileSnapshot) []string {
	// Common model file extensions across frameworks
	modelExtensions := []string{
		".pth", ".pt",           // PyTorch
		".h5", ".keras",         // TensorFlow/Keras
		".pkl", ".pickle",       // scikit-learn, general Python
		".ckpt",                 // TensorFlow checkpoints
		".pb",                   // TensorFlow protobuf
		".onnx",                 // ONNX
		".safetensors",          // Hugging Face
		".joblib",               // scikit-learn
		".model",                // Generic
	}

	var changedModels []string

	for path, afterFile := range after {
		beforeFile, existed := before[path]

		// Check if it's a model file
		isModel := false
		ext := filepath.Ext(path)
		for _, modelExt := range modelExtensions {
			if ext == modelExt {
				isModel = true
				break
			}
		}

		if !isModel {
			continue
		}

		// New file or modified file
		if !existed {
			changedModels = append(changedModels, path)
			println("🆕 [DETECT] New model file:", filepath.Base(path))
		} else if afterFile.ModTime.After(beforeFile.ModTime) || afterFile.Size != beforeFile.Size {
			changedModels = append(changedModels, path)
			println("♻️  [DETECT] Modified model file:", filepath.Base(path))
		}
	}

	return changedModels
}

// selectBestModel picks the most likely "final" model from a list of candidates
func (t *Trainer) selectBestModel(changedModels []string) string {
	if len(changedModels) == 0 {
		return ""
	}

	if len(changedModels) == 1 {
		return changedModels[0]
	}

	println("🤔 [SELECT] Multiple models detected, selecting best one...")

	// Priority 1: Look for "best", "final", or "trained" in filename
	for _, path := range changedModels {
		basename := filepath.Base(path)
		basenameLower := filepath.Base(filepath.Base(path))
		if containsAny(basenameLower, []string{"best", "final", "trained"}) {
			println("✨ [SELECT] Selected by keyword:", basename)
			return path
		}
	}

	// Priority 2: Prefer files in standard output directories
	for _, path := range changedModels {
		if containsAny(path, []string{"saved_models", "outputs", "checkpoints", "models"}) {
			println("📁 [SELECT] Selected from standard directory:", filepath.Base(path))
			return path
		}
	}

	// Priority 3: Largest file (usually the final model, not a checkpoint)
	var largestPath string
	var largestSize int64
	for _, path := range changedModels {
		if info, err := os.Stat(path); err == nil {
			if info.Size() > largestSize {
				largestSize = info.Size()
				largestPath = path
			}
		}
	}

	if largestPath != "" {
		println("📏 [SELECT] Selected largest file:", filepath.Base(largestPath), fmt.Sprintf("(%.2f MB)", float64(largestSize)/1024/1024))
		return largestPath
	}

	// Fallback: Return the last (newest by modification time) model
	var newestPath string
	var newestTime time.Time
	for _, path := range changedModels {
		if info, err := os.Stat(path); err == nil {
			if info.ModTime().After(newestTime) {
				newestTime = info.ModTime()
				newestPath = path
			}
		}
	}

	if newestPath != "" {
		println("⏰ [SELECT] Selected newest file:", filepath.Base(newestPath))
		return newestPath
	}

	return changedModels[len(changedModels)-1]
}

// containsAny checks if string contains any of the substrings
func containsAny(s string, substrings []string) bool {
	sLower := filepath.Base(s)
	for _, substr := range substrings {
		if contains(sLower, substr) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

// hasSubstring performs case-insensitive substring search
func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if matchesAt(s, substr, i) {
			return true
		}
	}
	return false
}

// matchesAt checks if substr matches s starting at position i (case-insensitive)
func matchesAt(s, substr string, i int) bool {
	for j := 0; j < len(substr); j++ {
		c1 := s[i+j]
		c2 := substr[j]
		if toLower(c1) != toLower(c2) {
			return false
		}
	}
	return true
}

// toLower converts a byte to lowercase
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}
