#!/usr/bin/env python3
"""
Simple MNIST-like training script for development/testing
This script trains a CNN on synthetic or MNIST data and reports accuracy
"""
import os
import sys
import json
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, TensorDataset
import numpy as np
from model import SimpleNet

def create_synthetic_dataset(num_samples=1000, num_classes=10):
    """
    Create synthetic dataset for quick testing
    Returns: train and test datasets
    """
    print("📊 Creating synthetic dataset...")

    # Generate random images (28x28 grayscale)
    train_images = torch.randn(num_samples, 1, 28, 28)
    test_images = torch.randn(num_samples // 5, 1, 28, 28)

    # Generate random labels
    train_labels = torch.randint(0, num_classes, (num_samples,))
    test_labels = torch.randint(0, num_classes, (num_samples // 5,))

    # Create datasets
    train_dataset = TensorDataset(train_images, train_labels)
    test_dataset = TensorDataset(test_images, test_labels)

    print(f"✅ Created {len(train_dataset)} training samples and {len(test_dataset)} test samples")

    return train_dataset, test_dataset

def try_load_mnist():
    """
    Try to load MNIST dataset if available
    Returns: train and test datasets, or None if unavailable
    """
    try:
        from torchvision import datasets, transforms

        print("📥 Attempting to download MNIST dataset...")

        transform = transforms.Compose([
            transforms.ToTensor(),
            transforms.Normalize((0.1307,), (0.3081,))
        ])

        data_dir = os.path.join(os.path.dirname(__file__), 'data')
        os.makedirs(data_dir, exist_ok=True)

        train_dataset = datasets.MNIST(
            data_dir,
            train=True,
            download=True,
            transform=transform
        )

        test_dataset = datasets.MNIST(
            data_dir,
            train=False,
            download=True,
            transform=transform
        )

        print(f"✅ MNIST loaded: {len(train_dataset)} training samples")
        return train_dataset, test_dataset

    except Exception as e:
        print(f"⚠️  Could not load MNIST: {e}")
        return None, None

def train_epoch(model, train_loader, criterion, optimizer, device, epoch):
    """Train for one epoch"""
    model.train()
    total_loss = 0
    correct = 0
    total = 0

    for batch_idx, (data, target) in enumerate(train_loader):
        data, target = data.to(device), target.to(device)

        optimizer.zero_grad()
        output = model(data)
        loss = criterion(output, target)
        loss.backward()
        optimizer.step()

        total_loss += loss.item()
        pred = output.argmax(dim=1, keepdim=True)
        correct += pred.eq(target.view_as(pred)).sum().item()
        total += target.size(0)

        if batch_idx % 10 == 0:
            progress = {
                "epoch": epoch + 1,
                "batch": batch_idx,
                "loss": loss.item(),
                "accuracy": 100. * correct / total,
                "status": "training"
            }
            print(f"PROGRESS: {json.dumps(progress)}")

    avg_loss = total_loss / len(train_loader)
    accuracy = 100. * correct / total

    return avg_loss, accuracy

def evaluate(model, test_loader, criterion, device):
    """Evaluate model on test set"""
    model.eval()
    test_loss = 0
    correct = 0
    total = 0

    with torch.no_grad():
        for data, target in test_loader:
            data, target = data.to(device), target.to(device)
            output = model(data)
            test_loss += criterion(output, target).item()
            pred = output.argmax(dim=1, keepdim=True)
            correct += pred.eq(target.view_as(pred)).sum().item()
            total += target.size(0)

    test_loss /= len(test_loader)
    accuracy = 100. * correct / total

    return test_loss, accuracy

def main():
    """Main training function"""
    print("🚀 Starting training...")

    # Hyperparameters
    batch_size = 64
    epochs = 5
    learning_rate = 0.001
    num_classes = 10

    # Device setup
    device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
    print(f"🖥️  Using device: {device}")

    # Load dataset (try MNIST first, fall back to synthetic)
    train_dataset, test_dataset = try_load_mnist()

    if train_dataset is None:
        print("📊 Using synthetic dataset for testing...")
        train_dataset, test_dataset = create_synthetic_dataset()

    # Create data loaders
    train_loader = DataLoader(train_dataset, batch_size=batch_size, shuffle=True)
    test_loader = DataLoader(test_dataset, batch_size=batch_size, shuffle=False)

    # Initialize model
    model = SimpleNet(num_classes=num_classes).to(device)
    criterion = nn.CrossEntropyLoss()
    optimizer = optim.Adam(model.parameters(), lr=learning_rate)

    print(f"📋 Model parameters: {sum(p.numel() for p in model.parameters()):,}")

    # Training loop
    best_accuracy = 0
    for epoch in range(epochs):
        print(f"\n📚 Epoch {epoch + 1}/{epochs}")

        # Train
        train_loss, train_acc = train_epoch(
            model, train_loader, criterion, optimizer, device, epoch
        )

        # Evaluate
        test_loss, test_acc = evaluate(model, test_loader, criterion, device)

        print(f"Train Loss: {train_loss:.4f} | Train Acc: {train_acc:.2f}%")
        print(f"Test Loss: {test_loss:.4f} | Test Acc: {test_acc:.2f}%")

        # Report progress
        progress = {
            "epoch": epoch + 1,
            "total_epochs": epochs,
            "train_loss": train_loss,
            "train_accuracy": train_acc,
            "test_loss": test_loss,
            "test_accuracy": test_acc,
            "status": "training"
        }
        print(f"PROGRESS: {json.dumps(progress)}")

        # Save best model
        if test_acc > best_accuracy:
            best_accuracy = test_acc
            model_path = os.path.join(os.path.dirname(__file__), 'models', 'best_model.pth')
            os.makedirs(os.path.dirname(model_path), exist_ok=True)
            torch.save({
                'epoch': epoch,
                'model_state_dict': model.state_dict(),
                'optimizer_state_dict': optimizer.state_dict(),
                'accuracy': best_accuracy,
                'train_loss': train_loss,
                'test_loss': test_loss,
            }, model_path)
            print(f"💾 Saved best model with accuracy: {best_accuracy:.2f}%")

    # Final results
    print(f"\n✅ Training complete!")
    print(f"🎯 Best test accuracy: {best_accuracy:.2f}%")

    # Save final results in format expected by training handler
    results = {
        "status": "completed",
        "accuracy": best_accuracy / 100.0,  # Convert to 0-1 range
        "final_epoch": epochs,
        "model_path": "models/best_model.pth"
    }

    results_path = os.path.join(os.path.dirname(__file__), 'training_results.json')
    with open(results_path, 'w') as f:
        json.dump(results, f, indent=2)

    print(f"📄 Results saved to {results_path}")

    # Print final progress for training handler
    final_progress = {
        "epoch": epochs,
        "total_epochs": epochs,
        "accuracy": best_accuracy,
        "loss": test_loss,
        "status": "completed"
    }
    print(f"PROGRESS: {json.dumps(final_progress)}")

if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"❌ Training failed: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()
        sys.exit(1)
