import { useContext } from "react";
import { ModelContext } from "../../context/modelContext.tsx";

interface FormProps {
  display: boolean;
  onClose: () => void;
}

const Form = ({ display, onClose }: FormProps) => {
  if (!display) return null;

  const ctx = useContext(ModelContext);
  if (!ctx) throw new Error("Form must be used within a ModelProvider");

  const { setName, setPicture, setFolder, send } = ctx;

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setName(e.target.value);
  };

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setPicture(e.target.files[0]);
    }
  };

  const handleFolderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setFolder(Array.from(e.target.files));
    }
  };

  const handleSubmit = async () => {
    await send();
    onClose();
  };

  return (
    <div 
      className="fixed inset-0 bg-black bg-opacity-60 flex items-center justify-center z-50"
      onClick={onClose}
    >
      <main 
        className="bg-white text-black p-8 rounded-lg shadow-xl w-96 max-w-sm mx-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-xl font-semibold text-gray-800 mb-6 text-center">
          Add new Model
        </h2>
        
        <div className="bg-gray-100 p-3 rounded-md font-mono text-sm mb-4 text-center">
          D:\S\C\MODE
        </div>
        
        <p className="text-gray-600 text-sm mb-6 text-center">
          % Priority network
        </p>

        <form className="space-y-4 w-full">
          <input
            type="text"
            placeholder="Name"
            value={ctx.name}
            onChange={handleNameChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Profile Image
            </label>
            <input 
              type="file" 
              onChange={handleImageChange}
              className="w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Model Folder
            </label>
            <input
              type="file"
              onChange={handleFolderChange}
              webkitdirectory="true"
              multiple
              className="w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-green-50 file:text-green-700 hover:file:bg-green-100"
            />
          </div>
        </form>

        <button
          onClick={handleSubmit}
          className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 mt-6"
        >
          Submit
        </button>
      </main>
    </div>
  );
};

export default Form;
