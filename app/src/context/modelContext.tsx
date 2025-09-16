import axios from "axios";
import { createContext, type ReactNode, useEffect, useState } from "react";

interface Model {
  _id?: string; // Add this if your backend returns MongoDB IDs
  name: string;
  picture: string; 
  folder: string[];
}

interface ModelContextType {
  name: string;
  picture: File | null;
  folder: File[] | null;
  setName: (name: string) => void;
  setPicture: (file: File | null) => void;
  setFolder: (files: File[] | null) => void;
  send: () => Promise<void>;
  models: Model[];
  loading: boolean;
  error: string | null;
}

export const ModelContext = createContext<ModelContextType | null>(null);

export const ModelProvider = ({ children }: { children: ReactNode }) => {
  const [name, setName] = useState<string>("");
  const [picture, setPicture] = useState<File | null>(null);
  const [folder, setFolder] = useState<File[] | null>(null);
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const send = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const formData = new FormData();
      formData.append("name", name);

      if (picture) {
        formData.append("picture", picture);
      }

      if (folder) {
        folder.forEach((file) => formData.append("folder", file));
      }

      console.log("Sending form data with:", {
        name,
        picture: picture?.name,
        folder: folder?.map(f => f.name)
      });

      const res = await axios.post(
        "http://localhost:8080/api/v1/insert",
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

  useEffect(() => {
    let socket: WebSocket;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 10;

    const connectWebSocket = () => {
      try {
        socket = new WebSocket("ws://localhost:8080/ws");
        
        socket.onopen = () => {
          console.log("WebSocket connected");
          reconnectAttempts = 0;
          setError(null);
        };

        socket.onmessage = (event) => {
          try {
            const updatedModels: Model[] = JSON.parse(event.data);
            console.log("Received models update:", updatedModels);
            setModels(updatedModels);
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
          } else {
            setError("WebSocket connection failed after multiple attempts");
          }
        };

        socket.onerror = (error) => {
          console.error("WebSocket error:", error);
          setError("WebSocket connection error");
        };

      } catch (error) {
        console.error("WebSocket connection failed:", error);
        setError("Failed to create WebSocket connection");
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
        setName, 
        setPicture, 
        setFolder, 
        send, 
        models,
        loading,
        error
      }}
    >
      {children}
    </ModelContext.Provider>
  );
};
