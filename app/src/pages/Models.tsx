import { useContext, useState, useEffect } from "react";
import { Plus, Cpu, HardDrive, Network, Trash2, Play, Upload, Store, Star, Download, FolderOpen, Cloud, AlertCircle, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { ModelContext } from "@/context/modelContext";
import { TrainingContext } from "@/context/trainingContext";
import { SubscriptionContext } from "@/context/subscriptionContext";
import { useNavigate } from "react-router-dom";
import { useToast } from "@/hooks/use-toast";

const Models = () => {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [publishDialogOpen, setPublishDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [modelToDelete, setModelToDelete] = useState<{ id: number; name: string } | null>(null);
  const [modelToPublish, setModelToPublish] = useState<any>(null);
  const [trainingModel, setTrainingModel] = useState<number | null>(null);
  const [myPublishedModels, setMyPublishedModels] = useState<any[]>([]);
  const [trainingMode, setTrainingMode] = useState<"local" | "server">("local");

  // Publish form state
  const [publishForm, setPublishForm] = useState({
    description: "",
    price: 0,
    license_type: "personal_use",
    category: "",
    tags: "",
    model_type: "",
    framework: "",
  });

  const {
    name, picture, folder, folderPath, trainingScript,
    setName, setPicture, setFolder, setFolderPath, setTrainingScript,
    send, deleteModel, models, loading
  } = useContext(ModelContext)!;

  const trainingContext = useContext(TrainingContext);
  const subscriptionContext = useContext(SubscriptionContext);
  const navigate = useNavigate();
  const { toast } = useToast();

  // Fetch user's published models
  useEffect(() => {
    fetchMyPublishedModels();
  }, []);

  const fetchMyPublishedModels = async () => {
    try {
      const token = localStorage.getItem("token");
      if (!token) return;

      const response = await fetch("http://localhost:8081/v1/my-published-models", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error("Failed to fetch published models");
      }

      const data = await response.json();
      setMyPublishedModels(data || []);
    } catch (error) {
      console.error("Error fetching published models:", error);
    }
  };

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

  const handlePublishClick = (model: any) => {
    setModelToPublish(model);
    setPublishDialogOpen(true);
  };

  const handlePublishSubmit = async () => {
    if (!modelToPublish) return;

    try {
      // TODO: Call publish API endpoint
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Error",
          description: "You must be logged in to publish models",
          variant: "destructive",
        });
        return;
      }

      // Prepare payload
      const payload: any = {
        model_id: modelToPublish.id,
        description: publishForm.description,
        price: publishForm.price,
        license_type: publishForm.license_type,
      };

      // Add optional fields if provided
      if (publishForm.category) payload.category = publishForm.category;
      if (publishForm.tags) payload.tags = publishForm.tags.split(',').map(t => t.trim()).filter(t => t);
      if (publishForm.model_type) payload.model_type = publishForm.model_type;
      if (publishForm.framework) payload.framework = publishForm.framework;

      console.log("Publishing model:", payload);

      const response = await fetch("http://localhost:8081/v1/publish", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || "Failed to publish model");
      }

      const result = await response.json();
      console.log("Publish result:", result);

      toast({
        title: "Model Published!",
        description: `${modelToPublish.name} is now available in the community`,
      });

      // Refresh published models list
      fetchMyPublishedModels();

      setPublishDialogOpen(false);
      setModelToPublish(null);
      // Reset form
      setPublishForm({
        description: "",
        price: 0,
        license_type: "personal_use",
        category: "",
        tags: "",
        model_type: "",
        framework: "",
      });
    } catch (error) {
      console.error("Publish error:", error);
      toast({
        title: "Publish Failed",
        description: "Failed to publish model",
        variant: "destructive",
      });
    }
  };

  const handleUnpublish = async (publishedModelId: number, modelName: string) => {
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Error",
          description: "You must be logged in",
          variant: "destructive",
        });
        return;
      }

      const response = await fetch(`http://localhost:8081/v1/published-models/${publishedModelId}/unpublish`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || "Failed to unpublish model");
      }

      toast({
        title: "Model Unpublished",
        description: `${modelName} has been removed from the community`,
      });

      // Refresh published models list
      fetchMyPublishedModels();
    } catch (error) {
      console.error("Unpublish error:", error);
      toast({
        title: "Unpublish Failed",
        description: "Failed to unpublish model",
        variant: "destructive",
      });
    }
  };

  // Check if a model is already published
  const isModelPublished = (modelId: number) => {
    return myPublishedModels.some(
      (publishedModel) => publishedModel.model_id === modelId && publishedModel.is_active
    );
  };

  return (
    <div className="space-y-6 animate-slide-up">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">AI Models</h2>
          <p className="text-muted-foreground mt-1">
            Manage and publish your AI models
          </p>
        </div>

        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button className="bg-gradient-primary hover:opacity-90 shadow-glow">
              <Plus className="w-4 h-4 mr-2" />
              Add Model
            </Button>
          </DialogTrigger>

          <DialogContent className="bg-card border-border max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle className="text-2xl">Add new Model</DialogTitle>
              <DialogDescription>
                Configure a new AI model for training
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              {/* Training Mode Selection */}
              <div className="space-y-3">
                <Label>Training Location</Label>
                <RadioGroup value={trainingMode} onValueChange={(value) => setTrainingMode(value as "local" | "server")}>
                  <div className="flex items-start space-x-2 p-3 rounded-lg border border-border hover:border-primary/50 transition-colors">
                    <RadioGroupItem value="local" id="local" />
                    <div className="flex-1">
                      <Label htmlFor="local" className="flex items-center gap-2 cursor-pointer">
                        <FolderOpen className="w-4 h-4" />
                        <span className="font-semibold">Train on My Machine</span>
                        <Badge variant="secondary" className="ml-auto">Free</Badge>
                      </Label>
                      <p className="text-xs text-muted-foreground mt-1">
                        Use your own hardware. Requires training agent to be running.
                      </p>
                    </div>
                  </div>

                  <div className="flex items-start space-x-2 p-3 rounded-lg border border-border hover:border-primary/50 transition-colors">
                    <RadioGroupItem
                      value="server"
                      id="server"
                      disabled={!subscriptionContext?.canTrainOnServer}
                    />
                    <div className="flex-1">
                      <Label htmlFor="server" className={`flex items-center gap-2 ${subscriptionContext?.canTrainOnServer ? 'cursor-pointer' : 'cursor-not-allowed opacity-50'}`}>
                        <Cloud className="w-4 h-4" />
                        <span className="font-semibold">Train on Server</span>
                        <Badge className="ml-auto bg-gradient-to-r from-blue-500 to-purple-500">Paid</Badge>
                      </Label>
                      <p className="text-xs text-muted-foreground mt-1">
                        Upload to our servers. Powerful GPUs, no setup required.
                      </p>
                    </div>
                  </div>
                </RadioGroup>

                {/* Upgrade prompt for free users trying to use server */}
                {!subscriptionContext?.canTrainOnServer && trainingMode === "server" && (
                  <Alert>
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription className="text-sm">
                      Server training requires a paid subscription.{" "}
                      <Button
                        variant="link"
                        className="p-0 h-auto font-semibold text-primary"
                        onClick={() => navigate("/pricing")}
                      >
                        View Plans
                      </Button>
                    </AlertDescription>
                  </Alert>
                )}

                {/* Agent connection status for local training */}
                {trainingMode === "local" && (
                  <Alert className={subscriptionContext?.isAgentConnected ? "border-green-500" : "border-yellow-500"}>
                    <AlertCircle className={`h-4 w-4 ${subscriptionContext?.isAgentConnected ? 'text-green-500' : 'text-yellow-500'}`} />
                    <AlertDescription className="text-sm">
                      {subscriptionContext?.isAgentConnected ? (
                        <span className="text-green-600 dark:text-green-400">Agent connected and ready!</span>
                      ) : (
                        <>
                          Agent not connected.{" "}
                          <Button
                            variant="link"
                            className="p-0 h-auto font-semibold"
                            onClick={() => navigate("/settings")}
                          >
                            Set up agent
                          </Button>
                        </>
                      )}
                    </AlertDescription>
                  </Alert>
                )}
              </div>

              {/* Model Name - Always shown */}
              <div className="space-y-2">
                <Label htmlFor="name">Model Name</Label>
                <Input
                  id="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g., GPT-4 Turbo"
                />
              </div>

              {/* Profile Image - Always shown */}
              <div className="space-y-2">
                <Label htmlFor="picture">Profile Image</Label>
                <Input
                  id="picture"
                  type="file"
                  onChange={(e) => setPicture(e.target.files?.[0] || null)}
                />
              </div>

              {/* Local Training Fields */}
              {trainingMode === "local" && (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="localPath">Local Model Path</Label>
                    <Input
                      id="localPath"
                      value={folderPath}
                      onChange={(e) => setFolderPath(e.target.value)}
                      placeholder="/home/user/my-models/my-model"
                    />
                    <p className="text-xs text-muted-foreground">
                      Absolute path to your model folder on your machine
                    </p>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="trainingScript">Training Script Path</Label>
                    <Input
                      id="trainingScript"
                      value={trainingScript}
                      onChange={(e) => setTrainingScript(e.target.value)}
                      placeholder="e.g., train.py or scripts/train.py"
                    />
                    <p className="text-xs text-muted-foreground">
                      Path to training script relative to the model folder
                    </p>
                  </div>
                </>
              )}

              {/* Server Training Fields */}
              {trainingMode === "server" && (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="folder">Model Files (.zip or multiple files)</Label>
                    <Input
                      id="folder"
                      type="file"
                      multiple
                      onChange={(e) => setFolder(e.target.files ? Array.from(e.target.files) : null)}
                    />
                    <p className="text-xs text-muted-foreground">
                      Upload your model files to our server
                    </p>
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
                      Path to training script relative to uploaded files
                    </p>
                  </div>

                  {subscriptionContext?.subscription && (
                    <Alert>
                      <AlertDescription className="text-sm">
                        <span className="font-semibold">
                          {subscriptionContext.subscription.training_credits} training credits
                        </span> remaining this month
                      </AlertDescription>
                    </Alert>
                  )}
                </>
              )}
            </div>

            <Button
              className="w-full bg-gradient-primary hover:opacity-90"
              onClick={handleSubmit}
              disabled={loading || (trainingMode === "server" && !subscriptionContext?.canTrainOnServer)}
            >
              {loading ? "Processing..." : trainingMode === "local" ? "Add Local Model" : "Upload & Add Model"}
            </Button>
          </DialogContent>
        </Dialog>
      </div>

      {/* Agent Status Banner */}
      {subscriptionContext?.isAgentConnected ? (
        <Alert className="border-green-500/50 bg-green-500/10">
          <AlertCircle className="h-4 w-4 text-green-500" />
          <AlertDescription className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <span className="text-sm">
                <span className="font-semibold text-green-500">Training Agent Connected</span>
                {subscriptionContext.agentSystemInfo && (
                  <span className="text-muted-foreground ml-2">
                    • {subscriptionContext.agentSystemInfo.platform}
                    • {subscriptionContext.agentSystemInfo.cuda_available ? `${subscriptionContext.agentSystemInfo.gpu_count}x GPU` : 'CPU'}
                  </span>
                )}
              </span>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigate("/settings")}
              className="ml-4"
            >
              View Details
            </Button>
          </AlertDescription>
        </Alert>
      ) : (
        <Alert className="border-yellow-500/50 bg-yellow-500/10">
          <AlertCircle className="h-4 w-4 text-yellow-500" />
          <AlertDescription className="flex items-center justify-between">
            <span className="text-sm">
              <span className="font-semibold text-yellow-500">Training Agent Not Connected</span>
              <span className="text-muted-foreground ml-2">
                Connect your training agent to train models on your machine
              </span>
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigate("/settings")}
              className="ml-4"
            >
              Setup Agent
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* Tabs for My Models and Published Models */}
      <Tabs defaultValue="my-models" className="w-full">
        <TabsList className="grid w-full max-w-md grid-cols-2">
          <TabsTrigger value="my-models">My Models</TabsTrigger>
          <TabsTrigger value="published">Published Models</TabsTrigger>
        </TabsList>

        {/* My Models Tab */}
        <TabsContent value="my-models" className="mt-6">
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
                  {model.trained_model_path && !isModelPublished(model.id) && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-secondary hover:text-secondary hover:bg-secondary/10"
                      onClick={() => handlePublishClick(model)}
                      title="Publish to Community"
                    >
                      <Upload className="w-4 h-4" />
                    </Button>
                  )}
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
        </TabsContent>

        {/* Published Models Tab */}
        <TabsContent value="published" className="mt-6">
          {myPublishedModels.length === 0 ? (
            <Card className="bg-gradient-card border-border shadow-card">
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Store className="w-16 h-16 text-muted-foreground mb-4" />
                <h3 className="text-xl font-semibold mb-2">No Published Models Yet</h3>
                <p className="text-muted-foreground text-center max-w-md">
                  Train a model and click the Upload button to publish it to the community marketplace.
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {myPublishedModels.map((model) => (
                <Card
                  key={model.id}
                  className="bg-gradient-card border-border hover:border-primary/50 transition-all shadow-card hover:shadow-glow group"
                >
                  <CardHeader>
                    <div className="flex items-start justify-between mb-3">
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
                        <Badge
                          className={`${
                            model.price === 0
                              ? "bg-green-500/20 text-green-500 border-green-500/30"
                              : "bg-primary/20 text-primary border-primary/30"
                          }`}
                        >
                          {model.price === 0 ? "Free" : `$${(model.price / 100).toFixed(2)}`}
                        </Badge>
                        {model.is_active && (
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                            onClick={() => handleUnpublish(model.id, model.name)}
                            title="Unpublish Model"
                          >
                            <XCircle className="w-4 h-4" />
                          </Button>
                        )}
                      </div>
                    </div>
                    <CardTitle>{model.name}</CardTitle>
                    <CardDescription className="line-clamp-2">
                      {model.description}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-4">
                        <div className="flex items-center gap-1">
                          <Star className="w-4 h-4 text-yellow-500" />
                          <span>{model.rating_average?.toFixed(1) || "0.0"}</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <Download className="w-4 h-4" />
                          <span>{model.downloads_count || 0}</span>
                        </div>
                      </div>
                      <Badge
                        variant={model.is_active ? "default" : "secondary"}
                        className="text-xs"
                      >
                        {model.is_active ? "Active" : "Inactive"}
                      </Badge>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>

      {/* Publish Model Dialog */}
      <Dialog open={publishDialogOpen} onOpenChange={setPublishDialogOpen}>
        <DialogContent className="bg-card border-border max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="text-2xl">Publish to Community</DialogTitle>
            <DialogDescription>
              Share your trained model "{modelToPublish?.name}" with the community
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* Description */}
            <div className="space-y-2">
              <Label htmlFor="description">Description *</Label>
              <textarea
                id="description"
                className="w-full min-h-[120px] px-3 py-2 rounded-md border border-border bg-background text-foreground resize-none"
                placeholder="Describe your model, its use case, training data, and performance..."
                value={publishForm.description}
                onChange={(e) => setPublishForm({ ...publishForm, description: e.target.value })}
                required
              />
            </div>

            {/* Price */}
            <div className="space-y-2">
              <Label htmlFor="price">Price * (USD)</Label>
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground">$</span>
                <Input
                  id="price"
                  type="number"
                  min="0"
                  step="0.01"
                  placeholder="0.00"
                  value={publishForm.price / 100}
                  onChange={(e) => setPublishForm({ ...publishForm, price: Math.round(parseFloat(e.target.value || "0") * 100) })}
                />
                <span className="text-xs text-muted-foreground whitespace-nowrap">
                  (0 = Free)
                </span>
              </div>
            </div>

            {/* License Type */}
            <div className="space-y-2">
              <Label htmlFor="license_type">License Type *</Label>
              <select
                id="license_type"
                className="w-full px-3 py-2 rounded-md border border-border bg-background text-foreground"
                value={publishForm.license_type}
                onChange={(e) => setPublishForm({ ...publishForm, license_type: e.target.value })}
              >
                <option value="personal_use">Personal Use Only</option>
                <option value="commercial">Commercial License</option>
                <option value="mit">MIT License</option>
                <option value="apache">Apache 2.0</option>
                <option value="gpl">GPL</option>
              </select>
            </div>

            {/* Optional Fields */}
            <div className="border-t border-border pt-4 space-y-4">
              <h4 className="font-semibold text-sm text-muted-foreground">Optional Information</h4>

              {/* Category */}
              <div className="space-y-2">
                <Label htmlFor="category">Category</Label>
                <Input
                  id="category"
                  placeholder="e.g., Image Classification, NLP, Computer Vision"
                  value={publishForm.category}
                  onChange={(e) => setPublishForm({ ...publishForm, category: e.target.value })}
                />
              </div>

              {/* Tags */}
              <div className="space-y-2">
                <Label htmlFor="tags">Tags (comma-separated)</Label>
                <Input
                  id="tags"
                  placeholder="e.g., pytorch, cnn, computer-vision"
                  value={publishForm.tags}
                  onChange={(e) => setPublishForm({ ...publishForm, tags: e.target.value })}
                />
              </div>

              {/* Model Type */}
              <div className="space-y-2">
                <Label htmlFor="model_type">Model Type</Label>
                <Input
                  id="model_type"
                  placeholder="e.g., ResNet50, Transformer, CNN"
                  value={publishForm.model_type}
                  onChange={(e) => setPublishForm({ ...publishForm, model_type: e.target.value })}
                />
              </div>

              {/* Framework */}
              <div className="space-y-2">
                <Label htmlFor="framework">Framework</Label>
                <select
                  id="framework"
                  className="w-full px-3 py-2 rounded-md border border-border bg-background text-foreground"
                  value={publishForm.framework}
                  onChange={(e) => setPublishForm({ ...publishForm, framework: e.target.value })}
                >
                  <option value="">Select Framework</option>
                  <option value="pytorch">PyTorch</option>
                  <option value="tensorflow">TensorFlow</option>
                  <option value="keras">Keras</option>
                  <option value="scikit-learn">scikit-learn</option>
                  <option value="jax">JAX</option>
                </select>
              </div>

            </div>
          </div>

          <div className="flex gap-3 justify-end">
            <Button
              variant="outline"
              onClick={() => setPublishDialogOpen(false)}
            >
              Cancel
            </Button>
            <Button
              className="bg-gradient-primary hover:opacity-90"
              onClick={handlePublishSubmit}
              disabled={!publishForm.description || !publishForm.license_type}
            >
              Publish to Community
            </Button>
          </div>
        </DialogContent>
      </Dialog>

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
