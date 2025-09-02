import { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./Nav.scss";
import { MdDarkMode } from "react-icons/md";
import { MdLightMode } from "react-icons/md";





const Nav = () => {
	const navigate = useNavigate();
	const [selected, setSelected] = useState("home");

	

		
	const handleChange = (value: string) => {
	  setSelected(value);
	  const routes: Record<string, string> = {
	    home: "/",
	    stats: "/stats",
	    settings: "/settings",
	  };
	  navigate(routes[value]);
	};


	const handleDarkTheme = () => {
	  document.documentElement.style.setProperty("--bg-color", "#212121");
	  document.documentElement.style.setProperty("--text-color", "#ffffff");
	};

	const handleLightTheme = () => {
	  document.documentElement.style.setProperty("--bg-color", "whitesmoke");
	  document.documentElement.style.setProperty("--text-color", "#000000");
	};
	  return (
	    <section>
	    <div className="radio-input">
		<label className="label">
		  <input
		    type="radio"
		    name="nav"
		    value="home"
		    checked={selected === "home"}
		    onChange={() => handleChange("home")}
		  />
		  <span className="text">Home</span>
		</label>

		<label className="label">
		  <input
		    type="radio"
		    name="nav"
		    value="stats"
		    checked={selected === "stats"}
		    onChange={() => handleChange("stats")}
		  />
		  <span className="text">Stats</span>
		</label>

		<label className="label">
		  <input
		    type="radio"
		    name="nav"
		    value="settings"
		    checked={selected === "settings"}
		    onChange={() => handleChange("settings")}
		  />
		  <span className="text">Settings</span>
		</label>
	      </div>

	      <div>
		<button id="dark" onClick={handleDarkTheme}><span><MdDarkMode/></span></button>

		<button id="light"onClick={handleLightTheme}><span>  <MdLightMode /> </span></button>
	      </div>
	    </section>
	  );
};

export default Nav;
