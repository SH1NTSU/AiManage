# AI Model Training CLI

Train your AI models locally on your own machine!

## Installation

```bash
pip install -r requirements.txt
```

## Usage

```bash
# Train a model
python train_local.py --data ./your_data_folder --model-name my_model

# With custom parameters
python train_local.py \
  --data ./data \
  --model-name my_model \
  --epochs 10 \
  --batch-size 32 \
  --learning-rate 0.001
```

## Requirements

- Python 3.8+
- PyTorch
- Your training data folder with:
  - `train_data.txt` or `train.csv`
  - `test_data.txt` or `test.csv` (optional)

## Uploading Trained Models

After training, you can upload your trained model back to the platform:

```bash
python upload_model.py \
  --model-path ./models/my_model.pth \
  --api-key YOUR_API_KEY \
  --server-url https://your-platform.com
```

## Features

- ✅ Train models locally on your own hardware
- ✅ Full control over training parameters
- ✅ No server costs - use your own compute
- ✅ Upload trained models to share or deploy
- ✅ Track training progress in real-time
