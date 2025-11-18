import { useContext, useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { AuthContext } from "@/context/authContext";
import { useToast } from "@/hooks/use-toast";
import { validatePassword } from "@/lib/validation";
import { signIn as authSignIn, providers } from "@/lib/auth";
import { Check, X } from "lucide-react";

const Auth = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const [showPasswordReqs, setShowPasswordReqs] = useState(false);
  const { login, register, loginWithGoogle } = useContext(AuthContext);
  const { toast } = useToast();

  const passwordValidation = validatePassword(password);

  // Check which OAuth providers are configured (Auth.js style)
  const hasGoogleOAuth = providers.google.clientId && providers.google.clientId !== '';
  const hasGitHubOAuth = providers.github.clientId && providers.github.clientId !== '';
  const hasAnyOAuth = hasGoogleOAuth || hasGitHubOAuth;

  // Debug: Log OAuth configuration status
  useEffect(() => {
    console.log('🔍 OAuth Configuration Check:');
    console.log(`  Google: ${hasGoogleOAuth ? '✅ Configured' : '❌ Not configured'}`);
    console.log(`  GitHub: ${hasGitHubOAuth ? '✅ Configured' : '❌ Not configured'}`);
    if (hasGoogleOAuth) {
      console.log(`  Google Client ID: ${providers.google.clientId.substring(0, 10)}...`);
      console.log(`  Google Redirect: ${providers.google.redirectUri}`);
    }
    if (hasGitHubOAuth) {
      console.log(`  GitHub Client ID: ${providers.github.clientId.substring(0, 10)}...`);
      console.log(`  GitHub Redirect: ${providers.github.redirectUri}`);
    }
  }, [hasGoogleOAuth, hasGitHubOAuth]);

  // Auth.js-style OAuth handlers
  const handleGoogleLogin = async () => {
    try {
      await authSignIn("google");
    } catch (error: any) {
      toast({
        title: "Google Login Failed",
        description: error.message || "Could not sign in with Google. Please try again.",
        variant: "destructive",
      });
    }
  };

  const handleGitHubLogin = async () => {
    try {
      await authSignIn("github");
    } catch (error: any) {
      toast({
        title: "GitHub Login Failed",
        description: error.message || "Could not sign in with GitHub. Please try again.",
        variant: "destructive",
      });
    }
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login(email, password);
      toast({
        title: "Welcome back!",
        description: "You have successfully logged in.",
        variant: "default",
      });
    } catch (error) {
      toast({
        title: "Login Failed",
        description: "Invalid email or password. Please try again.",
        variant: "destructive",
      });
    }
  };

  const handleSignup = async (e: React.FormEvent) => {
      e.preventDefault();

    // Validate password before submitting
    if (!passwordValidation.isValid) {
      toast({
        title: "Invalid Password",
        description: passwordValidation.errors.join(". "),
        variant: "destructive",
      });
      return;
    }

    try {
      await register(email, password, username);
      toast({
        title: "Account Created!",
        description: "Your account has been successfully created. You can now log in.",
        variant: "default",
      });
      // Clear form
      setEmail("");
      setPassword("");
      setUsername("");
      setShowPasswordReqs(false);
    } catch (error: any) {
      toast({
        title: "Registration Failed",
        description: error.message || "Could not create account. Username or email might already be in use.",
        variant: "destructive",
      });
    }
  };
  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md border-border">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold bg-gradient-primary bg-clip-text text-transparent">
            AI Model Manager
          </CardTitle>
          <CardDescription className="text-muted-foreground">
            Sign in to your account or create a new one
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="login" className="w-full">
            <TabsList className="grid w-full grid-cols-2 mb-6">
              <TabsTrigger value="login">Login</TabsTrigger>
              <TabsTrigger value="signup">Sign Up</TabsTrigger>
            </TabsList>

            <TabsContent value="login">
              <form onSubmit={handleLogin} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="login-email">Email</Label>
                  <Input
                    id="login-email"
                    type="email"
                    placeholder="name@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="login-password">Password</Label>
                  <Input
                    id="login-password"
                    type="password"
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                  />
                </div>
                <Button type="submit" className="w-full" onClick={handleLogin}>
                  Sign In
                </Button>

                {hasAnyOAuth && (
                  <>
                    <div className="relative my-6">
                      <Separator />
                      <span className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 bg-background px-2 text-xs text-muted-foreground">
                        OR CONTINUE WITH
                      </span>
                    </div>

                    <div className="flex gap-2 justify-center">
                      {hasGoogleOAuth && (
                        <Button
                          type="button"
                          variant="outline"
                          className="flex-1"
                          onClick={() => handleGoogleLogin()}
                          title="Continue with Google"
                        >
                          <svg className="w-5 h-5" viewBox="0 0 24 24">
                            <path
                              fill="#4285F4"
                              d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                            />
                            <path
                              fill="#34A853"
                              d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                            />
                            <path
                              fill="#FBBC05"
                              d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                            />
                            <path
                              fill="#EA4335"
                              d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                            />
                          </svg>
                        </Button>
                      )}
                      {hasGitHubOAuth && (
                        <Button
                          type="button"
                          variant="outline"
                          className="flex-1"
                          onClick={handleGitHubLogin}
                          title="Continue with GitHub"
                        >
                          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                          </svg>
                        </Button>
                      )}
                    </div>
                  </>
                )}
              </form>
            </TabsContent>

            <TabsContent value="signup">
              <form onSubmit={handleSignup} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="signup-username">Username</Label>
                  <Input
                    id="signup-username"
                    type="text"
                    placeholder="johndoe"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                    minLength={3}
                    maxLength={50}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="signup-email">Email</Label>
                  <Input
                    id="signup-email"
                    type="email"
                    placeholder="name@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="signup-password">Password</Label>
                  <Input
                    id="signup-password"
                    type="password"
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    onFocus={() => setShowPasswordReqs(true)}
                    required
                  />
                  {showPasswordReqs && password && (
                    <div className="mt-2 p-3 bg-muted rounded-md space-y-2 text-sm">
                      <p className="font-medium text-muted-foreground mb-2">Password requirements:</p>
                      <div className="space-y-1">
                        <div className={`flex items-center gap-2 ${passwordValidation.requirements.minLength ? 'text-green-600' : 'text-destructive'}`}>
                          {passwordValidation.requirements.minLength ? <Check className="h-4 w-4" /> : <X className="h-4 w-4" />}
                          <span>At least 8 characters</span>
                        </div>
                        <div className={`flex items-center gap-2 ${passwordValidation.requirements.hasLetter ? 'text-green-600' : 'text-destructive'}`}>
                          {passwordValidation.requirements.hasLetter ? <Check className="h-4 w-4" /> : <X className="h-4 w-4" />}
                          <span>Contains at least one letter</span>
                        </div>
                        <div className={`flex items-center gap-2 ${passwordValidation.requirements.hasNumber ? 'text-green-600' : 'text-destructive'}`}>
                          {passwordValidation.requirements.hasNumber ? <Check className="h-4 w-4" /> : <X className="h-4 w-4" />}
                          <span>Contains at least one number</span>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
                <Button type="submit" className="w-full" onSubmit={handleSignup}>
                  Create Account
                </Button>

                {hasAnyOAuth && (
                  <>
                    <div className="relative my-6">
                      <Separator />
                      <span className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 bg-background px-2 text-xs text-muted-foreground">
                        OR CONTINUE WITH
                      </span>
                    </div>

                    <div className="flex gap-2 justify-center">
                      {hasGoogleOAuth && (
                        <Button
                          type="button"
                          variant="outline"
                          className="flex-1"
                          onClick={() => handleGoogleLogin()}
                          title="Continue with Google"
                        >
                          <svg className="w-5 h-5" viewBox="0 0 24 24">
                            <path
                              fill="#4285F4"
                              d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                            />
                            <path
                              fill="#34A853"
                              d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                            />
                            <path
                              fill="#FBBC05"
                              d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                            />
                            <path
                              fill="#EA4335"
                              d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                            />
                          </svg>
                        </Button>
                      )}
                      {hasGitHubOAuth && (
                        <Button
                          type="button"
                          variant="outline"
                          className="flex-1"
                          onClick={handleGitHubLogin}
                          title="Continue with GitHub"
                        >
                          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                          </svg>
                        </Button>
                      )}
                    </div>
                  </>
                )}
              </form>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
};

export default Auth;
