import { useState, useContext, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useTheme } from "next-themes";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { LogOut, Bell, Palette, User, Mail, Monitor, Sun, Moon } from "lucide-react";
import { AuthContext } from "@/context/authContext";
import { useToast } from "@/hooks/use-toast";
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
  const { theme, setTheme } = useTheme();

  // Appearance settings
  const [animations, setAnimations] = useState(() => {
    const saved = localStorage.getItem("settings_animations");
    return saved === null ? true : saved === "true";
  });
  const [compactMode, setCompactMode] = useState(() => {
    const saved = localStorage.getItem("settings_compactMode");
    return saved === "true";
  });

  // Notification settings
  const [trainingAlerts, setTrainingAlerts] = useState(() => {
    const saved = localStorage.getItem("settings_trainingAlerts");
    return saved === null ? true : saved === "true";
  });
  const [downloadNotifications, setDownloadNotifications] = useState(() => {
    const saved = localStorage.getItem("settings_downloadNotifications");
    return saved === null ? true : saved === "true";
  });
  const [communityUpdates, setCommunityUpdates] = useState(() => {
    const saved = localStorage.getItem("settings_communityUpdates");
    return saved === "true";
  });
  const [emailNotifications, setEmailNotifications] = useState(() => {
    const saved = localStorage.getItem("settings_emailNotifications");
    return saved === "true";
  });

  // Apply and auto-save animations setting
  useEffect(() => {
    if (animations) {
      document.documentElement.classList.remove("no-animations");
    } else {
      document.documentElement.classList.add("no-animations");
    }
    localStorage.setItem("settings_animations", animations.toString());
  }, [animations]);

  // Apply and auto-save compact mode
  useEffect(() => {
    if (compactMode) {
      document.documentElement.classList.add("compact-mode");
    } else {
      document.documentElement.classList.remove("compact-mode");
    }
    localStorage.setItem("settings_compactMode", compactMode.toString());
  }, [compactMode]);

  // Auto-save notification settings
  useEffect(() => {
    localStorage.setItem("settings_trainingAlerts", trainingAlerts.toString());
  }, [trainingAlerts]);

  useEffect(() => {
    localStorage.setItem("settings_downloadNotifications", downloadNotifications.toString());
  }, [downloadNotifications]);

  useEffect(() => {
    localStorage.setItem("settings_communityUpdates", communityUpdates.toString());
  }, [communityUpdates]);

  useEffect(() => {
    localStorage.setItem("settings_emailNotifications", emailNotifications.toString());
  }, [emailNotifications]);

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

  const handleThemeChange = (newTheme: string) => {
    setTheme(newTheme);
    toast({
      title: "Theme changed",
      description: `Theme set to ${newTheme}`,
    });
  };

  return (
    <div className="space-y-6 animate-slide-up pb-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground mt-1">
          Manage your account preferences and settings
        </p>
      </div>

      <div className="grid gap-6">
        {/* Account Section */}
        <Card className="bg-gradient-card border-border shadow-card">
          <CardHeader>
            <div className="flex items-center gap-2">
              <User className="w-5 h-5 text-primary" />
              <CardTitle>Account</CardTitle>
            </div>
            <CardDescription>Manage your account and sessions</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Logout</Label>
                <p className="text-sm text-muted-foreground">
                  Sign out of your account on this device
                </p>
              </div>
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive" size="sm">
                    <LogOut className="w-4 h-4 mr-2" />
                    Logout
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Are you sure you want to logout?</AlertDialogTitle>
                    <AlertDialogDescription>
                      You will need to login again to access your models and community features.
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
            </div>
          </CardContent>
        </Card>

        {/* Appearance Section */}
        <Card className="bg-gradient-card border-border shadow-card">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Palette className="w-5 h-5 text-primary" />
              <CardTitle>Appearance</CardTitle>
            </div>
            <CardDescription>Customize the look and feel of your application</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Theme</Label>
                <p className="text-sm text-muted-foreground">
                  Select your preferred color theme
                </p>
              </div>
              <Select value={theme} onValueChange={handleThemeChange}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Select theme" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="light">
                    <div className="flex items-center gap-2">
                      <Sun className="w-4 h-4" />
                      <span>Light</span>
                    </div>
                  </SelectItem>
                  <SelectItem value="dark">
                    <div className="flex items-center gap-2">
                      <Moon className="w-4 h-4" />
                      <span>Dark</span>
                    </div>
                  </SelectItem>
                  <SelectItem value="system">
                    <div className="flex items-center gap-2">
                      <Monitor className="w-4 h-4" />
                      <span>System</span>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Animations</Label>
                <p className="text-sm text-muted-foreground">
                  Enable smooth transitions and effects
                </p>
              </div>
              <Switch
                checked={animations}
                onCheckedChange={setAnimations}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Compact Mode</Label>
                <p className="text-sm text-muted-foreground">
                  Display more content with reduced spacing
                </p>
              </div>
              <Switch
                checked={compactMode}
                onCheckedChange={setCompactMode}
              />
            </div>
          </CardContent>
        </Card>

        {/* Notifications Section */}
        <Card className="bg-gradient-card border-border shadow-card">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Bell className="w-5 h-5 text-secondary" />
              <CardTitle>Notifications</CardTitle>
            </div>
            <CardDescription>Manage your notification preferences</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Training Alerts</Label>
                <p className="text-sm text-muted-foreground">
                  Get notified when model training completes or fails
                </p>
              </div>
              <Switch
                checked={trainingAlerts}
                onCheckedChange={setTrainingAlerts}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Download Notifications</Label>
                <p className="text-sm text-muted-foreground">
                  Receive alerts when your published models are downloaded
                </p>
              </div>
              <Switch
                checked={downloadNotifications}
                onCheckedChange={setDownloadNotifications}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Community Updates</Label>
                <p className="text-sm text-muted-foreground">
                  Stay updated with new models and community features
                </p>
              </div>
              <Switch
                checked={communityUpdates}
                onCheckedChange={setCommunityUpdates}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5 flex-1">
                <Label className="flex items-center gap-2">
                  <Mail className="w-4 h-4" />
                  Email Notifications
                </Label>
                <p className="text-sm text-muted-foreground">
                  Receive notification updates via email
                </p>
              </div>
              <Switch
                checked={emailNotifications}
                onCheckedChange={setEmailNotifications}
              />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Settings;
