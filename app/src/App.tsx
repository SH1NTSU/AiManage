import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Layout } from "./components/Layout";
import Models from "./pages/Models";
import Statistics from "./pages/Statistics";
import Settings from "./pages/Settings";
import Auth from "./pages/Auth";
import NotFound from "./pages/NotFound";
import { AuthProvider } from "./context/authContext";
import { ModelProvider } from "./context/modelContext";
import { TrainingProvider } from "./context/trainingContext";
import ProtectedRoute from "./components/ProtectedRoute";

const queryClient = new QueryClient();


const App = () => (
    <BrowserRouter>
	<AuthProvider>
	  <ModelProvider>
	    <TrainingProvider>
	      <QueryClientProvider client={queryClient}>
	    <TooltipProvider>
	      <Toaster />
	      <Sonner />
		<Routes>
		  {/* Public */}
		  <Route path="/auth" element={<Auth />} />

		  {/* Protected */}
		  <Route
		    path="/"
		    element={
		      <ProtectedRoute>
			<Layout><Models /></Layout>
		      </ProtectedRoute>
		    }
		  />

		  <Route
		    path="/statistics"
		    element={
		      <ProtectedRoute>
			<Layout><Statistics /></Layout>
		      </ProtectedRoute>
		    }
		  />

		  <Route
		    path="/settings"
		    element={
		      <ProtectedRoute>
			<Layout><Settings /></Layout>
		      </ProtectedRoute>
		    }
		  />

		  <Route path="*" element={<NotFound />} />
		</Routes>
	    </TooltipProvider>
	      </QueryClientProvider>
	    </TrainingProvider>
	  </ModelProvider>
	  </AuthProvider>
    </BrowserRouter>
);

export default App;
