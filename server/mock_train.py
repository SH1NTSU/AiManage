#!/usr/bin/env python3
"""
Mock training script for testing the AI Agent metrics system.
Simulates a realistic training run without requiring PyTorch/data.
"""

import time
import random
import math

def simulate_training():
    """Simulate a training run with realistic metrics"""

    epochs = 10
    print(f"Starting mock training for {epochs} epochs...")
    print("=" * 60)

    # Initial metrics (poor performance)
    train_loss = 2.5
    val_loss = 2.7
    train_acc = 0.25
    val_acc = 0.22

    for epoch in range(1, epochs + 1):
        # Simulate epoch taking some time
        time.sleep(2)  # 2 seconds per epoch

        # Improve metrics over time (with some noise)
        improvement_factor = 1 - (epoch / epochs) * 0.7
        noise = random.uniform(-0.05, 0.05)

        train_loss = train_loss * improvement_factor + noise * 0.3
        val_loss = val_loss * improvement_factor + noise * 0.4  # Val loss improves slower

        train_acc = min(0.95, train_acc + (0.95 - train_acc) * 0.25 + random.uniform(-0.02, 0.03))
        val_acc = min(0.90, val_acc + (0.90 - val_acc) * 0.22 + random.uniform(-0.02, 0.02))

        # Ensure val_loss > train_loss (realistic overfitting)
        if val_loss < train_loss:
            val_loss = train_loss * 1.15

        # Print in the format that the agent expects
        print(f"Epoch {epoch}/{epochs}, Train Loss: {train_loss:.4f}, Val Loss: {val_loss:.4f}, Train Accuracy: {train_acc*100:.2f}%, Val Accuracy: {val_acc*100:.2f}%")

    print("=" * 60)
    print("Training completed!")
    print(f"Final Train Accuracy: {train_acc*100:.2f}%")
    print(f"Final Val Accuracy: {val_acc*100:.2f}%")
    print(f"Final Train Loss: {train_loss:.4f}")
    print(f"Final Val Loss: {val_loss:.4f}")

if __name__ == "__main__":
    simulate_training()
