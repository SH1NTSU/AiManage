import { useContext, useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, TrendingUp, Clock, Zap, Play, CheckCircle, AlertCircle, Loader2, Download, RefreshCw, Cloud, FolderOpen } from "lucide-react";
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { TrainingContext } from "@/context/trainingContext";
import { ModelContext } from "@/context/modelContext";
import { SubscriptionContext } from "@/context/subscriptionContext";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { SmoothProgressBar } from "@/components/SmoothProgressBar";
import { useToast } from "@/hooks/use-toast";

const Statistics = () => {
  const trainingContext = useContext(TrainingContext);
  const modelContext = useContext(ModelContext);
  const subscriptionContext = useContext(SubscriptionContext);
  const { toast } = useToast();
  const [isPolling, setIsPolling] = useState(false);
  const [latestTrainingId, setLatestTrainingId] = useState<string | null>(null);

  // Extract model name from training ID (format: "ModelName_timestamp")
  const getModelNameFromTrainingId = (trainingId: string): string => {
    return trainingId.split('_')[0];
  };

  // Find model by name
  const findModelByName = (modelName: string) => {
    return modelContext?.models.find(m => m.name === modelName);
  };

  // Refresh models manually
  const handleRefreshModels = async () => {
    // The modelContext should automatically fetch via WebSocket, but we can trigger a re-render
    if (modelContext) {
      toast({
        title: "Refreshing Models",
        description: "Fetching latest model data...",
      });
    }
  };

  // Download trained model
  const handleDownloadModel = (modelId: number, modelName: string) => {
    const token = localStorage.getItem("token");
    if (!token) {
      toast({
        title: "Error",
        description: "You must be logged in to download models",
        variant: "destructive"
      });
      return;
    }

    const downloadUrl = `http://localhost:8081/v1/downloadModel?model_id=${modelId}`;

    // Create a temporary link and trigger download
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.setAttribute('download', ''); // This suggests to download

    // Add auth header by fetching first, then creating blob URL
    fetch(downloadUrl, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
      .then(response => {
        if (!response.ok) throw new Error('Download failed');
        return response.blob();
      })
      .then(blob => {
        const url = window.URL.createObjectURL(blob);
        link.href = url;
        link.click();
        window.URL.revokeObjectURL(url);

        toast({
          title: "Download Started",
          description: `Downloading ${modelName}...`,
        });
      })
      .catch(error => {
        console.error('Download error:', error);
        toast({
          title: "Download Failed",
          description: "Failed to download the trained model",
          variant: "destructive"
        });
      });
  };

  // Poll for training progress
  useEffect(() => {
    if (!trainingContext) return;

    const pollProgress = async () => {
      await trainingContext.getAllTrainings();

      // Get the latest training
      if (trainingContext.trainings.size > 0) {
        const entries = Array.from(trainingContext.trainings.entries());
        const latest = entries[entries.length - 1];
        setLatestTrainingId(latest[0]);

        // Check if still running
        const isRunning = latest[1].status === "running";
        setIsPolling(isRunning);

        // If completed, analyze it
        if (latest[1].status === "completed" && !trainingContext.metrics) {
          await trainingContext.analyzeResults(latest[0], false);
        }
      } else {
        // No trainings, stop polling
        setIsPolling(false);
      }
    };

    // Initial fetch
    pollProgress();

    // Set up polling interval that checks current state
    const interval = setInterval(() => {
      pollProgress();
    }, 2000); // Poll every 2 seconds

    return () => clearInterval(interval);
  }, [trainingContext]); // Removed isPolling from dependencies to prevent interval recreation

  if (!trainingContext) {
    return <div>Loading...</div>;
  }

  const { trainings, metrics, loading } = trainingContext;

  // Get current training
  const currentTraining = latestTrainingId ? trainings.get(latestTrainingId) : null;

  // Show metrics if available
  if (metrics) {
    // Get the model info for download button
    const modelName = latestTrainingId ? getModelNameFromTrainingId(latestTrainingId) : null;
    const model = modelName ? findModelByName(modelName) : null;


    return (
      <div className="space-y-6 animate-slide-up">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Training Results</h2>
            <p className="text-muted-foreground mt-1">
              Comprehensive performance metrics and insights
            </p>
          </div>
          <div className="flex gap-3">
            {model && model.trained_model_path ? (
              <Button
                className="bg-gradient-primary hover:opacity-90 shadow-glow"
                onClick={() => handleDownloadModel(model.id, model.name)}
              >
                <Download className="w-4 h-4 mr-2" />
                Download Trained Model
              </Button>
            ) : (
              <Button
                variant="outline"
                onClick={handleRefreshModels}
                className="border-primary/30"
              >
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh (Model might be ready)
              </Button>
            )}
            <Button
              variant="outline"
              onClick={() => trainingContext.setMetrics(null)}
            >
              Back to Overview
            </Button>
          </div>
        </div>

        {/* Trained Model Info Card */}
        {model && model.trained_model_path && (
          <Card className="bg-gradient-to-br from-primary/10 to-secondary/10 border-primary/30 shadow-glow">
            <CardContent className="pt-6">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-lg bg-primary/20 flex items-center justify-center">
                    <CheckCircle className="w-6 h-6 text-primary" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-lg">Trained Model Ready</h3>
                    <p className="text-sm text-muted-foreground">
                      {model.trained_model_path.split('/').pop()}
                    </p>
                    {model.trained_at && (
                      <p className="text-xs text-muted-foreground mt-1">
                        Trained: {new Date(model.trained_at).toLocaleString()}
                      </p>
                    )}
                  </div>
                </div>
                <Button
                  size="sm"
                  className="bg-primary hover:bg-primary/90"
                  onClick={() => handleDownloadModel(model.id, model.name)}
                >
                  <Download className="w-4 h-4 mr-2" />
                  Download
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Performance Overview Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Overall Score</CardTitle>
              <TrendingUp className="w-4 h-4 text-primary" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-primary">
                {metrics.overall_score !== undefined ? metrics.overall_score.toFixed(1) : 'N/A'}/100
              </div>
              <Badge className={`mt-2 ${
                metrics.performance_level === 'excellent' ? 'bg-green-500/20 text-green-500' :
                metrics.performance_level === 'good' ? 'bg-blue-500/20 text-blue-500' :
                metrics.performance_level === 'fair' ? 'bg-yellow-500/20 text-yellow-500' :
                'bg-red-500/20 text-red-500'
              }`}>
                {metrics.performance_level?.toUpperCase() || 'UNKNOWN'}
              </Badge>
            </CardContent>
          </Card>

          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Final Accuracy</CardTitle>
              <CheckCircle className="w-4 h-4 text-secondary" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-secondary">
                {/* Prefer test_accuracy, fall back to final_val_accuracy */}
                {metrics.test_accuracy !== undefined && metrics.test_accuracy > 0
                  ? metrics.test_accuracy.toFixed(2)
                  : metrics.final_val_accuracy !== undefined
                  ? metrics.final_val_accuracy.toFixed(2)
                  : 'N/A'}%
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                +{metrics.accuracy_improvement_percent !== undefined ? metrics.accuracy_improvement_percent.toFixed(1) : 'N/A'}% improvement
              </p>
            </CardContent>
          </Card>

          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Training Time</CardTitle>
              <Clock className="w-4 h-4 text-chart-3" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold" style={{ color: "hsl(280, 70%, 60%)" }}>
                {metrics.total_duration_seconds !== undefined
                  ? (metrics.total_duration_seconds / 60).toFixed(1)
                  : 'N/A'}m
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {metrics.average_epoch_time_seconds !== undefined
                  ? `${metrics.average_epoch_time_seconds.toFixed(1)}s per epoch`
                  : 'N/A'}
              </p>
            </CardContent>
          </Card>

          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Loss Reduction</CardTitle>
              <Zap className="w-4 h-4 text-chart-4" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold" style={{ color: "hsl(45, 100%, 60%)" }}>
                {metrics.loss_improvement_percent !== undefined
                  ? `${metrics.loss_improvement_percent.toFixed(1)}%`
                  : 'N/A'}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {metrics.is_converging !== undefined
                  ? (metrics.is_converging ? '✓ Converged' : '⚠ Not converged')
                  : 'Unknown'}
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Loss Chart */}
          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader>
              <CardTitle>Loss Over Time</CardTitle>
              <CardDescription>Training and validation loss progression</CardDescription>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={metrics.epoch_data}>
                  <CartesianGrid strokeDasharray="3 3" stroke="hsl(0, 0%, 20%)" />
                  <XAxis dataKey="epoch" stroke="hsl(0, 0%, 60%)" />
                  <YAxis stroke="hsl(0, 0%, 60%)" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "hsl(0, 0%, 10%)",
                      border: "1px solid hsl(0, 0%, 20%)",
                      borderRadius: "0.5rem",
                    }}
                  />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="train_loss"
                    name="Train Loss"
                    stroke="hsl(180, 80%, 50%)"
                    strokeWidth={2}
                    dot={{ r: 4 }}
                  />
                  <Line
                    type="monotone"
                    dataKey="val_loss"
                    name="Val Loss"
                    stroke="hsl(210, 100%, 60%)"
                    strokeWidth={2}
                    dot={{ r: 4 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* Accuracy Chart */}
          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader>
              <CardTitle>Accuracy Over Time</CardTitle>
              <CardDescription>Training and validation accuracy progression</CardDescription>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <AreaChart data={metrics.epoch_data}>
                  <defs>
                    <linearGradient id="colorTrain" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="hsl(180, 80%, 50%)" stopOpacity={0.8}/>
                      <stop offset="95%" stopColor="hsl(180, 80%, 50%)" stopOpacity={0}/>
                    </linearGradient>
                    <linearGradient id="colorVal" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="hsl(210, 100%, 60%)" stopOpacity={0.8}/>
                      <stop offset="95%" stopColor="hsl(210, 100%, 60%)" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="hsl(0, 0%, 20%)" />
                  <XAxis dataKey="epoch" stroke="hsl(0, 0%, 60%)" />
                  <YAxis domain={[0, 100]} stroke="hsl(0, 0%, 60%)" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "hsl(0, 0%, 10%)",
                      border: "1px solid hsl(0, 0%, 20%)",
                      borderRadius: "0.5rem",
                    }}
                  />
                  <Legend />
                  <Area
                    type="monotone"
                    dataKey="train_accuracy"
                    name="Train Accuracy"
                    stroke="hsl(180, 80%, 50%)"
                    fillOpacity={1}
                    fill="url(#colorTrain)"
                  />
                  <Area
                    type="monotone"
                    dataKey="val_accuracy"
                    name="Val Accuracy"
                    stroke="hsl(210, 100%, 60%)"
                    fillOpacity={1}
                    fill="url(#colorVal)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </div>

        {/* Insights & Recommendations */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Insights */}
          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <CheckCircle className="w-5 h-5 text-green-500" />
                Insights
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-2">
                {metrics.insights.map((insight, i) => (
                  <li key={i} className="text-sm text-muted-foreground flex items-start gap-2">
                    <span className="text-green-500 mt-1">✓</span>
                    {insight}
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>

          {/* Warnings */}
          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <AlertCircle className="w-5 h-5 text-yellow-500" />
                Warnings
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-2">
                {metrics.warnings.length > 0 ? (
                  metrics.warnings.map((warning, i) => (
                    <li key={i} className="text-sm text-muted-foreground flex items-start gap-2">
                      <span className="text-yellow-500 mt-1">⚠</span>
                      {warning}
                    </li>
                  ))
                ) : (
                  <p className="text-sm text-muted-foreground">No warnings</p>
                )}
              </ul>
            </CardContent>
          </Card>

          {/* Recommendations */}
          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <TrendingUp className="w-5 h-5 text-blue-500" />
                Recommendations
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-2">
                {metrics.recommendations.map((rec, i) => (
                  <li key={i} className="text-sm text-muted-foreground flex items-start gap-2">
                    <span className="text-blue-500 mt-1">→</span>
                    {rec}
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  // Show training progress
  if (currentTraining && currentTraining.status === "running") {
    const progress = currentTraining.total_epochs > 0
      ? (currentTraining.current_epoch / currentTraining.total_epochs) * 100
      : 0;

    const latestMetric = currentTraining.metrics[currentTraining.metrics.length - 1];

    return (
      <div className="space-y-6 animate-slide-up">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Training in Progress</h2>
            <p className="text-muted-foreground mt-1">
              Monitoring training metrics in real-time
            </p>
          </div>
          <div className="flex items-center gap-2">
            {subscriptionContext?.canTrainOnServer ? (
              <Badge className="bg-gradient-to-r from-blue-500 to-purple-500 text-white">
                <Cloud className="w-3 h-3 mr-1" />
                Server Training
              </Badge>
            ) : (
              <Badge variant="secondary">
                <FolderOpen className="w-3 h-3 mr-1" />
                Local Training
              </Badge>
            )}
          </div>
        </div>

        {/* Progress Card */}
        <Card className="bg-gradient-card border-border shadow-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Loader2 className="w-5 h-5 animate-spin text-primary" />
              Training Progress
            </CardTitle>
            <CardDescription>
              Epoch {currentTraining.current_epoch} / {currentTraining.total_epochs}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <SmoothProgressBar
              targetProgress={progress}
              animationSpeed={500}
              showPercentage={true}
            />

            {latestMetric && (
              <div className="grid grid-cols-2 gap-4 mt-4">
                <div>
                  <p className="text-sm text-muted-foreground">Train Loss</p>
                  <p className="text-2xl font-bold text-primary">
                    {latestMetric.train_loss !== undefined ? latestMetric.train_loss.toFixed(4) : 'N/A'}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Train Accuracy</p>
                  <p className="text-2xl font-bold text-secondary">
                    {latestMetric.train_accuracy !== undefined
                      ? `${(latestMetric.train_accuracy * 100).toFixed(2)}%`
                      : 'N/A'}
                  </p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Live Log */}
        <Card className="bg-gradient-card border-border shadow-card">
          <CardHeader>
            <CardTitle>Training Logs</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="bg-black text-green-400 p-4 rounded font-mono text-sm h-64 overflow-y-auto">
              {currentTraining.logs.slice(-20).map((log, i) => (
                <div key={i}>{log}</div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Default: No training data
  return (
    <div className="space-y-6 animate-slide-up">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Statistics</h2>
        <p className="text-muted-foreground mt-1">
          Train a model to see performance metrics
        </p>
      </div>

      <Card className="bg-gradient-card border-border shadow-card">
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Play className="w-16 h-16 text-muted-foreground mb-4" />
          <h3 className="text-xl font-semibold mb-2">No Training Data</h3>
          <p className="text-muted-foreground text-center max-w-md mb-4">
            Start training a model from the Models page to see comprehensive performance metrics and insights.
          </p>

          {/* Subscription Tier Info */}
          <div className="flex items-center gap-2 mb-6">
            {subscriptionContext?.canTrainOnServer ? (
              <Badge className="bg-gradient-to-r from-blue-500 to-purple-500 text-white">
                <Cloud className="w-3 h-3 mr-1" />
                Server Training Available
              </Badge>
            ) : (
              <Badge variant="secondary">
                <FolderOpen className="w-3 h-3 mr-1" />
                Free - Local Training Only
              </Badge>
            )}
            {subscriptionContext?.isAgentConnected && !subscriptionContext?.canTrainOnServer && (
              <Badge variant="default" className="bg-green-500">
                <CheckCircle className="w-3 h-3 mr-1" />
                Agent Connected
              </Badge>
            )}
          </div>

          <Button className="mt-2" onClick={() => window.location.href = '/'}>
            Go to Models
          </Button>
        </CardContent>
      </Card>
    </div>
  );
};

export default Statistics;
