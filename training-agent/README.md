# Training Agent

Connect your computer to the platform and train models using your own resources!

## How It Works

1. **Run the agent** on your machine
2. **Keep it running** in the background
3. **Start training** from the web interface
4. **Training happens** on your machine
5. **Monitor progress** in real-time on the web

## Benefits

- ✅ **Free forever** - use your own compute
- ✅ **Your hardware** - use your GPU if you have one
- ✅ **Web interface** - manage training from anywhere
- ✅ **Real-time monitoring** - see training progress live
- ✅ **Privacy** - data never leaves your machine

## Installation

```bash
pip install websockets torch
```

## Usage

### 1. Get your API Key
- Log in to the platform
- Go to Settings > API Keys
- Copy your API key

### 2. Run the Agent

```bash
python train_agent.py --api-key YOUR_API_KEY
```

Or connect to custom server:

```bash
python train_agent.py \
  --api-key YOUR_API_KEY \
  --server-url ws://your-server.com
```

### 3. Start Training

- Go to the web interface
- Upload your model configuration
- **Specify the local folder path** on your machine
- Click "Start Training"
- The agent will receive the job and start training!

## Example

```bash
# Start the agent
python train_agent.py --api-key abc123xyz

# Output:
# 🤖 AI Training Agent
# ✅ GPU Available: NVIDIA GeForce RTX 3080
# 🔌 Connecting to server...
# ✅ Connected to server!
# 📡 Waiting for training jobs...
```

Then from the web interface:
1. Click "New Training"
2. Enter local path: `/home/user/my-model`
3. Click "Start"

The agent will train your model and stream results to the web!

## Requirements

- Python 3.8+
- PyTorch
- websockets library
- Internet connection (for server communication)

## Security

- Agent only accepts commands for folders you specify
- Uses API key authentication
- Can be stopped anytime (Ctrl+C)
- Never uploads your data to the server

## Troubleshooting

**Agent can't connect:**
- Check your API key is correct
- Make sure server URL is correct
- Check firewall/network settings

**Training fails:**
- Verify the folder path exists on your machine
- Check train.py exists in the folder
- Make sure you have necessary dependencies installed

**GPU not detected:**
- Install CUDA-enabled PyTorch
- Check NVIDIA drivers are installed

## Keep It Running

### Linux/Mac (using screen):
```bash
screen -S training-agent
python train_agent.py --api-key YOUR_KEY
# Press Ctrl+A then D to detach
```

### Windows (using pythonw):
```bash
pythonw train_agent.py --api-key YOUR_KEY
```

### As a Service (systemd):
Create `/etc/systemd/system/training-agent.service`:
```ini
[Unit]
Description=AI Training Agent
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/training-agent
ExecStart=/usr/bin/python3 train_agent.py --api-key YOUR_KEY
Restart=always

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl enable training-agent
sudo systemctl start training-agent
```

## FAQ

**Q: Does this upload my data?**
A: No! Training happens on your machine. Only training metrics are sent to the server.

**Q: Can I train multiple models?**
A: One at a time per agent. Run multiple agents on different ports for parallel training.

**Q: What if I close my laptop?**
A: Training will stop. Keep your machine awake or use a desktop/server.

**Q: Is this really free?**
A: Yes! You provide the compute, we provide the management interface.

**Q: Can I still use cloud training?**
A: Yes! Upgrade to a paid plan to train on our servers instead.
