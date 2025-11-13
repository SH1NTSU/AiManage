import { createContext, useState, useEffect, ReactNode } from "react";
import axios from "axios";

export interface Subscription {
  tier: "free" | "basic" | "pro" | "enterprise";
  status: "active" | "canceled" | "expired" | "past_due";
  training_credits: number;
  start_date?: string;
  end_date?: string;
}

export interface AgentSystemInfo {
  python_version?: string;
  pytorch_version?: string;
  cuda_available?: boolean;
  gpu_count?: number;
  gpu_name?: string;
  platform?: string;
}

interface SubscriptionContextType {
  subscription: Subscription | null;
  loading: boolean;
  refreshSubscription: () => Promise<void>;
  canTrainOnServer: boolean;
  isAgentConnected: boolean;
  agentStatus: string;
  agentSystemInfo: AgentSystemInfo | null;
}

export const SubscriptionContext = createContext<SubscriptionContextType | null>(null);

export const SubscriptionProvider = ({ children }: { children: ReactNode }) => {
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [loading, setLoading] = useState(true);
  const [isAgentConnected, setIsAgentConnected] = useState(false);
  const [agentStatus, setAgentStatus] = useState("disconnected");
  const [agentSystemInfo, setAgentSystemInfo] = useState<AgentSystemInfo | null>(null);

  const fetchSubscription = async () => {
    // Only fetch if user is authenticated
    const token = localStorage.getItem("token");
    if (!token) {
      setSubscription({
        tier: "free",
        status: "active",
        training_credits: 0,
      });
      setLoading(false);
      return;
    }

    try {
      console.log("Fetching subscription from server...");
      const response = await axios.get("http://localhost:8081/v1/subscription", {
        headers: { Authorization: `Bearer ${token}` }
      });
      console.log("Subscription API response:", response.data);
      setSubscription(response.data.subscription);
    } catch (error) {
      console.error("Failed to fetch subscription:", error);
      // Set default free tier on error
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
    // Only fetch if user is authenticated
    const token = localStorage.getItem("token");
    if (!token) {
      setIsAgentConnected(false);
      setAgentStatus("disconnected");
      return;
    }

    try {
      const response = await axios.get("http://localhost:8081/v1/agent/status", {
        headers: { Authorization: `Bearer ${token}` }
      });
      setIsAgentConnected(response.data.connected);
      setAgentStatus(response.data.status);
      setAgentSystemInfo(response.data.system_info || null);
    } catch (error) {
      console.error("Failed to fetch agent status:", error);
      setIsAgentConnected(false);
      setAgentStatus("disconnected");
      setAgentSystemInfo(null);
    }
  };

  useEffect(() => {
    fetchSubscription();

    // Fetch initial agent status
    fetchAgentStatus();

    // Set up WebSocket for real-time agent status updates
    const token = localStorage.getItem("token");
    if (!token) return;

    const ws = new WebSocket(`ws://localhost:8081/v1/ws?token=${token}`);

    ws.onopen = () => {
      console.log("✅ WebSocket connected for agent status");
      // Fetch agent status again when WebSocket connects to ensure we have the latest state
      fetchAgentStatus();
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        // Check if this is an agent status message
        if (data.type === "agent_status" && data.data) {
          console.log("📡 Received agent status update:", data.data);
          setIsAgentConnected(data.data.connected);
          setAgentStatus(data.data.status);
          setAgentSystemInfo(data.data.system_info || null);
        }
      } catch (error) {
        console.error("Failed to parse WebSocket message:", error);
      }
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    ws.onclose = () => {
      console.log("WebSocket closed");
      // When WebSocket closes, poll to check if agent is still connected
      fetchAgentStatus();
    };

    // Poll agent status every 10 seconds to ensure we catch status changes
    // even if WebSocket fails or disconnects
    const interval = setInterval(fetchAgentStatus, 10000);

    return () => {
      ws.close();
      clearInterval(interval);
    };
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
        agentSystemInfo,
      }}
    >
      {children}
    </SubscriptionContext.Provider>
  );
};
