import axios from "axios";
import { createContext, type ReactNode, useEffect, useState } from "react";
interface Model {
	name: string;
	picture: string; 
	folder: string[];
}

interface ModelContextType {
  name: string;
  picture: File | null;          // store actual File
  folder: File[] | null;
  setName: (name: string) => void;
  setPicture: (file: File | null) => void;
  setFolder: (files: File[] | null) => void;
  send: () => Promise<void>;
  models: Model[];
}



export const ModelContext = createContext<ModelContextType | null>(null);

export const ModelProvider = ({ children }: { children: ReactNode }) => {
  const [name, setName] = useState<string>("");
  const [picture, setPicture] = useState<File | null>(null);
  const [folder, setFolder] = useState<File[] | null>(null);

  const [models, setModels] = useState<Model[]>([]);
const send = async () => {
  try {
    const formData = new FormData();
    formData.append("name", name);

    if (picture) {
      formData.append("picture", picture);
    }

    if (folder) {
      folder.forEach((file) => formData.append("folder", file));
    }

    const res = await axios.post(
      "http://localhost:8080/api/v1/insert",
      formData,
      { headers: { "Content-Type": "multipart/form-data" } }
    );

    console.log("Upload successful:", res.data);
  } catch (err) {
    console.error("Upload failed:", err);
  }
};

useEffect(() => {
  const socket = new WebSocket("ws://localhost:8080/ws");
   

  socket.onmessage = (event) => {
    const updatedModels: Model[] = JSON.parse(event.data);
    setModels(updatedModels);
  };

  return () => socket.close();
}, []);




  return (
    <ModelContext.Provider
      value={{ name, picture, folder, setName, setPicture, setFolder, send ,models}}
    >
      {children}
    </ModelContext.Provider>
  );
};
