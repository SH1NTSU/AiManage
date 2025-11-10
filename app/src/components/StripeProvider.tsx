import { ReactNode } from "react";
import { Elements } from "@stripe/react-stripe-js";
import { loadStripe } from "@stripe/stripe-js";

// For mock mode, we'll use a test publishable key
// In production, replace this with your actual Stripe publishable key
const stripePromise = loadStripe(
  import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY ||
  "pk_test_51QP000000000000000000000000000000000000000000000000000000000000"
);

interface StripeProviderProps {
  children: ReactNode;
}

const StripeProvider = ({ children }: StripeProviderProps) => {
  return (
    <Elements
      stripe={stripePromise}
      options={{
        appearance: {
          theme: 'night',
          variables: {
            colorPrimary: 'hsl(180, 80%, 50%)',
            colorBackground: 'hsl(0, 0%, 10%)',
            colorText: 'hsl(0, 0%, 98%)',
            colorDanger: 'hsl(0, 70%, 55%)',
            fontFamily: 'system-ui, sans-serif',
            spacingUnit: '4px',
            borderRadius: '8px',
          },
        },
      }}
    >
      {children}
    </Elements>
  );
};

export default StripeProvider;
