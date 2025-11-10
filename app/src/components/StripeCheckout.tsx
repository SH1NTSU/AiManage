import { useState } from "react";
import { CardElement, useStripe, useElements } from "@stripe/react-stripe-js";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Card, CardContent } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/hooks/use-toast";
import { CreditCard, Lock, ShieldCheck, CheckCircle2, Loader2 } from "lucide-react";

interface StripeCheckoutProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  modelName: string;
  price: number;
  onSuccess: () => void;
  mockMode?: boolean;
}

const StripeCheckout = ({
  open,
  onOpenChange,
  modelName,
  price,
  onSuccess,
  mockMode = true,
}: StripeCheckoutProps) => {
  const stripe = useStripe();
  const elements = useElements();
  const { toast } = useToast();
  const [processing, setProcessing] = useState(false);
  const [succeeded, setSucceeded] = useState(false);

  const priceInDollars = (price / 100).toFixed(2);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!stripe || !elements) {
      return;
    }

    setProcessing(true);

    try {
      if (mockMode) {
        // Mock payment - simulate processing
        await new Promise((resolve) => setTimeout(resolve, 2000));

        setSucceeded(true);
        toast({
          title: "Payment successful!",
          description: `You've purchased ${modelName}`,
        });

        // Wait a moment to show success state
        setTimeout(() => {
          onSuccess();
          onOpenChange(false);
          setSucceeded(false);
        }, 1500);
      } else {
        // Real Stripe payment would go here
        const cardElement = elements.getElement(CardElement);

        if (!cardElement) {
          throw new Error("Card element not found");
        }

        // In production, you'd create a payment intent on your backend
        // and confirm the payment here
        const { error, paymentMethod } = await stripe.createPaymentMethod({
          type: "card",
          card: cardElement,
        });

        if (error) {
          throw error;
        }

        // Call your backend to complete the payment
        // const response = await fetch('/api/payment', {
        //   method: 'POST',
        //   body: JSON.stringify({ paymentMethodId: paymentMethod.id }),
        // });

        setSucceeded(true);
        toast({
          title: "Payment successful!",
          description: `You've purchased ${modelName}`,
        });

        setTimeout(() => {
          onSuccess();
          onOpenChange(false);
          setSucceeded(false);
        }, 1500);
      }
    } catch (error: any) {
      console.error("Payment error:", error);
      toast({
        title: "Payment failed",
        description: error.message || "Something went wrong with your payment",
        variant: "destructive",
      });
    } finally {
      setProcessing(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px] bg-gradient-card border-border">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-2xl">
            <CreditCard className="w-6 h-6 text-primary" />
            Complete Purchase
          </DialogTitle>
          <DialogDescription>
            Secure checkout powered by Stripe
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Order Summary */}
          <Card className="bg-card/50 border-border/50">
            <CardContent className="p-4 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Model</span>
                <span className="font-semibold">{modelName}</span>
              </div>
              <Separator />
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Price</span>
                <span className="font-semibold">${priceInDollars}</span>
              </div>
              <Separator />
              <div className="flex items-center justify-between text-lg">
                <span className="font-semibold">Total</span>
                <span className="font-bold text-primary">${priceInDollars}</span>
              </div>
            </CardContent>
          </Card>

          {/* Payment Form */}
          {!succeeded && (
            <>
              <div className="space-y-3">
                <label className="text-sm font-medium flex items-center gap-2">
                  <CreditCard className="w-4 h-4" />
                  Card Information
                </label>
                <div className="p-4 border border-border rounded-lg bg-card/30 focus-within:ring-2 focus-within:ring-primary/50 transition-all">
                  {mockMode ? (
                    <div className="space-y-2 text-sm text-muted-foreground">
                      <p className="flex items-center gap-2">
                        <ShieldCheck className="w-4 h-4 text-green-500" />
                        Mock Payment Mode (Development)
                      </p>
                      <p className="text-xs">
                        Click "Pay Now" to simulate a successful payment. No real payment will be processed.
                      </p>
                    </div>
                  ) : (
                    <CardElement
                      options={{
                        style: {
                          base: {
                            fontSize: "16px",
                            color: "hsl(0 0% 98%)",
                            "::placeholder": {
                              color: "hsl(0 0% 60%)",
                            },
                          },
                          invalid: {
                            color: "hsl(0 70% 55%)",
                          },
                        },
                        hidePostalCode: false,
                      }}
                    />
                  )}
                </div>
              </div>

              {/* Security Notice */}
              <div className="flex items-start gap-2 p-3 bg-primary/10 border border-primary/20 rounded-lg">
                <Lock className="w-4 h-4 text-primary mt-0.5 flex-shrink-0" />
                <div className="text-xs text-muted-foreground">
                  <p className="font-medium text-foreground mb-1">
                    Your payment is secure
                  </p>
                  <p>
                    We use industry-standard encryption to protect your payment information.
                    Your card details are never stored on our servers.
                  </p>
                </div>
              </div>
            </>
          )}

          {/* Success State */}
          {succeeded && (
            <div className="flex flex-col items-center justify-center py-8 space-y-4">
              <div className="w-16 h-16 rounded-full bg-green-500/20 flex items-center justify-center">
                <CheckCircle2 className="w-10 h-10 text-green-500" />
              </div>
              <div className="text-center">
                <h3 className="text-xl font-semibold mb-1">Payment Successful!</h3>
                <p className="text-sm text-muted-foreground">
                  Starting your download...
                </p>
              </div>
            </div>
          )}

          <DialogFooter className="gap-2">
            {!succeeded && (
              <>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => onOpenChange(false)}
                  disabled={processing}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  disabled={processing || !stripe}
                  className="bg-gradient-primary hover:opacity-90"
                >
                  {processing ? (
                    <>
                      <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                      Processing...
                    </>
                  ) : (
                    <>
                      <Lock className="w-4 h-4 mr-2" />
                      Pay ${priceInDollars}
                    </>
                  )}
                </Button>
              </>
            )}
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};

export default StripeCheckout;
