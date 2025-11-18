# Training Script Format Guide

This document explains how to format your training scripts to work seamlessly with the AiManage platform, including proper progress reporting and accuracy tracking.

## Overview

AiManage tracks your training progress in real-time by parsing `PROGRESS` JSON messages that your training script prints to stdout. This allows the platform to:

- Display training metrics in real-time
- Track accuracy and loss over epochs
- Save the final accuracy to the database
- Detect when training is complete

## PROGRESS JSON Format

### During Training (Per Epoch)

Print a JSON object prefixed with `PROGRESS:` after each epoch:

```python
import json

progress = {
    "epoch": current_epoch,           # Current epoch number (1-indexed)
    "total_epochs": total_epochs,     # Total number of epochs
    "train_loss": train_loss,         # Training loss (float)
    "train_accuracy": train_acc,      # Training accuracy in PERCENTAGE (e.g., 95.50 for 95.5%)
    "test_loss": test_loss,           # Test/validation loss (float)
    "test_accuracy": test_acc,        # Test accuracy in PERCENTAGE (e.g., 92.30 for 92.3%)
    "status": "training"              # Status: "training" or "completed"
}

print(f"PROGRESS: {json.dumps(progress)}")
```

### Final Progress (After Training Completes)

After training completes, print a final PROGRESS message with `status: "completed"`:

```python
final_progress = {
    "epoch": total_epochs,            # Final epoch number
    "total_epochs": total_epochs,     # Total number of epochs
    "test_accuracy": best_accuracy,   # IMPORTANT: Use "test_accuracy" field
    "train_loss": final_train_loss,   # Optional: final training loss
    "test_loss": final_test_loss,     # Optional: final test loss
    "status": "completed"             # IMPORTANT: Set status to "completed"
}

print(f"PROGRESS: {json.dumps(final_progress)}")
```

## Field Specifications

### Required Fields

| Field | Type | Description | Format |
|-------|------|-------------|--------|
| `epoch` | int | Current epoch number | 1-indexed (1, 2, 3, ...) |
| `total_epochs` | int | Total epochs to train | Positive integer |
| `status` | string | Training status | `"training"` or `"completed"` |

### Accuracy Fields

**IMPORTANT:** Accuracy values should be in **percentage format** (e.g., 95.50 for 95.5%).

| Field | Type | Description | Priority |
|-------|------|-------------|----------|
| `test_accuracy` | float | Test/validation accuracy in % | **Highest** (recommended for final accuracy) |
| `val_accuracy` | float | Validation accuracy in % | Medium |
| `train_accuracy` | float | Training accuracy in % | Lowest |
| `accuracy` | float | Generic accuracy in % | Falls back to test_accuracy |

**Priority for Database Storage:**
1. `test_accuracy` (preferred)
2. `val_accuracy`
3. `train_accuracy`

### Loss Fields

| Field | Type | Description |
|-------|------|-------------|
| `train_loss` | float | Training loss |
| `test_loss` | float | Test/validation loss |
| `val_loss` | float | Validation loss |
| `loss` | float | Generic loss (falls back to train_loss) |

## Common Mistakes

### ❌ Wrong: Using generic "accuracy" field in final message
```python
# This might not save accuracy correctly
final_progress = {
    "epoch": 10,
    "total_epochs": 10,
    "accuracy": best_accuracy,  # ❌ Ambiguous
    "status": "completed"
}
```

### ✅ Correct: Using "test_accuracy" field in final message
```python
# This will save accuracy correctly
final_progress = {
    "epoch": 10,
    "total_epochs": 10,
    "test_accuracy": best_accuracy,  # ✅ Explicit
    "status": "completed"
}
```

### ❌ Wrong: Accuracy in 0-1 range
```python
# This will be interpreted incorrectly
progress = {
    "test_accuracy": 0.9550  # ❌ Will be saved as 95.50% instead of 0.9550%
}
```

### ✅ Correct: Accuracy in percentage
```python
# Accuracy as percentage
progress = {
    "test_accuracy": 95.50  # ✅ Will be correctly saved as 95.50%
}
```

## Complete Example

Here's a complete PyTorch training script with proper progress reporting:

