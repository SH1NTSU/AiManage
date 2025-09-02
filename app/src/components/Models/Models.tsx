import "./Models.scss";

const Models: any = () => {
	

	return (
		<div className="Models">
			<div className="Model">
				<img src="{img}" alt="photo"/>
				<span id="name">Pokemon Model</span>
				<div>
				<button className="fancy-btn model-btn">train</button>
				<button className="fancy-btn model-btn" >delete</button>
				</div>
			</div>
		</div>
	)
}


export default Models;
