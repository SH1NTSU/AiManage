import { useState, useEffect, useContext } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Check } from "lucide-react";
import axios from "axios";
import { useToast } from "@/hooks/use-toast";
import { SubscriptionContext } from "@/context/subscriptionContext";

interface PricingTier {
  tier: string;
  name: string;
  price: number;
  training_credits: number;
  features: string[];
}

const Pricing = () => {
  const [pricing, setPricing] = useState<PricingTier[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentTier, setCurrentTier] = useState("free");
  const { toast } = useToast();
  const subscriptionContext = useContext(SubscriptionContext);

  useEffect(() => {
    fetchPricing();
    fetchCurrentSubscription();
  }, []);

  const fetchPricing = async () => {
    try {
      const response = await axios.get("http://localhost:8081/v1/pricing");
      setPricing(response.data.pricing);
    } catch (error) {
      console.error("Failed to fetch pricing:", error);
    } finally {
      setLoading(false);
    }
  };

  const fetchCurrentSubscription = async () => {
    try {
      const token = localStorage.getItem("token");
      if (!token) return;

      const response = await axios.get("http://localhost:8081/v1/subscription", {
        headers: { Authorization: `Bearer ${token}` }
      });
      setCurrentTier(response.data.subscription.tier);
    } catch (error) {
      console.error("Failed to fetch subscription:", error);
    }
  };

  const handleSubscribe = async (tier: string) => {
    if (tier === "free") {
      toast({
        title: "Already on Free Plan",
        description: "You're currently on the free plan. Upgrade to access server training!",
      });
      return;
    }

    try {
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Authentication Required",
          description: "Please log in to subscribe",
          variant: "destructive",
        });
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
          description: "Failed to create checkout session",
          variant: "destructive",
        });
      }
    } catch (error: any) {
      console.error("Checkout error:", error);
      toast({
        title: "Checkout Failed",
        description: error.response?.data?.message || error.message || "Please try again later.",
        variant: "destructive",
      });
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-16 text-center">
        <p>Loading pricing...</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-16">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold mb-4">Choose Your Plan</h1>
        <p className="text-xl text-muted-foreground">
          Train locally for free, or upgrade for powerful server infrastructure
        </p>
      </div>

      {/* Training Options Banner */}
      <div className="bg-muted p-6 rounded-lg mb-12 text-center">
        <h2 className="text-2xl font-semibold mb-2">🚀 Two Ways to Train</h2>
        <div className="grid md:grid-cols-2 gap-6 mt-6">
          <div className="bg-background p-6 rounded-lg">
            <h3 className="text-xl font-semibold mb-2">💻 Local Training</h3>
            <p className="text-muted-foreground mb-4">
              Train on your own machine - free forever!
            </p>
            <Button variant="outline" onClick={() => window.open("/training-cli", "_blank")}>
              Download Training CLI
            </Button>
          </div>
          <div className="bg-background p-6 rounded-lg">
            <h3 className="text-xl font-semibold mb-2">☁️ Server Training</h3>
            <p className="text-muted-foreground mb-4">
              Train on our powerful GPUs - requires subscription
            </p>
            <Button variant="default">Upgrade to Access</Button>
          </div>
        </div>
      </div>

      {/* Pricing Cards */}
      <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
        {pricing.map((plan) => (
          <Card
            key={plan.tier}
            className={`relative ${
              plan.tier === "pro" ? "border-primary shadow-lg scale-105" : ""
            } ${currentTier === plan.tier ? "border-green-500" : ""}`}
          >
            {plan.tier === "pro" && (
              <div className="absolute -top-4 left-1/2 -translate-x-1/2 bg-primary text-primary-foreground px-4 py-1 rounded-full text-sm font-semibold">
                Most Popular
              </div>
            )}

            {currentTier === plan.tier && (
              <div className="absolute -top-4 right-4 bg-green-500 text-white px-3 py-1 rounded-full text-xs font-semibold">
                Current Plan
              </div>
            )}

            <CardHeader>
              <CardTitle className="text-2xl">{plan.name}</CardTitle>
              <CardDescription>
                {plan.tier === "free" ? (
                  <span className="text-3xl font-bold">Free</span>
                ) : (
                  <span>
                    <span className="text-3xl font-bold">
                      ${(plan.price / 100).toFixed(2)}
                    </span>
                    <span className="text-muted-foreground">/month</span>
                  </span>
                )}
              </CardDescription>
            </CardHeader>

            <CardContent>
              {plan.training_credits === 0 ? (
                <div className="mb-4 p-3 bg-muted rounded-md">
                  <p className="text-sm font-semibold">Local Training Only</p>
                  <p className="text-xs text-muted-foreground">
                    Train on your own machine
                  </p>
                </div>
              ) : plan.training_credits === 999 ? (
                <div className="mb-4 p-3 bg-primary/10 rounded-md">
                  <p className="text-sm font-semibold">Unlimited Server Training</p>
                  <p className="text-xs text-muted-foreground">
                    Train as much as you need
                  </p>
                </div>
              ) : (
                <div className="mb-4 p-3 bg-primary/10 rounded-md">
                  <p className="text-sm font-semibold">
                    {plan.training_credits} Server Training Jobs
                  </p>
                  <p className="text-xs text-muted-foreground">per month</p>
                </div>
              )}

              <ul className="space-y-2">
                {plan.features.map((feature, index) => (
                  <li key={index} className="flex items-start gap-2 text-sm">
                    <Check className="h-4 w-4 text-green-600 mt-0.5 flex-shrink-0" />
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>
            </CardContent>

            <CardFooter>
              <Button
                className="w-full"
                variant={currentTier === plan.tier ? "outline" : "default"}
                onClick={() => handleSubscribe(plan.tier)}
                disabled={currentTier === plan.tier}
              >
                {currentTier === plan.tier
                  ? "Current Plan"
                  : plan.tier === "free"
                  ? "Get Started"
                  : "Upgrade"}
              </Button>
            </CardFooter>
          </Card>
        ))}
      </div>

      {/* FAQ Section */}
      <div className="mt-16 text-center">
        <h2 className="text-2xl font-bold mb-4">Frequently Asked Questions</h2>
        <div className="max-w-2xl mx-auto text-left space-y-4">
          <div>
            <h3 className="font-semibold mb-2">Can I train for free?</h3>
            <p className="text-muted-foreground">
              Yes! Download our training CLI and train on your own machine completely free.
            </p>
          </div>
          <div>
            <h3 className="font-semibold mb-2">What happens if I run out of credits?</h3>
            <p className="text-muted-foreground">
              Your credits reset monthly. You can always train locally or upgrade to a higher tier.
            </p>
          </div>
          <div>
            <h3 className="font-semibold mb-2">Can I cancel anytime?</h3>
            <p className="text-muted-foreground">
              Yes! Cancel anytime from your account settings. No questions asked.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Pricing;
