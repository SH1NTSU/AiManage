import { useContext } from "react";
import { Navigate } from "react-router-dom";
import { AuthContext } from "@/context/authContext";
const ProtectedRoute = ({ children }: { children: JSX.Element }) => {
  const auth = useContext(AuthContext);

  if (!auth?.token) {
    return <Navigate to="/auth" replace />;
  }

  return children;
};


export default ProtectedRoute;
