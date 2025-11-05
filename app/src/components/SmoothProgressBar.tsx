import { useEffect, useState, useRef } from "react";

interface SmoothProgressBarProps {
  targetProgress: number; // 0-100
  className?: string;
  showPercentage?: boolean;
  animationSpeed?: number; // milliseconds for smooth transition
}

export const SmoothProgressBar = ({
  targetProgress,
  className = "",
  showPercentage = true,
  animationSpeed = 300,
}: SmoothProgressBarProps) => {
  const [currentProgress, setCurrentProgress] = useState(0);
  const [displayProgress, setDisplayProgress] = useState(0);
  const animationFrameRef = useRef<number>();

  // Smooth interpolation function (easeOutQuad)
  const easeOutQuad = (t: number) => t * (2 - t);

  useEffect(() => {
    const start = currentProgress;
    const end = targetProgress;
    const startTime = Date.now();
    const duration = animationSpeed;

    const animate = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);
      const easedProgress = easeOutQuad(progress);

      const interpolated = start + (end - start) * easedProgress;
      setDisplayProgress(interpolated);

      if (progress < 1) {
        animationFrameRef.current = requestAnimationFrame(animate);
      } else {
        setCurrentProgress(end);
        setDisplayProgress(end);
      }
    };

    // Cancel any ongoing animation
    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
    }

    // Start animation
    animationFrameRef.current = requestAnimationFrame(animate);

    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
    };
  }, [targetProgress, animationSpeed]);

  return (
    <div className={`space-y-2 ${className}`}>
      <div className="w-full bg-muted rounded-full h-4 overflow-hidden relative shadow-inner">
        {/* Animated progress bar */}
        <div
          className="bg-gradient-to-r from-primary via-secondary to-primary h-full transition-all relative animate-pulse-glow"
          style={{
            width: `${displayProgress}%`,
            transition: `width ${animationSpeed}ms cubic-bezier(0.4, 0.0, 0.2, 1)`,
          }}
        >
          {/* Shimmer effect */}
          <div
            className="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent animate-shimmer"
            style={{
              backgroundSize: '200% 100%',
            }}
          />
        </div>

        {/* Progress indicator line (moving dot) */}
        {displayProgress > 0 && displayProgress < 100 && (
          <div
            className="absolute top-0 h-full w-1 bg-white shadow-lg transition-all"
            style={{
              left: `${displayProgress}%`,
              transition: `left ${animationSpeed}ms cubic-bezier(0.4, 0.0, 0.2, 1)`,
            }}
          />
        )}
      </div>

      {showPercentage && (
        <div className="flex justify-between items-center text-sm">
          <p className="text-muted-foreground">
            Training Progress
          </p>
          <p className="text-primary font-semibold tabular-nums">
            {displayProgress.toFixed(1)}%
          </p>
        </div>
      )}
    </div>
  );
};
