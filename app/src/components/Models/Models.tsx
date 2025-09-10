import { useContext } from "react";
import "./Models.scss";
import { ModelContext } from "../../context/modelContext";

const Models: any = () => {
	const {models} = useContext(ModelContext)!;


	if (!models) return <div>Loading model...</div>

	return (
	  <div className="Models">
	    {models.map((model) => (
	      <div key={model.name} className="Model">
		<img src={model.image} alt="photo" />
		<span id="name">{model.name}</span>
		<div>
		  <button className="fancy-btn model-btn">train</button>
		  <button className="fancy-btn model-btn">delete</button>
		</div>
	      </div>
	    ))}
	  </div>
	);
}


export default Models;
