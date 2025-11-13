import axios from "axios";
import { createContext, type ReactNode, useEffect, useState } from "react";

interface Model {
  id: number;
  user_id: number;
  name: string;
  picture: string;
  folder: string[];
  training_script?: string; // Path to training script (e.g., "train.py" or "PokemonModel/train.py")
  trained_model_path?: string; // Path to trained model file (e.g., "ModelName/best_model.pth")
  trained_at?: string; // Timestamp when model was trained
  created_at?: string;
  updated_at?: string;
}

interface ModelContextType {
  name: string;
  picture: File | null;
  folder: File[] | null;
  folderPath: string;
  trainingScript: string;
  setName: (name: string) => void;
  setPicture: (file: File | null) => void;
  setFolder: (files: File[] | null) => void;
  setFolderPath: (path: string) => void;
  setTrainingScript: (script: string) => void;
  send: () => Promise<void>;
  deleteModel: (modelId: number, modelName: string) => Promise<void>;
  models: Model[];
  loading: boolean;
  error: string | null;
}

export const ModelContext = createContext<ModelContextType | null>(null);

export const ModelProvider = ({ children }: { children: ReactNode }) => {
  const [name, setName] = useState<string>("");
  const [picture, setPicture] = useState<File | null>(null);
  const [folder, setFolder] = useState<File[] | null>(null);
  const [folderPath, setFolderPath] = useState<string>("");
  const [trainingScript, setTrainingScript] = useState<string>("train.py");
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const send = async () => {
    setLoading(true);
    setError(null);

    try {
      const formData = new FormData();
      formData.append("name", name);
      formData.append("training_script", trainingScript);

      // For local mode: send folder path instead of files
      if (folderPath) {
        formData.append("folder_path", folderPath);
        console.log("Sending form data (local mode) with:", {
          name,
          picture: picture?.name,
          folder_path: folderPath,
          training_script: trainingScript
        });
      } else {
        // For server mode: send files
        if (folder) {
          folder.forEach((file) => formData.append("folder", file));
        }
        console.log("Sending form data (server mode) with:", {
          name,
          picture: picture?.name,
          folder: folder?.map(f => f.name),
          training_script: trainingScript
        });
      }

      if (picture) {
        formData.append("picture", picture);
      }

      const res = await axios.post(
        "http://localhost:8081/v1/insert",
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data"
          },
          timeout: 30000 // 30 second timeout
        }
      );

      console.log("Upload successful:", res.data);

      // Reset form after successful upload
      setName("");
      setPicture(null);
      setFolder(null);
      setFolderPath("");
      setTrainingScript("train.py");

    } catch (err: any) {
      console.error("Upload failed:", err);

      if (err.response) {
        // Server responded with error status
        setError(`Server error: ${err.response.status} - ${err.response.data}`);
      } else if (err.request) {
        // Request was made but no response received
        setError("Network error - no response from server");
      } else {
        // Something else happened
        setError(`Error: ${err.message}`);
      }
    } finally {
      setLoading(false);
    }
  };

  const deleteModel = async (modelId: number, modelName: string) => {
    setLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem("token");
      if (!token) {
        setError("No authentication token found");
        return;
      }

      console.log(`Deleting model: ${modelName} (ID: ${modelId})`);

      const res = await axios.delete(
        "http://localhost:8081/v1/deleteModel",
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json"
          },
          data: {
            model_id: modelId,
            name: modelName
          },
          timeout: 10000 // 10 second timeout
        }
      );

      console.log("Delete successful:", res.data);

      // Optimistically remove the model from local state
      setModels(prevModels => prevModels.filter(model => model.id !== modelId));

    } catch (err: any) {
      console.error("Delete failed:", err);

      if (err.response) {
        setError(`Server error: ${err.response.status} - ${err.response.data}`);
      } else if (err.request) {
        setError("Network error - no response from server");
      } else {
        setError(`Error: ${err.message}`);
      }
    } finally {
      setLoading(false);
    }
  };

  // Fetch initial models from authenticated endpoint
  useEffect(() => {
    const fetchModels = async () => {
      setLoading(true);
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          console.log("No token found, skipping model fetch");
          setModels([]);
          return;
        }

        const res = await axios.get("http://localhost:8081/v1/getModels", {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });

        console.log("Fetched user models:", res.data);
        setModels(res.data || []);
        setError(null);
      } catch (err: any) {
        console.error("Error fetching models:", err);
        if (err.response?.status === 401) {
          setError("Unauthorized - please login");
        } else {
          setError("Failed to fetch models");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchModels();
  }, []);

  // WebSocket for real-time updates with authentication
  useEffect(() => {
    let socket: WebSocket;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 10;

    const connectWebSocket = () => {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          console.log("No token found, skipping WebSocket connection");
          return;
        }

        // Include JWT token as query parameter
        const wsUrl = `ws://localhost:8081/v1/ws?token=${encodeURIComponent(token)}`;
        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
          console.log("WebSocket connected (authenticated)");
          reconnectAttempts = 0;
        };

        socket.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);

            // Check if it's a typed message (agent_status, training_update, etc.)
            if (message.type) {
              console.log(`📥 [WebSocket] Received ${message.type}:`, message.data);

              if (message.type === "training_update") {
                // Training status update
                const { training_id, status, message: msg, error_message } = message.data;
                console.log(`🎯 Training ${training_id}: ${status}`);
                if (msg) console.log(`   Message: ${msg}`);
                if (error_message) console.error(`   Error: ${error_message}`);

                // TODO: Update UI with training status
                // You could show a toast notification or update a training progress component
              } else if (message.type === "training_output") {
                // Live training output
                const { training_id, output } = message.data;
                console.log(`📝 [Training ${training_id}] ${output}`);

                // TODO: Stream output to a log viewer in the UI
              } else if (message.type === "agent_status") {
                // Agent status updates are handled by subscriptionContext
                console.log("📡 Agent status update (handled by subscriptionContext)");
              }
            } else {
              // Legacy format: assume it's a model update array
              const updatedModels: Model[] = message;
              console.log("📥 [WebSocket] Received user-specific models update:", updatedModels);

              // DEBUG: Log trained model paths
              updatedModels.forEach(model => {
                if (model.trained_model_path) {
                  console.log(`  ✅ Model "${model.name}" has trained_model_path: ${model.trained_model_path}`);
                } else {
                  console.log(`  ⚠️  Model "${model.name}" has NO trained_model_path`);
                }
              });

              setModels(updatedModels);
            }
          } catch (error) {
            console.error("Error parsing WebSocket message:", error);
          }
        };

        socket.onclose = (event) => {
          console.log("WebSocket disconnected:", event.code, event.reason);
          if (reconnectAttempts < maxReconnectAttempts) {
            setTimeout(() => {
              reconnectAttempts++;
              console.log(`Reconnecting WebSocket (attempt ${reconnectAttempts})`);
              connectWebSocket();
            }, 1000 * reconnectAttempts);
          }
        };

        socket.onerror = (error) => {
          console.error("WebSocket error:", error);
        };

      } catch (error) {
        console.error("WebSocket connection failed:", error);
      }
    };

    connectWebSocket();

    return () => {
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.close();
      }
    };
  }, []);

  return (
    <ModelContext.Provider
      value={{
        name,
        picture,
        folder,
        folderPath,
        trainingScript,
        setName,
        setPicture,
        setFolder,
        setFolderPath,
        setTrainingScript,
        send,
        deleteModel,
        models,
        loading,
        error
      }}
    >
      {children}
    </ModelContext.Provider>
  );
};
