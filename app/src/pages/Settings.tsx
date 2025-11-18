import { useState, useContext, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useTheme } from "next-themes";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  LogOut, Bell, Palette, User, CreditCard, Key, Activity,
  Monitor, Sun, Moon, Check, Copy, Download, Zap, Crown, Sparkles
} from "lucide-react";
import { AuthContext } from "@/context/authContext";
import { SubscriptionContext } from "@/context/subscriptionContext";
import { useToast } from "@/hooks/use-toast";
import axios from "axios";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const Settings = () => {
  const navigate = useNavigate();
  const { toast } = useToast();
  const authContext = useContext(AuthContext);
  const subscriptionContext = useContext(SubscriptionContext);
  const { theme, setTheme } = useTheme();

  const [activeTab, setActiveTab] = useState("account");
  const [apiKey, setApiKey] = useState("");
  const [mockPaymentProcessing, setMockPaymentProcessing] = useState(false);
  const [userEmail, setUserEmail] = useState("");
  const [username, setUsername] = useState("");

  // Appearance settings
  const [animations, setAnimations] = useState(() => {
    const saved = localStorage.getItem("settings_animations");
    return saved === null ? true : saved === "true";
  });

  // Notification settings
  const [trainingAlerts, setTrainingAlerts] = useState(() => {
    const saved = localStorage.getItem("settings_trainingAlerts");
    return saved === null ? true : saved === "true";
  });

  useEffect(() => {
    if (animations) {
      document.documentElement.classList.remove("no-animations");
    } else {
      document.documentElement.classList.add("no-animations");
    }
    localStorage.setItem("settings_animations", animations.toString());
  }, [animations]);

  useEffect(() => {
    localStorage.setItem("settings_trainingAlerts", trainingAlerts.toString());
  }, [trainingAlerts]);

  // Fetch user info and API key
  useEffect(() => {
    const fetchUserInfo = async () => {
      try {
        const token = localStorage.getItem("token");
        if (!token) return;

        const response = await axios.get("http://localhost:8081/v1/me", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (response.data) {
          setUserEmail(response.data.email || "");
          setUsername(response.data.username || "");
          setApiKey(response.data.api_key || "");
        }
      } catch (error) {
        console.error("Failed to fetch user info:", error);
      }
    };

    fetchUserInfo();

    // Handle mock checkout success
    const urlParams = new URLSearchParams(window.location.search);
    const mockCheckout = urlParams.get('mock_checkout');
    const tier = urlParams.get('tier');

    if (mockCheckout === 'true' && tier) {
      toast({
        title: "Mock Checkout",
        description: `This is a demo. In production, you would be charged for the ${tier} tier. Stripe integration needed.`,
      });
      // Clean up URL
      window.history.replaceState({}, '', '/settings');
    }
  }, []);

  const handleLogout = () => {
    if (authContext) {
      authContext.logout();
      toast({
        title: "Logged out",
        description: "You have been successfully logged out",
      });
      navigate("/auth");
    }
  };

  const copyApiKey = () => {
    navigator.clipboard.writeText(apiKey);
    toast({
      title: "Copied!",
      description: "API key copied to clipboard",
    });
  };

  const regenerateApiKey = async () => {
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Error",
          description: "You must be logged in to regenerate your API key",
          variant: "destructive",
        });
        return;
      }

      const response = await axios.post(
        "http://localhost:8081/v1/regenerate-api-key",
        {},
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );

      if (response.data.success && response.data.api_key) {
        setApiKey(response.data.api_key);
        toast({
          title: "Success",
          description: "API key regenerated successfully",
        });
      } else {
        throw new Error("Invalid response from server");
      }
    } catch (error: any) {
      console.error("Failed to regenerate API key:", error);
      toast({
        title: "Error",
        description: error.response?.data?.message || "Failed to regenerate API key",
        variant: "destructive",
      });
    }
  };

  const downloadAgent = () => {
    toast({
      title: "Download Started",
      description: "Training agent package is downloading...",
    });

    // Create a temporary link and trigger download
    const link = document.createElement('a');
    link.href = "http://localhost:8081/uploads/training-agent.zip";
    link.setAttribute('download', 'training-agent.zip');
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const handleUpgrade = async (tier: string) => {
    setMockPaymentProcessing(true);

    try {
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Authentication Required",
          description: "Please log in to subscribe",
          variant: "destructive",
        });
        setMockPaymentProcessing(false);
        return;
      }

      toast({
        title: "Creating checkout session...",
        description: "Please wait...",
      });

      const response = await axios.post("http://localhost:8081/v1/subscription/checkout", {
        tier: tier,
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });


      if (response.data.checkout_url) {
        // Redirect to Stripe checkout
        window.location.href = response.data.checkout_url;
      } else {
        toast({
          title: "Error",
          description: response.data.message || "Failed to create checkout session",
          variant: "destructive",
        });
        setMockPaymentProcessing(false);
      }
    } catch (error: any) {
      console.error("Checkout error:", error);
      toast({
        title: "Checkout Failed",
        description: error.response?.data?.message || error.message || "Please try again later.",
        variant: "destructive",
      });
      setMockPaymentProcessing(false);
    }
  };

  const getTierBadge = (tier: string) => {
    const badges: Record<string, { icon: any; color: string }> = {
      free: { icon: null, color: "bg-gray-500" },
      basic: { icon: Zap, color: "bg-blue-500" },
      pro: { icon: Crown, color: "bg-purple-500" },
      enterprise: { icon: Sparkles, color: "bg-gradient-to-r from-yellow-500 to-orange-500" },
    };

    const badge = badges[tier] || badges.free;
    const Icon = badge.icon;

    return (
      <Badge className={`${badge.color} text-white`}>
        {Icon && <Icon className="w-3 h-3 mr-1" />}
        {tier.toUpperCase()}
      </Badge>
    );
  };

  return (
    <div className="space-y-6 animate-slide-up pb-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground mt-1">
          Manage your account, subscription, and preferences
        </p>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="account">
            <User className="w-4 h-4 mr-2" />
            Account
          </TabsTrigger>
          <TabsTrigger value="subscription">
            <CreditCard className="w-4 h-4 mr-2" />
            Subscription
          </TabsTrigger>
          <TabsTrigger value="agent">
            <Activity className="w-4 h-4 mr-2" />
            Agent
          </TabsTrigger>
          <TabsTrigger value="appearance">
            <Palette className="w-4 h-4 mr-2" />
            Appearance
          </TabsTrigger>
          <TabsTrigger value="notifications">
            <Bell className="w-4 h-4 mr-2" />
            Notifications
          </TabsTrigger>
        </TabsList>

        {/* Account Tab */}
        <TabsContent value="account" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Account Information</CardTitle>
              <CardDescription>Your account details and authentication</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label>Email</Label>
                <Input value={userEmail} disabled className="mt-1.5" />
              </div>
              <div>
                <Label>Username</Label>
                <Input value={username} disabled className="mt-1.5" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>API Keys</CardTitle>
              <CardDescription>Use this key to connect your training agent</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-2">
                <Input value={apiKey} readOnly />
                <Button variant="outline" size="icon" onClick={copyApiKey}>
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
              <Button variant="outline" onClick={regenerateApiKey}>
                <Key className="w-4 h-4 mr-2" />
                Regenerate API Key
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Danger Zone</CardTitle>
              <CardDescription>Irreversible account actions</CardDescription>
            </CardHeader>
            <CardContent>
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive">
                    <LogOut className="w-4 h-4 mr-2" />
                    Logout
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Are you sure you want to logout?</AlertDialogTitle>
                    <AlertDialogDescription>
                      You will need to login again to access your models.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction onClick={handleLogout} className="bg-destructive text-destructive-foreground">
                      Logout
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Subscription Tab */}
        <TabsContent value="subscription" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Current Plan</CardTitle>
              <CardDescription>Manage your subscription and billing</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between p-4 border rounded-lg">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <h3 className="text-lg font-semibold">
                      {subscriptionContext?.subscription?.tier?.toUpperCase() || "FREE"} Plan
                    </h3>
                    {subscriptionContext?.subscription && getTierBadge(subscriptionContext.subscription.tier)}
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {subscriptionContext?.subscription?.tier === "free"
                      ? "Train on your own machine"
                      : `${subscriptionContext?.subscription?.training_credits} training credits remaining`}
                  </p>
                </div>
                {subscriptionContext?.subscription?.tier !== "free" && (
                  <Button variant="outline" size="sm">
                    Manage Billing
                  </Button>
                )}
              </div>

              {subscriptionContext?.subscription?.tier === "free" && (
                <div className="bg-muted p-4 rounded-lg">
                  <h4 className="font-semibold mb-2">🎉 Free Forever</h4>
                  <p className="text-sm text-muted-foreground mb-3">
                    Train unlimited models on your own machine. Upgrade for server training with powerful GPUs.
                  </p>
                  <Button onClick={() => navigate("/pricing")}>
                    View Plans
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Mock Upgrade Options */}
          {subscriptionContext?.subscription?.tier === "free" && (
            <Card>
              <CardHeader>
                <CardTitle>Upgrade Your Plan</CardTitle>
                <CardDescription>Get access to server training and more features</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid md:grid-cols-3 gap-4">
                  <Card>
                    <CardHeader>
                      <Badge className="w-fit bg-blue-500">Basic</Badge>
                      <CardTitle className="text-2xl">$9.99/mo</CardTitle>
                      <CardDescription>10 server trainings/month</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <ul className="space-y-2 text-sm">
                        <li className="flex items-center gap-2">
                          <Check className="h-4 w-4 text-green-600" />
                          Shared GPU access
                        </li>
                        <li className="flex items-center gap-2">
                          <Check className="h-4 w-4 text-green-600" />
                          Priority queue
                        </li>
                      </ul>
                    </CardContent>
                    <CardFooter>
                      <Button
                        className="w-full"
                        onClick={() => handleUpgrade("basic")}
                        disabled={mockPaymentProcessing}
                      >
                        {mockPaymentProcessing ? "Processing..." : "Upgrade to Basic"}
                      </Button>
                    </CardFooter>
                  </Card>

                  <Card className="border-primary">
                    <CardHeader>
                      <Badge className="w-fit bg-purple-500">Pro</Badge>
                      <CardTitle className="text-2xl">$29.99/mo</CardTitle>
                      <CardDescription>50 server trainings/month</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <ul className="space-y-2 text-sm">
                        <li className="flex items-center gap-2">
                          <Check className="h-4 w-4 text-green-600" />
                          Dedicated V100 GPU
                        </li>
                        <li className="flex items-center gap-2">
                          <Check className="h-4 w-4 text-green-600" />
                          API access
                        </li>
                      </ul>
                    </CardContent>
                    <CardFooter>
                      <Button
                        className="w-full"
                        onClick={() => handleUpgrade("pro")}
                        disabled={mockPaymentProcessing}
                      >
                        {mockPaymentProcessing ? "Processing..." : "Upgrade to Pro"}
                      </Button>
                    </CardFooter>
                  </Card>

                  <Card>
                    <CardHeader>
                      <Badge className="w-fit bg-gradient-to-r from-yellow-500 to-orange-500">Enterprise</Badge>
                      <CardTitle className="text-2xl">$99.99/mo</CardTitle>
                      <CardDescription>Unlimited training</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <ul className="space-y-2 text-sm">
                        <li className="flex items-center gap-2">
                          <Check className="h-4 w-4 text-green-600" />
                          Dedicated A100 GPU
                        </li>
                        <li className="flex items-center gap-2">
                          <Check className="h-4 w-4 text-green-600" />
                          24/7 support
                        </li>
                      </ul>
                    </CardContent>
                    <CardFooter>
                      <Button
                        className="w-full"
                        onClick={() => handleUpgrade("enterprise")}
                        disabled={mockPaymentProcessing}
                      >
                        {mockPaymentProcessing ? "Processing..." : "Upgrade to Enterprise"}
                      </Button>
                    </CardFooter>
                  </Card>
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* Agent Tab */}
        <TabsContent value="agent" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Training Agent</CardTitle>
              <CardDescription>Connect your computer to train models using your own resources</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between p-4 border rounded-lg">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <div className={`w-2 h-2 rounded-full ${subscriptionContext?.isAgentConnected ? 'bg-green-500' : 'bg-gray-400'}`} />
                    <h4 className="font-semibold">Agent Status</h4>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {subscriptionContext?.isAgentConnected ? "Connected and ready" : "Not connected"}
                  </p>
                </div>
                <Badge variant={subscriptionContext?.isAgentConnected ? "default" : "secondary"}>
                  {subscriptionContext?.agentStatus}
                </Badge>
              </div>

              {subscriptionContext?.isAgentConnected && subscriptionContext?.agentSystemInfo && (
                <div className="p-4 border rounded-lg space-y-3">
                  <h4 className="font-semibold text-sm">System Information</h4>
                  <div className="grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <p className="text-muted-foreground">Platform</p>
                      <p className="font-mono">{subscriptionContext.agentSystemInfo.platform || 'N/A'}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">Python Version</p>
                      <p className="font-mono text-xs">{subscriptionContext.agentSystemInfo.python_version?.split('\n')[0] || 'N/A'}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">PyTorch</p>
                      <p className="font-mono">{subscriptionContext.agentSystemInfo.pytorch_version || 'N/A'}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">GPU</p>
                      <p className="font-mono">
                        {subscriptionContext.agentSystemInfo.cuda_available
                          ? `${subscriptionContext.agentSystemInfo.gpu_count}x ${subscriptionContext.agentSystemInfo.gpu_name}`
                          : 'CPU Only'}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {!subscriptionContext?.isAgentConnected && (
                <div className="bg-muted p-4 rounded-lg space-y-3">
                  <h4 className="font-semibold">Get Started with Agent Training</h4>
                  <ol className="text-sm text-muted-foreground space-y-2 list-decimal list-inside">
                    <li>Download the training agent</li>
                    <li>Run it with your API key</li>
                    <li>Start training from the web interface</li>
                  </ol>
                  <Button onClick={downloadAgent} className="w-full">
                    <Download className="w-4 h-4 mr-2" />
                    Download Training Agent
                  </Button>
                </div>
              )}

              <div className="p-4 bg-blue-50 dark:bg-blue-950 rounded-lg">
                <h4 className="font-semibold text-sm mb-2">💡 How it works</h4>
                <p className="text-xs text-muted-foreground">
                  The training agent runs on your computer and connects to our platform.
                  You manage everything from this web interface, but training happens on your machine using your resources.
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Agent Setup Instructions</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2 text-sm font-mono bg-muted p-4 rounded-lg">
                <p># 1. Install dependencies</p>
                <p>pip install websockets torch</p>
                <p className="mt-2"># 2. Run the agent</p>
                <p>python train_agent.py --api-key {apiKey}</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Appearance Tab */}
        <TabsContent value="appearance" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Appearance</CardTitle>
              <CardDescription>Customize the look and feel</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>Theme</Label>
                  <p className="text-sm text-muted-foreground">
                    Select your preferred color theme
                  </p>
                </div>
                <Select value={theme} onValueChange={setTheme}>
                  <SelectTrigger className="w-[180px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="light">
                      <div className="flex items-center gap-2">
                        <Sun className="w-4 h-4" />
                        Light
                      </div>
                    </SelectItem>
                    <SelectItem value="dark">
                      <div className="flex items-center gap-2">
                        <Moon className="w-4 h-4" />
                        Dark
                      </div>
                    </SelectItem>
                    <SelectItem value="system">
                      <div className="flex items-center gap-2">
                        <Monitor className="w-4 h-4" />
                        System
                      </div>
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Separator />
              <div className="flex items-center justify-between">
                <div>
                  <Label>Animations</Label>
                  <p className="text-sm text-muted-foreground">
                    Enable smooth transitions
                  </p>
                </div>
                <Switch checked={animations} onCheckedChange={setAnimations} />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Notifications Tab */}
        <TabsContent value="notifications" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Notifications</CardTitle>
              <CardDescription>Manage your notification preferences</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>Training Alerts</Label>
                  <p className="text-sm text-muted-foreground">
                    Get notified when training completes
                  </p>
                </div>
                <Switch checked={trainingAlerts} onCheckedChange={setTrainingAlerts} />
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default Settings;