```python
#!/usr/bin/env python3
import json
import torch
import torch.nn as nn
import torch.optim as optim

def train_epoch(model, train_loader, criterion, optimizer, device):
    model.train()
    total_loss = 0
    correct = 0
    total = 0

    for data, target in train_loader:
        data, target = data.to(device), target.to(device)
        optimizer.zero_grad()
        output = model(data)
        loss = criterion(output, target)
        loss.backward()
        optimizer.step()

        total_loss += loss.item()
        pred = output.argmax(dim=1)
        correct += pred.eq(target).sum().item()
        total += target.size(0)

    avg_loss = total_loss / len(train_loader)
    accuracy = 100.0 * correct / total  # Percentage format
    return avg_loss, accuracy

def evaluate(model, test_loader, criterion, device):
    model.eval()
    test_loss = 0
    correct = 0
    total = 0

    with torch.no_grad():
        for data, target in test_loader:
            data, target = data.to(device), target.to(device)
            output = model(data)
            test_loss += criterion(output, target).item()
            pred = output.argmax(dim=1)
            correct += pred.eq(target).sum().item()
            total += target.size(0)

    test_loss /= len(test_loader)
    accuracy = 100.0 * correct / total  # Percentage format
    return test_loss, accuracy

def main():
    # Setup
    device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
    model = YourModel().to(device)
    criterion = nn.CrossEntropyLoss()
    optimizer = optim.Adam(model.parameters())

    # Training loop
    epochs = 10
    best_accuracy = 0
    final_train_loss = 0
    final_test_loss = 0

    for epoch in range(epochs):
        # Train and evaluate
        train_loss, train_acc = train_epoch(model, train_loader, criterion, optimizer, device)
        test_loss, test_acc = evaluate(model, test_loader, criterion, device)

        print(f"Epoch {epoch+1}/{epochs}")
        print(f"Train Loss: {train_loss:.4f} | Train Acc: {train_acc:.2f}%")
        print(f"Test Loss: {test_loss:.4f} | Test Acc: {test_acc:.2f}%")

        # Report progress to platform
        progress = {
            "epoch": epoch + 1,
            "total_epochs": epochs,
            "train_loss": train_loss,
            "train_accuracy": train_acc,  # In percentage
            "test_loss": test_loss,
            "test_accuracy": test_acc,    # In percentage
            "status": "training"
        }
        print(f"PROGRESS: {json.dumps(progress)}")

        # Save best model
        if test_acc > best_accuracy:
            best_accuracy = test_acc
            final_train_loss = train_loss
            final_test_loss = test_loss
            torch.save(model.state_dict(), 'models/best_model.pth')

    # Training complete - report final status
    print("\n✅ Training complete!")
    print(f"🎯 Best test accuracy: {best_accuracy:.2f}%")

    # IMPORTANT: Send final progress with status="completed"
    final_progress = {
        "epoch": epochs,
        "total_epochs": epochs,
        "test_accuracy": best_accuracy,  # Best accuracy in percentage
        "train_loss": final_train_loss,
        "test_loss": final_test_loss,
        "status": "completed"            # Mark as completed
    }
    print(f"PROGRESS: {json.dumps(final_progress)}")

if __name__ == "__main__":
    main()
```

## How Accuracy is Stored

1. **During training**, all PROGRESS messages are parsed and stored in memory
2. **When status="completed"**, the platform:
   - Looks for the message with `"status": "completed"`
   - Extracts accuracy from that message (preferring test_accuracy > val_accuracy > train_accuracy)
   - Stores the accuracy in the database
3. **Accuracy is stored as a percentage** in the database (e.g., 95.50 for 95.5%)

## Backward Compatibility

The platform also supports regex-based parsing for legacy scripts that don't use PROGRESS JSON:

```
Epoch 1/10, Train Loss: 0.5432, Train Accuracy: 85.5%
```

However, **JSON format is strongly recommended** for accurate and reliable tracking.

## Testing Your Script

To test if your script reports progress correctly:

1. Run your script locally:
   ```bash
   python train.py
   ```

2. Check that output includes lines like:
   ```
   PROGRESS: {"epoch": 1, "total_epochs": 10, "train_accuracy": 85.50, ...}
   ```

3. Verify the final line has `"status": "completed"`:
   ```
   PROGRESS: {"epoch": 10, "total_epochs": 10, "test_accuracy": 95.50, "status": "completed"}
   ```

4. Verify accuracy values are in percentage format (> 1.0)

## Need Help?

- See `demo_model/train.py` for a complete working example
- Check the TRAINING_GUIDE.md for general training information
- Open an issue on GitHub if you encounter problems

---

**Key Takeaways:**
- Always use `PROGRESS: {json}` format for reliable tracking
- Use `test_accuracy` field for final accuracy
- Report accuracy as **percentage** (e.g., 95.50, not 0.9550)
- Include `"status": "completed"` in the final PROGRESS message
- Track `final_train_loss` and `final_test_loss` to include in the final message
