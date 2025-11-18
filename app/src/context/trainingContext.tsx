import axios from "axios";
import { createContext, type ReactNode, useState, useCallback } from "react";

const API_BASE = "http://localhost:8081/v1";

// Training Types
export interface TrainingMetrics {
  epoch: number;
  train_loss: number;
  val_loss: number;
  train_accuracy: number;
  val_accuracy: number;
  duration_seconds?: number;
}

export interface TrainingProgress {
  status: "pending" | "running" | "completed" | "failed";
  current_epoch: number;
  total_epochs: number;
  start_time: string;
  end_time?: string;
  logs: string[];
  metrics: TrainingMetrics[];
  error_message?: string;
  model_path?: string;
}

export interface DetailedMetrics {
  // Overview
  training_status: string;
  total_duration_seconds: number;
  completed_epochs: number;
  total_epochs: number;
  average_epoch_time_seconds: number;

  // Performance
  overall_score: number;
  performance_level: string; // "excellent" | "good" | "fair" | "poor"

  // Loss
  initial_loss: number;
  final_loss: number;
  best_loss: number;
  loss_improvement_percent: number;

  // Accuracy
  final_accuracy: number;
  final_val_accuracy: number;
  test_accuracy?: number; // Test accuracy from FinalMetrics
  initial_accuracy: number;
  accuracy_improvement_percent: number;

  // Behavior
  is_converging: boolean;
  is_overfitting: boolean;
  is_underfitting: boolean;
  train_val_gap: number;
  loss_variability: string;

  // Chart Data
  loss_history: number[];
  val_loss_history: number[];
  accuracy_history: number[];
  val_accuracy_history: number[];
  epoch_data: TrainingMetrics[];

  // Insights
  insights: string[];
  warnings: string[];
  recommendations: string[];
}

export interface TrainingContextType {
  // State
  trainings: Map<string, TrainingProgress>;
  selectedTraining: string | null;
  metrics: DetailedMetrics | null;
  loading: boolean;
  error: string | null;

  // Actions
  startTraining: (folderName: string, scriptName?: string) => Promise<string | null>;
  getProgress: (trainingId?: string) => Promise<TrainingProgress | null>;
  getAllTrainings: () => Promise<void>;
  analyzeResults: (trainingId: string, useAI?: boolean) => Promise<DetailedMetrics | null>;
  selectTraining: (trainingId: string) => void;
  setMetrics: (metrics: DetailedMetrics | null) => void;
  clearError: () => void;
}

export const TrainingContext = createContext<TrainingContextType | null>(null);

export const TrainingProvider = ({ children }: { children: ReactNode }) => {
  const [trainings, setTrainings] = useState<Map<string, TrainingProgress>>(new Map());
  const [selectedTraining, setSelectedTraining] = useState<string | null>(null);
  const [metrics, setMetrics] = useState<DetailedMetrics | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Start Training
  const startTraining = useCallback(async (
    folderName: string,
    scriptName: string = "train.py"
  ): Promise<string | null> => {
    setLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem("token");
      if (!token) {
        setError("No authentication token");
        return null;
      }

      const response = await axios.post(
        `${API_BASE}/train/start`,
        {
          folder_name: folderName,
          script_name: scriptName,
          python_command: "python3"
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json"
          }
        }
      );


      // Extract training ID from response or create one
      const trainingId = `${folderName}_${Date.now()}`;

      if (response.data.progress) {
        setTrainings(prev => new Map(prev).set(trainingId, response.data.progress));
        setSelectedTraining(trainingId);
      }

      return trainingId;
    } catch (err: any) {
      console.error("Failed to start training:", err);
      setError(err.response?.data?.message || "Failed to start training");
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  // Get Training Progress
  const getProgress = useCallback(async (
    trainingId?: string
  ): Promise<TrainingProgress | null> => {
    try {
      const token = localStorage.getItem("token");
      if (!token) return null;

      const url = trainingId
        ? `${API_BASE}/train/progress?id=${trainingId}`
        : `${API_BASE}/train/progress`;

      const response = await axios.get(url, {
        headers: { Authorization: `Bearer ${token}` }
      });

      if (trainingId) {
        // Single training
        const progress = response.data.progress;
        setTrainings(prev => new Map(prev).set(trainingId, progress));
        return progress;
      } else {
        // All trainings
        const trainingsData = response.data.trainings || {};
        const newMap = new Map<string, TrainingProgress>();

        Object.entries(trainingsData).forEach(([id, prog]) => {
          newMap.set(id, prog as TrainingProgress);
        });

        setTrainings(newMap);
        return null;
      }
    } catch (err: any) {
      console.error("Failed to get progress:", err);
      return null;
    }
  }, []);

  // Get All Trainings
  const getAllTrainings = useCallback(async () => {
    await getProgress();
  }, [getProgress]);

  // Analyze Results
  const analyzeResults = useCallback(async (
    trainingId: string,
    useAI: boolean = false
  ): Promise<DetailedMetrics | null> => {
    setLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem("token");
      if (!token) {
        setError("No authentication token");
        return null;
      }

      const response = await axios.post(
        `${API_BASE}/train/analyze`,
        {
          training_id: trainingId,
          use_ai: useAI
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json"
          }
        }
      );


      const metricsData = response.data.metrics || response.data.analysis;
      setMetrics(metricsData);
      return metricsData;
    } catch (err: any) {
      console.error("Failed to analyze results:", err);
      setError(err.response?.data?.message || "Failed to analyze results");
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  // Select Training
  const selectTraining = useCallback((trainingId: string) => {
    setSelectedTraining(trainingId);
  }, []);

  // Set Metrics (for manual control)
  const setMetricsManually = useCallback((metricsData: DetailedMetrics | null) => {
    setMetrics(metricsData);
  }, []);

  // Clear Error
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return (
    <TrainingContext.Provider
      value={{
        trainings,
        selectedTraining,
        metrics,
        loading,
        error,
        startTraining,
        getProgress,
        getAllTrainings,
        analyzeResults,
        selectTraining,
        setMetrics: setMetricsManually,
        clearError
      }}
    >
      {children}
    </TrainingContext.Provider>
  );
};
