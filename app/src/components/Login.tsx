import { useContext, useState } from "react";
import { AuthContext } from "../context/authContext";
import "./AuthForm.scss";

export default function Login() {
  const auth = useContext(AuthContext);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (auth) await auth.login(email, password);
  };

  return (
    <form className="auth-form" onSubmit={handleSubmit}>
      <h2>Login</h2>
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
        autoComplete="current-password"
        onChange={e => setPassword(e.target.value)}
        required
      />
      <button type="submit" disabled={auth?.loading}>Login</button>
      {auth?.error && <div className="error">{auth.error}</div>}
    </form>
  );
}