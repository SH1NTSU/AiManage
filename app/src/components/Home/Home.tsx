import "./Home.scss";
import Models from "../Models/Models.tsx";
import { FiPlus } from "react-icons/fi";
import { useState } from "react";
import Form from "../Form/Form.tsx"



const Home: any = () => {
	const [display, setDisplay] = useState(false);
	
	




	return (
		<>
		<main>
		<h1>Your Models</h1>
		<Models></Models>
		<button onClick={() => { setDisplay(true)}} id="plus"><FiPlus /></button>
		</main>	
		<Form display={display}   onClose={() => setDisplay(false)}></Form>

		</>
	)

}

export default Home;
