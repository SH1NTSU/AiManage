import { useEffect, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { handleCallback } from "@/lib/auth";
import { useToast } from "@/hooks/use-toast";

// OAuth callback handler (Auth.js style)
const AuthCallback = () => {
  const { provider } = useParams<{ provider: "google" | "github" }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Immediate log to verify component is rendering
  useEffect(() => {
    console.log("🚀 AuthCallback component mounted");
    console.log("📋 Provider from URL:", provider);
    console.log("📋 Search params:", Object.fromEntries(searchParams.entries()));
    console.log("📋 Current URL:", window.location.href);
  }, []);

  useEffect(() => {
    const processCallback = async () => {
      try {
        console.log("🔄 Processing OAuth callback...");
        console.log("📋 Provider:", provider);
        console.log("📋 URL params:", Object.fromEntries(searchParams.entries()));

        if (!provider || (provider !== "google" && provider !== "github")) {
          console.error("❌ Invalid provider:", provider);
          setStatus("error");
          toast({
            title: "Invalid Provider",
            description: "Invalid OAuth provider specified.",
            variant: "destructive",
          });
          setTimeout(() => navigate("/auth"), 2000);
          return;
        }

        const code = searchParams.get("code");
        const state = searchParams.get("state");
        const error = searchParams.get("error");

        console.log("📋 Code:", code ? `${code.substring(0, 10)}...` : "missing");
        console.log("📋 State:", state ? `${state.substring(0, 10)}...` : "missing");
        console.log("📋 Error:", error || "none");

        if (error) {
          console.error("❌ OAuth error from provider:", error);
          setStatus("error");
          toast({
            title: "OAuth Error",
            description: error || "Authentication failed.",
            variant: "destructive",
          });
          setTimeout(() => navigate("/auth"), 2000);
          return;
        }

        if (!code || !state) {
          console.error("❌ Missing required parameters");
          setStatus("error");
          toast({
            title: "Missing Parameters",
            description: "Missing authorization code or state parameter.",
            variant: "destructive",
          });
          setTimeout(() => navigate("/auth"), 2000);
          return;
        }

        console.log("🔄 Calling handleCallback...");
        const { token } = await handleCallback(provider, code, state);
        
        if (!token) {
          throw new Error("No token received from backend");
        }

        console.log("✅ Token received, storing...");
        // Store token
        localStorage.setItem("token", token);
        
        setStatus("success");
        toast({
          title: "Success!",
          description: `Successfully signed in with ${provider === "google" ? "Google" : "GitHub"}.`,
          variant: "default",
        });
        
        // Redirect to home
        setTimeout(() => navigate("/"), 1000);
      } catch (error: any) {
        console.error("❌ Callback processing error:", error);
        const errorMsg = error.message || "Failed to complete authentication.";
        setErrorMessage(errorMsg);
        setStatus("error");
        toast({
          title: "Authentication Failed",
          description: errorMsg,
          variant: "destructive",
        });
        setTimeout(() => navigate("/auth"), 2000);
      }
    };

    processCallback();
  }, [provider, searchParams, navigate, toast]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md border-border">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold bg-gradient-primary bg-clip-text text-transparent">
            {status === "loading" && "Completing Sign In..."}
            {status === "success" && "Sign In Successful!"}
            {status === "error" && "Sign In Failed"}
          </CardTitle>
          <CardDescription className="text-muted-foreground">
            {status === "loading" && "Please wait while we complete your authentication."}
            {status === "success" && "Redirecting you to the application..."}
            {status === "error" && "Redirecting you back to the sign in page..."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center gap-4">
            {status === "loading" && (
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            )}
            {status === "success" && (
              <div className="text-green-500 text-4xl">✓</div>
            )}
            {status === "error" && (
              <>
                <div className="text-red-500 text-4xl">✗</div>
                {errorMessage && (
                  <p className="text-sm text-destructive text-center max-w-md">
                    {errorMessage}
                  </p>
                )}
                <p className="text-xs text-muted-foreground text-center mt-2">
                  Check browser console (F12) for more details
                </p>
              </>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default AuthCallback;

