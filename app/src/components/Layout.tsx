import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/AppSidebar";
import { Badge } from "@/components/ui/badge";
import { ReactNode, useContext } from "react";
import { SubscriptionContext } from "@/context/subscriptionContext";
import { Activity, Zap, Crown, Sparkles } from "lucide-react";

interface LayoutProps {
  children: ReactNode;
}

export const Layout = ({ children }: LayoutProps) => {
  const subscriptionContext = useContext(SubscriptionContext);

  const getTierIcon = (tier: string) => {
    switch (tier) {
      case "basic":
        return Zap;
      case "pro":
        return Crown;
      case "enterprise":
        return Sparkles;
      default:
        return null;
    }
  };

  const TierIcon = subscriptionContext?.subscription ? getTierIcon(subscriptionContext.subscription.tier) : null;

  return (
    <SidebarProvider>
      <div className="min-h-screen flex w-full bg-background">
        <AppSidebar />
        <main className="flex-1 flex flex-col">
          <header className="h-16 border-b border-border flex items-center justify-between px-6 bg-card/50 backdrop-blur-sm sticky top-0 z-10">
            <div className="flex items-center">
              <SidebarTrigger className="mr-4" />
              <h1 className="text-xl font-semibold bg-gradient-primary bg-clip-text text-transparent">
                AI Model Manager
              </h1>
            </div>

            <div className="flex items-center gap-3">
              {/* Agent Status Indicator */}
              <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-muted/50">
                <Activity className={`w-4 h-4 ${subscriptionContext?.isAgentConnected ? 'text-green-500' : 'text-gray-400'}`} />
                <div className={`w-2 h-2 rounded-full ${subscriptionContext?.isAgentConnected ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`} />
                <span className="text-xs font-medium">
                  {subscriptionContext?.isAgentConnected ? 'Agent Connected' : 'Agent Offline'}
                </span>
              </div>

              {/* Subscription Tier Badge */}
              {subscriptionContext?.subscription && subscriptionContext.subscription.tier !== "free" && (
                <Badge className="bg-gradient-to-r from-blue-500 to-purple-500 text-white">
                  {TierIcon && <TierIcon className="w-3 h-3 mr-1" />}
                  {subscriptionContext.subscription.tier.toUpperCase()}
                </Badge>
              )}
            </div>
          </header>
          <div className="flex-1 p-6">
            {children}
          </div>
        </main>
      </div>
    </SidebarProvider>
  );
};
