import { createContext, useState, useEffect, ReactNode } from "react";
import axios from "axios";

export interface Subscription {
  tier: "free" | "basic" | "pro" | "enterprise";
  status: "active" | "canceled" | "expired" | "past_due";
  training_credits: number;
  start_date?: string;
  end_date?: string;
}

interface SubscriptionContextType {
  subscription: Subscription | null;
  loading: boolean;
  refreshSubscription: () => Promise<void>;
  canTrainOnServer: boolean;
  isAgentConnected: boolean;
  agentStatus: string;
}

export const SubscriptionContext = createContext<SubscriptionContextType | null>(null);

export const SubscriptionProvider = ({ children }: { children: ReactNode }) => {
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [loading, setLoading] = useState(true);
  const [isAgentConnected, setIsAgentConnected] = useState(false);
  const [agentStatus, setAgentStatus] = useState("disconnected");

  const fetchSubscription = async () => {
    try {
      const response = await axios.get("http://localhost:8081/v1/subscription");
      setSubscription(response.data.subscription);
    } catch (error) {
      console.error("Failed to fetch subscription:", error);
      // Set default free tier
      setSubscription({
        tier: "free",
        status: "active",
        training_credits: 0,
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchAgentStatus = async () => {
    try {
      const response = await axios.get("http://localhost:8081/v1/agent/status");
      setIsAgentConnected(response.data.connected);
      setAgentStatus(response.data.status);
    } catch (error) {
      console.error("Failed to fetch agent status:", error);
      setIsAgentConnected(false);
      setAgentStatus("disconnected");
    }
  };

  useEffect(() => {
    fetchSubscription();
    fetchAgentStatus();

    // Poll agent status every 30 seconds
    const interval = setInterval(fetchAgentStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  const canTrainOnServer = subscription
    ? subscription.tier !== "free" && subscription.status === "active"
    : false;

  return (
    <SubscriptionContext.Provider
      value={{
        subscription,
        loading,
        refreshSubscription: fetchSubscription,
        canTrainOnServer,
        isAgentConnected,
        agentStatus,
      }}
    >
      {children}
    </SubscriptionContext.Provider>
  );
};
