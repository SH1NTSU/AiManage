package aiAgent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"
)

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

	defer func() {
		endTime := time.Now()
		progress.mu.Lock()
		progress.EndTime = &endTime
		if progress.Status == StatusRunning {
			progress.Status = StatusCompleted
			println("✅ [EXECUTE] Training completed successfully")
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

	// Prepare command
	pythonCmd := req.PythonCommand
	if pythonCmd == "" {
		pythonCmd = "python3"
	}

	scriptPath := filepath.Join(t.navigator.BaseUploadPath, req.FolderName, req.ScriptName)
	workingDir := filepath.Join(t.navigator.BaseUploadPath, req.FolderName)

	// Check and install requirements before training
	if err := t.installRequirements(ctx, workingDir, progress); err != nil {
		println("⚠️  [EXECUTE] Requirements installation warning:", err.Error())
		// Don't fail training if requirements installation fails - the packages might already be installed
	}

	println("📍 [EXECUTE] Working directory:", workingDir)
	println("🐍 [EXECUTE] Python command:", pythonCmd)
	println("📜 [EXECUTE] Script path:", scriptPath)

	// Use only the script name since we're setting the working directory
	args := append([]string{req.ScriptName}, req.Args...)
	println("🔧 [EXECUTE] Full command:", pythonCmd, args)

	cmd := exec.CommandContext(ctx, pythonCmd, args...)
	cmd.Dir = workingDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, val := range req.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}

	// Create pipes for stdout and stderr
	println("📡 [EXECUTE] Creating output pipes...")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		println("❌ [EXECUTE] Failed to create stdout pipe:", err.Error())
		t.setError(progress, fmt.Errorf("failed to create stdout pipe: %w", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		println("❌ [EXECUTE] Failed to create stderr pipe:", err.Error())
		t.setError(progress, fmt.Errorf("failed to create stderr pipe: %w", err))
		return
	}

	// Start command
	println("🚀 [EXECUTE] Starting Python process...")
	if err := cmd.Start(); err != nil {
		println("❌ [EXECUTE] Failed to start process:", err.Error())
		t.setError(progress, fmt.Errorf("failed to start training: %w", err))
		return
	}
	println("✅ [EXECUTE] Python process started successfully!")

	// Read output in goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	println("👀 [EXECUTE] Starting output readers...")
	go func() {
		defer wg.Done()
		t.readOutput(stdout, progress, false)
	}()

	go func() {
		defer wg.Done()
		t.readOutput(stderr, progress, true)
	}()

	wg.Wait()
	println("📖 [EXECUTE] Finished reading output")

	// Wait for command to finish
	println("⏳ [EXECUTE] Waiting for process to complete...")
	if err := cmd.Wait(); err != nil {
		println("❌ [EXECUTE] Process failed:", err.Error())
		t.setError(progress, fmt.Errorf("training failed: %w", err))
		return
	}

	// Training completed successfully
	progress.mu.Lock()
	progress.Status = StatusCompleted
	progress.mu.Unlock()
}

// readOutput reads and processes output from the training script
func (t *Trainer) readOutput(reader io.Reader, progress *TrainingProgress, isError bool) {
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
func (t *Trainer) setError(progress *TrainingProgress, err error) {
	progress.mu.Lock()
	defer progress.mu.Unlock()
	progress.Status = StatusFailed
	progress.ErrorMessage = err.Error()
	endTime := time.Now()
	progress.EndTime = &endTime
}

// installRequirements checks for requirements.txt and installs dependencies
func (t *Trainer) installRequirements(ctx context.Context, workingDir string, progress *TrainingProgress) error {
	// Check if requirements.txt exists
	requirementsPath := filepath.Join(workingDir, "requirements.txt")
	if _, err := os.Stat(requirementsPath); os.IsNotExist(err) {
		println("ℹ️  [INSTALL] No requirements.txt found, skipping installation")
		return nil
	}

	println("📦 [INSTALL] Found requirements.txt, installing dependencies...")
	progress.mu.Lock()
	progress.Logs = append(progress.Logs, "📦 Installing Python dependencies from requirements.txt...")
	progress.mu.Unlock()

	// Run pip install -r requirements.txt
	cmd := exec.CommandContext(ctx, "python3", "-m", "pip", "install", "-r", "requirements.txt")
	cmd.Dir = workingDir

	// Capture output
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Log the installation output
	progress.mu.Lock()
	progress.Logs = append(progress.Logs, outputStr)
	progress.mu.Unlock()

	println("📦 [INSTALL] Installation output:")
	println(outputStr)

	if err != nil {
		println("❌ [INSTALL] Installation failed:", err.Error())
		progress.mu.Lock()
		progress.Logs = append(progress.Logs, fmt.Sprintf("⚠️  Dependency installation failed: %s", err.Error()))
		progress.mu.Unlock()
		return fmt.Errorf("failed to install requirements: %w", err)
	}

	println("✅ [INSTALL] Dependencies installed successfully")
	progress.mu.Lock()
	progress.Logs = append(progress.Logs, "✅ Dependencies installed successfully")
	progress.mu.Unlock()

	return nil
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
