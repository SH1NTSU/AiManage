import axios from "axios";
import { createContext, type ReactNode, useEffect, useState, useContext } from "react";
import { AuthContext } from "./authContext";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8081";

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
  const authContext = useContext(AuthContext);
  const token = authContext?.token || null;
  
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
      } else {
        // For server mode: send files
        if (folder) {
          folder.forEach((file) => formData.append("folder", file));
        }
      }

      if (picture) {
        formData.append("picture", picture);
      }

      const res = await axios.post(
        `${API_URL}/v1/insert`,
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data"
          },
          timeout: 30000 // 30 second timeout
        }
      );


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


      const res = await axios.delete(
        `${API_URL}/v1/deleteModel`,
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

  // Fetch models when token changes (login/logout)
  useEffect(() => {
    const fetchModels = async () => {
      // Clear models immediately if no token (logout)
      if (!token) {
        setModels([]);
        setError(null);
        return;
      }

      setLoading(true);
      try {
        const res = await axios.get(`${API_URL}/v1/getModels`, {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });

        setModels(res.data || []);
        setError(null);
      } catch (err: any) {
        console.error("Error fetching models:", err);
        if (err.response?.status === 401) {
          setError("Unauthorized - please login");
          setModels([]); // Clear models on unauthorized
        } else {
          setError("Failed to fetch models");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchModels();
  }, [token]); // Re-fetch when token changes

  // WebSocket for real-time updates with authentication
  useEffect(() => {
    // Don't connect if no token
    if (!token) {
      return;
    }

    let socket: WebSocket;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 10;

    const connectWebSocket = () => {
      try {
        // Use token from context (already checked above)
        // Include JWT token as query parameter
        const wsProtocol = API_URL.startsWith('https') ? 'wss' : 'ws';
        const wsHost = API_URL.replace(/^https?:\/\//, '');
        const wsUrl = `${wsProtocol}://${wsHost}/v1/ws?token=${encodeURIComponent(token)}`;
        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
          reconnectAttempts = 0;
        };

        socket.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);

            // Check if it's a typed message (agent_status, training_update, etc.)
            if (message.type) {
              if (message.type === "training_update") {
                // Training status update
                const { training_id, status, message: msg, error_message } = message.data;
                if (error_message) console.error(`Training ${training_id} error: ${error_message}`);

                // TODO: Update UI with training status
                // You could show a toast notification or update a training progress component
              } else if (message.type === "training_output") {
                // Live training output
                // TODO: Stream output to a log viewer in the UI
              } else if (message.type === "agent_status") {
                // Agent status updates are handled by subscriptionContext
              }
            } else {
              // Legacy format: assume it's a model update array
              const updatedModels: Model[] = message;
              setModels(updatedModels);
            }
          } catch (error) {
            console.error("Error parsing WebSocket message:", error);
          }
        };

        socket.onclose = (event) => {
          if (reconnectAttempts < maxReconnectAttempts) {
            setTimeout(() => {
              reconnectAttempts++;
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
  }, [token]); // Reconnect when token changes

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
