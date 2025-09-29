import { useContext, useState } from "react";
import { AuthContext } from "../context/authContext";
import "./AuthForm.scss";

export default function Register() {
  const auth = useContext(AuthContext);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (auth) await auth.register(email, password);
  };

  return (
    <form className="auth-form" onSubmit={handleSubmit}>
      <h2>Register</h2>
      <input
        type="email"
        placeholder="Email"
        value={email}
        autoComplete="username"
        onChange={e => setEmail(e.target.value)}
        required
      />
      <input
        type="password"
        placeholder="Password"
        value={password}
        autoComplete="new-password"
        onChange={e => setPassword(e.target.value)}
        required
      />
      <button type="submit" disabled={auth?.loading}>Register</button>
      {auth?.error && <div className="error">{auth.error}</div>}
    </form>
  );
}