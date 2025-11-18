import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { ThemeProvider } from "@/components/theme-provider";
import { GoogleOAuthProvider } from "@react-oauth/google";
import { Layout } from "./components/Layout";
import Models from "./pages/Models";
import Community from "./pages/Community";
import ModelDetail from "./pages/ModelDetail";
import Statistics from "./pages/Statistics";
import Settings from "./pages/Settings";
import Pricing from "./pages/Pricing";
// import HuggingFace from "./pages/HuggingFace";
import Auth from "./pages/Auth";
import AuthCallback from "./pages/AuthCallback";
import NotFound from "./pages/NotFound";
import { AuthProvider } from "./context/authContext";
import { ModelProvider } from "./context/modelContext";
import { TrainingProvider } from "./context/trainingContext";
import { SubscriptionProvider } from "./context/subscriptionContext";
import ProtectedRoute from "./components/ProtectedRoute";
import { OAUTH_CONFIG } from "./lib/oauth";

const queryClient = new QueryClient();


const App = () => (
  // Always wrap with GoogleOAuthProvider to avoid hook errors
  // Use fallback client ID if none configured
  <GoogleOAuthProvider clientId={OAUTH_CONFIG.google.clientId || 'dummy-client-id'}>
    <BrowserRouter>
      <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
        <AuthProvider>
	  <SubscriptionProvider>
	    <ModelProvider>
	      <TrainingProvider>
	        <QueryClientProvider client={queryClient}>
	    <TooltipProvider>
	      <Toaster />
	      <Sonner />
		<Routes>
		  {/* Public */}
		  <Route path="/auth" element={<Auth />} />
		  <Route path="/auth/callback/:provider" element={<AuthCallback />} />

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
		    path="/community"
		    element={
		      <ProtectedRoute>
			<Layout><Community /></Layout>
		      </ProtectedRoute>
		    }
		  />

		  <Route
		    path="/community/:id"
		    element={
		      <ProtectedRoute>
			<Layout><ModelDetail /></Layout>
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

		  <Route
		    path="/pricing"
		    element={
		      <ProtectedRoute>
			<Layout><Pricing /></Layout>
		      </ProtectedRoute>
		    }
		  />

		  {/* HuggingFace integration - commented out
		  <Route
		    path="/huggingface"
		    element={
		      <ProtectedRoute>
			<Layout><HuggingFace /></Layout>
		      </ProtectedRoute>
		    }
		  />
		  */}

		  <Route path="*" element={<NotFound />} />
		</Routes>
	    </TooltipProvider>
	        </QueryClientProvider>
	      </TrainingProvider>
	    </ModelProvider>
	  </SubscriptionProvider>
	  </AuthProvider>
	  </ThemeProvider>
      </BrowserRouter>
  </GoogleOAuthProvider>
);

export default App;
