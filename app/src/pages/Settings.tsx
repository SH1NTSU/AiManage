import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Save, Bell, Shield, Database, Palette } from "lucide-react";

const Settings = () => {
  return (
    <div className="space-y-6 animate-slide-up">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground mt-1">
          Configure your AI model manager preferences
        </p>
      </div>

      <div className="grid gap-6">
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
                <Label>Dark Mode</Label>
                <p className="text-sm text-muted-foreground">
                  Enable dark theme for the interface
                </p>
              </div>
              <Switch defaultChecked />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Animations</Label>
                <p className="text-sm text-muted-foreground">
                  Enable smooth transitions and effects
                </p>
              </div>
              <Switch defaultChecked />
            </div>
          </CardContent>
        </Card>

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
                <Label>Model Alerts</Label>
                <p className="text-sm text-muted-foreground">
                  Receive alerts when models fail or respond slowly
                </p>
              </div>
              <Switch defaultChecked />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Usage Reports</Label>
                <p className="text-sm text-muted-foreground">
                  Get weekly usage reports via email
                </p>
              </div>
              <Switch />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-gradient-card border-border shadow-card">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Database className="w-5 h-5 text-chart-3" />
              <CardTitle>Storage</CardTitle>
            </div>
            <CardDescription>Configure storage paths and limits</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="model-path">Default Model Path</Label>
              <Input
                id="model-path"
                placeholder="D:\AI\Models"
                defaultValue="D:\AI\Models"
                className="bg-muted border-input"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="cache-path">Cache Directory</Label>
              <Input
                id="cache-path"
                placeholder="D:\AI\Cache"
                defaultValue="D:\AI\Cache"
                className="bg-muted border-input"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="max-storage">Max Storage (GB)</Label>
              <Input
                id="max-storage"
                type="number"
                placeholder="100"
                defaultValue="100"
                className="bg-muted border-input"
              />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-gradient-card border-border shadow-card">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Shield className="w-5 h-5 text-chart-4" />
              <CardTitle>Security</CardTitle>
            </div>
            <CardDescription>Manage security and access controls</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="api-key">API Key</Label>
              <Input
                id="api-key"
                type="password"
                placeholder="••••••••••••••••"
                className="bg-muted border-input"
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Two-Factor Authentication</Label>
                <p className="text-sm text-muted-foreground">
                  Add an extra layer of security
                </p>
              </div>
              <Switch />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Auto-lock</Label>
                <p className="text-sm text-muted-foreground">
                  Lock after 15 minutes of inactivity
                </p>
              </div>
              <Switch defaultChecked />
            </div>
          </CardContent>
        </Card>

        <div className="flex justify-end gap-4">
          <Button variant="outline" className="border-border">
            Reset to Defaults
          </Button>
          <Button className="bg-gradient-primary hover:opacity-90 shadow-glow">
            <Save className="w-4 h-4 mr-2" />
            Save Changes
          </Button>
        </div>
      </div>
    </div>
  );
};

export default Settings;
