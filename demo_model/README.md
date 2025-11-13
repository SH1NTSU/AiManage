# Demo PyTorch Model

A simple CNN for image classification, designed for development and testing purposes.

## Model Architecture

- **Input**: 28x28 grayscale images
- **Output**: 10 classes
- **Architecture**:
  - 2 Convolutional layers (32, 64 filters)
  - Batch Normalization
  - Max Pooling
  - 2 Fully Connected layers
  - Dropout for regularization

## Dataset

The training script will automatically try to:
1. Download MNIST dataset (if torchvision is available)
2. Fall back to synthetic data if MNIST download fails

## Training

```bash
python3 train.py
```

The script will:
- Train for 5 epochs
- Report progress after each batch and epoch
- Save the best model to `models/best_model.pth`
- Output final accuracy in the format expected by the training system

## Expected Accuracy

- **MNIST dataset**: 95-98% accuracy
- **Synthetic data**: 10-20% accuracy (random baseline)

## Files

- `model.py` - Model architecture
- `train.py` - Training script
- `requirements.txt` - Python dependencies
- `models/best_model.pth` - Saved model checkpoint
- `training_results.json` - Final training results
