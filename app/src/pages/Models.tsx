import { useContext, useState } from "react";
import { Plus, Cpu, HardDrive, Network, Trash2, Play } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ModelContext } from "@/context/modelContext";
import { TrainingContext } from "@/context/trainingContext";
import { useNavigate } from "react-router-dom";
import { useToast } from "@/hooks/use-toast";

const Models = () => {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [modelToDelete, setModelToDelete] = useState<{ id: number; name: string } | null>(null);
  const [trainingModel, setTrainingModel] = useState<number | null>(null);

  const {
    name, picture, folder, trainingScript,
    setName, setPicture, setFolder, setTrainingScript,
    send, deleteModel, models, loading
  } = useContext(ModelContext)!;

  const trainingContext = useContext(TrainingContext);
  const navigate = useNavigate();
  const { toast } = useToast();

  const handleSubmit = async () => {
    await send();
    setIsDialogOpen(false); // Close modal on success
  };

  const handleDeleteClick = (modelId: number, modelName: string) => {
    setModelToDelete({ id: modelId, name: modelName });
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (modelToDelete) {
      await deleteModel(modelToDelete.id, modelToDelete.name);
      setDeleteDialogOpen(false);
      setModelToDelete(null);
    }
  };

  const handleTrainModel = async (modelId: number, modelName: string, scriptPath?: string) => {
    if (!trainingContext) return;

    setTrainingModel(modelId);

    // Use provided script path, or fallback to "train.py"
    const script = scriptPath || "train.py";
    // const script = "mock_train.py"
    try {
      toast({
        title: "Starting Training",
        description: `Training ${modelName} with ${script}...`,
      });

      const trainingId = await trainingContext.startTraining(modelName, script);

      if (trainingId) {
        toast({
          title: "Training Started!",
          description: "Redirecting to statistics page...",
          variant: "default",
        });

        // Navigate to statistics page after a short delay
        setTimeout(() => {
          navigate("/statistics");
        }, 1500);
      } else {
        toast({
          title: "Training Failed",
          description: trainingContext.error || "Could not start training",
          variant: "destructive",
        });
      }
    } catch (error) {
      console.error("Training error:", error);
      toast({
        title: "Error",
        description: "Failed to start training",
        variant: "destructive",
      });
    } finally {
      setTrainingModel(null);
    }
  };

  return (
    <div className="space-y-6 animate-slide-up">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">AI Models</h2>
          <p className="text-muted-foreground mt-1">
            Manage your AI models and configurations
          </p>
        </div>

        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button className="bg-gradient-primary hover:opacity-90 shadow-glow">
              <Plus className="w-4 h-4 mr-2" />
              Add Model
            </Button>
          </DialogTrigger>

          <DialogContent className="bg-card border-border">
            <DialogHeader>
              <DialogTitle className="text-2xl">Add new Model</DialogTitle>
              <DialogDescription>
                Configure a new AI model for your system
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Model Name</Label>
                <Input
                  id="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g., GPT-4 Turbo"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="picture">Profile Image</Label>
                <Input
                  id="picture"
                  type="file"
                  onChange={(e) => setPicture(e.target.files?.[0] || null)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="folder">Model Folder (multiple files)</Label>
                <Input
                  id="folder"
                  type="file"
                  multiple
                  onChange={(e) => setFolder(e.target.files ? Array.from(e.target.files) : null)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="trainingScript">Training Script Path</Label>
                <Input
                  id="trainingScript"
                  value={trainingScript}
                  onChange={(e) => setTrainingScript(e.target.value)}
                  placeholder="e.g., train.py or PokemonModel/train.py"
                />
                <p className="text-xs text-muted-foreground">
                  Path to the training script relative to the model folder
                </p>
              </div>
            </div>

            <Button
              className="w-full bg-gradient-primary hover:opacity-90"
              onClick={handleSubmit}
              disabled={loading}
            >
              {loading ? "Uploading..." : "Submit"}
            </Button>
          </DialogContent>
        </Dialog>
      </div>

      {/* Show Models from WebSocket */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {models.map((model) => (
          <Card
            key={model.id}
            className="bg-gradient-card border-border hover:border-primary/50 transition-all shadow-card hover:shadow-glow group"
          >
            <CardHeader>
              <div className="flex items-start justify-between">
                {model.picture ? (
                  <div className="w-16 h-16 rounded-xl overflow-hidden shadow-lg ring-2 ring-primary/20 group-hover:ring-primary/40 transition-all">
                    <img
                      src={`http://localhost:8081${model.picture}`}
                      alt={model.name}
                      className="w-full h-full object-cover"
                    />
                  </div>
                ) : (
                  <div className="w-16 h-16 rounded-xl bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center group-hover:from-primary/30 group-hover:to-secondary/30 transition-all shadow-lg">
                    <Cpu className="w-8 h-8 text-primary" />
                  </div>
                )}

                <div className="flex items-center gap-2">
                  <Badge className="bg-primary/20 text-primary border-primary/30">
                    Synced
                  </Badge>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-primary hover:text-primary hover:bg-primary/10"
                    onClick={() => handleTrainModel(model.id, model.name, model.training_script)}
                    disabled={trainingModel === model.id || loading}
                    title={`Train Model${model.training_script ? ` (${model.training_script})` : ''}`}
                  >
                    <Play className={`w-4 h-4 ${trainingModel === model.id ? 'animate-pulse' : ''}`} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                    onClick={() => handleDeleteClick(model.id, model.name)}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>

              <CardTitle className="mt-4">{model.name}</CardTitle>
              <CardDescription className="flex items-center gap-1 text-xs">
                <HardDrive className="w-3 h-3" />
                {model.folder?.length || 0} files uploaded
              </CardDescription>
            </CardHeader>

            <CardContent>
              <div className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Network className="w-4 h-4" />
                  Model ID
                </div>
                <span className="font-medium text-primary truncate max-w-[120px]">
                  {model.id}
                </span>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="bg-card border-border">
          <DialogHeader>
            <DialogTitle className="text-2xl text-destructive">Delete Model</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete <span className="font-semibold text-foreground">{modelToDelete?.name}</span>?
              This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="flex gap-3 justify-end pt-4">
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteConfirm}
              disabled={loading}
              className="bg-destructive hover:bg-destructive/90"
            >
              {loading ? "Deleting..." : "Delete"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default Models;
