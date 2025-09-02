import { useState } from "react";
import "./Form.scss";

interface FormProps {
  display: boolean;           // whether modal is visible
  onClose: () => void;        // callback to close modal
}

const Form = ({ display, onClose }: FormProps) => {
  if (!display) return null; // don't render if not visible



  const [imageFile, setImageFile] = useState<File | null>(null);
  const [folderFiles, setFolderFiles] = useState<FileList | null>(null);
  const [name, setName] = useState<string>("");
  

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
	if (e.target.value) {
		setName(e.target.value);
	}
  }

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setImageFile(e.target.files[0]);
    }
  };

  const handleFolderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setFolderFiles(e.target.files);
    }
  };






  return (
    <div className="overlay" onClick={onClose}>
      <main className="window" onClick={(e) => e.stopPropagation()}>
        <h2>Add new Model</h2>
	<form>
	<input type="text" placeholder="Name" className="input model-input" onChange={handleNameChange}/>
	<input type="file" onChange={handleImageChange}/>
	<input type="file" onChange={handleFolderChange} webkitdirectory multiple/>
	
	</form>	
        <button onClick={onClose}>Submit</button>
      </main>
    </div>
  );
};

export default Form;
