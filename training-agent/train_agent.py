#!/usr/bin/env python3
"""
AI Training Agent - Runs on user's machine and connects to server
This allows users to train models using their own compute resources
"""
import asyncio
import websockets
import json
import subprocess
import os
import sys
from pathlib import Path
import argparse
import torch
import time
import aiohttp

class TrainingAgent:
    def __init__(self, api_key: str, server_url: str = "ws://localhost:8081"):
        self.api_key = api_key
        self.server_url = server_url.replace("http://", "ws://").replace("https://", "wss://")
        self.websocket = None
        self.is_training = False
        self.current_process = None

    async def connect(self):
        """Connect to the server via WebSocket"""
        print("🔌 Connecting to server...")
        print(f"   Server: {self.server_url}")

        try:
            uri = f"{self.server_url}/v1/ws/agent?api_key={self.api_key}"
            self.websocket = await websockets.connect(uri)
            print("✅ WebSocket connection established!")

            # Wait for welcome message to confirm server accepted connection
            try:
                welcome = await asyncio.wait_for(self.websocket.recv(), timeout=5.0)
                welcome_data = json.loads(welcome)
                if welcome_data.get("type") == "connected":
                    print("✅ Server accepted connection!")
                    print(f"   Message: {welcome_data.get('message', 'N/A')}")
                    print("📡 Waiting for training jobs...")
                    return True
                else:
                    print("⚠️  Unexpected welcome message:", welcome_data)
                    return True  # Continue anyway
            except asyncio.TimeoutError:
                print("⚠️  No welcome message received, but connection seems active")
                return True

        except websockets.exceptions.InvalidStatusCode as e:
            print(f"❌ Connection rejected by server: HTTP {e.status_code}")
            if e.status_code == 401:
                print("   Reason: Invalid or missing API key")
                print("   Please check your API key and try again")
            elif e.status_code == 500:
                print("   Reason: Server error")
                print("   The server might be having issues or the database is down")
            return False
        except ConnectionRefusedError:
            print(f"❌ Connection refused - is the server running at {self.server_url}?")
            return False
        except Exception as e:
            print(f"❌ Connection failed: {type(e).__name__}: {str(e)}")
            return False

    async def listen(self):
        """Listen for training commands from server"""
        try:
            async for message in self.websocket:
                data = json.loads(message)
                await self.handle_message(data)
        except websockets.exceptions.ConnectionClosed as e:
            print(f"⚠️  Connection closed by server (code: {e.code}, reason: {e.reason})")
        except json.JSONDecodeError as e:
            print(f"❌ Failed to parse message from server: {str(e)}")
        except Exception as e:
            print(f"❌ Error in message handling: {type(e).__name__}: {str(e)}")

    async def handle_message(self, data: dict):
        """Handle messages from server"""
        msg_type = data.get("type")

        if msg_type == "ping":
            await self.send_message({"type": "pong"})

        elif msg_type == "system_info_request":
            print("📊 Server requesting system information...")
            info = self.get_system_info()
            await self.send_message({
                "type": "system_info",
                "data": info
            })
            print("✅ System information sent to server")

        elif msg_type == "train":
            await self.handle_training(data.get("data", {}))

        elif msg_type == "stop":
            await self.stop_training()

        elif msg_type == "connected":
            # Already handled in connect(), but just in case
            pass

        else:
            print(f"⚠️  Unknown message type from server: {msg_type}")

    def get_system_info(self):
        """Get system information"""
        return {
            "python_version": sys.version,
            "pytorch_version": torch.__version__,
            "cuda_available": torch.cuda.is_available(),
            "gpu_count": torch.cuda.device_count() if torch.cuda.is_available() else 0,
            "gpu_name": torch.cuda.get_device_name(0) if torch.cuda.is_available() else "None",
            "platform": sys.platform,
        }

    async def handle_training(self, train_data: dict):
        """Handle training request"""
        if self.is_training:
            await self.send_message({
                "type": "error",
                "message": "Already training a model"
            })
            return

        print("\n" + "="*60)
        print("🚀 New Training Job Received!")
        print("="*60)

        training_id = train_data.get("training_id")
        folder_path = train_data.get("folder_path")
        script_name = train_data.get("script_name", "train.py")
        python_cmd = train_data.get("python_command", "python3")

        print(f"📁 Folder: {folder_path}")
        print(f"📜 Script: {script_name}")
        print(f"🆔 Training ID: {training_id}")

        # Validate folder exists
        if not os.path.exists(folder_path):
            await self.send_message({
                "type": "error",
                "training_id": training_id,
                "message": f"Folder not found: {folder_path}"
            })
            return

        script_path = os.path.join(folder_path, script_name)
        if not os.path.exists(script_path):
            await self.send_message({
                "type": "error",
                "training_id": training_id,
                "message": f"Script not found: {script_path}"
            })
            return

        self.is_training = True

        try:
            await self.send_message({
                "type": "training_started",
                "training_id": training_id
            })

            # Capture file snapshot before training
            before_snapshot = self.capture_file_snapshot(folder_path)

            # Run training script
            success = await self.run_training_script(
                training_id,
                folder_path,
                script_path,
                python_cmd
            )

            # Detect trained model if training succeeded
            model_path = None
            if success:
                after_snapshot = self.capture_file_snapshot(folder_path)
                model_path = self.detect_trained_model(folder_path, before_snapshot, after_snapshot)
                if model_path:
                    print(f"💾 Detected trained model: {model_path}")

                    # Upload model file to server
                    full_model_path = os.path.join(folder_path, model_path)
                    server_path = await self.upload_model_to_server(
                        training_id,
                        full_model_path,
                        model_path
                    )

                    # Use server path if upload succeeded, otherwise use local path
                    if server_path:
                        model_path = server_path
                        print(f"✅ Model uploaded to server: {server_path}")

                # Send completion message with model path
                await self.send_message({
                    "type": "training_completed",
                    "training_id": training_id,
                    "model_path": model_path
                })

        except Exception as e:
            await self.send_message({
                "type": "training_failed",
                "training_id": training_id,
                "error": str(e)
            })
        finally:
            self.is_training = False
            self.current_process = None

    async def run_training_script(self, training_id, folder_path, script_path, python_cmd):
        """Run the training script and stream output"""
        print(f"\n🔄 Starting training...\n")

        try:
            # Start the training process
            process = subprocess.Popen(
                [python_cmd, script_path],
                cwd=folder_path,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=1
            )

            self.current_process = process

            # Stream output
            while True:
                # Check if process is still running
                if process.poll() is not None:
                    break

                # Read stdout
                output = process.stdout.readline()
                if output:
                    print(output.strip())
                    await self.send_message({
                        "type": "training_output",
                        "training_id": training_id,
                        "output": output.strip()
                    })

                await asyncio.sleep(0.1)

            # Get final return code
            return_code = process.poll()

            if return_code == 0:
                print("\n✅ Training completed successfully!")
                return True  # Success
            else:
                stderr = process.stderr.read()
                print(f"\n❌ Training failed with code {return_code}")
                print(stderr)
                await self.send_message({
                    "type": "training_failed",
                    "training_id": training_id,
                    "error": stderr
                })
                return False  # Failed

        except Exception as e:
            print(f"\n❌ Error running training: {str(e)}")
            await self.send_message({
                "type": "training_failed",
                "training_id": training_id,
                "error": str(e)
            })
            return False  # Failed

    async def stop_training(self):
        """Stop current training"""
        if self.current_process:
            print("\n⚠️  Stopping training...")
            self.current_process.terminate()
            self.current_process = None
            self.is_training = False
            print("✅ Training stopped")

    async def send_message(self, data: dict):
        """Send message to server"""
        if self.websocket:
            await self.websocket.send(json.dumps(data))

    def capture_file_snapshot(self, folder_path):
        """Capture snapshot of all files in directory"""
        snapshot = {}
        for root, dirs, files in os.walk(folder_path):
            for file in files:
                file_path = os.path.join(root, file)
                try:
                    stat = os.stat(file_path)
                    snapshot[file_path] = {
                        'mtime': stat.st_mtime,
                        'size': stat.st_size
                    }
                except:
                    pass
        print(f"📸 Captured snapshot of {len(snapshot)} files")
        return snapshot

    def detect_trained_model(self, folder_path, before, after):
        """Detect new or modified model files"""
        model_extensions = [
            '.pth', '.pt',           # PyTorch
            '.h5', '.keras',         # TensorFlow/Keras
            '.pkl', '.pickle',       # scikit-learn
            '.ckpt',                 # TensorFlow checkpoints
            '.pb',                   # TensorFlow protobuf
            '.onnx',                 # ONNX
            '.safetensors',          # Hugging Face
            '.joblib',               # scikit-learn
            '.model',                # Generic
        ]

        changed_models = []

        for file_path, after_info in after.items():
            # Check if it's a model file
            if not any(file_path.endswith(ext) for ext in model_extensions):
                continue

            # New file or modified file
            before_info = before.get(file_path)
            if not before_info:
                changed_models.append(file_path)
                print(f"🆕 New model file: {os.path.basename(file_path)}")
            elif after_info['mtime'] > before_info['mtime'] or after_info['size'] != before_info['size']:
                changed_models.append(file_path)
                print(f"♻️  Modified model file: {os.path.basename(file_path)}")

        if not changed_models:
            print("ℹ️  No model files detected")
            return None

        # Select the best model
        return self.select_best_model(changed_models, folder_path)

    def select_best_model(self, models, folder_path):
        """Select the most likely final model from candidates"""
        if len(models) == 1:
            return os.path.relpath(models[0], folder_path)

        print(f"🤔 Multiple models detected, selecting best one...")

        # Priority 1: Keywords in filename
        for model in models:
            basename = os.path.basename(model).lower()
            if any(keyword in basename for keyword in ['best', 'final', 'trained']):
                print(f"✨ Selected by keyword: {os.path.basename(model)}")
                return os.path.relpath(model, folder_path)

        # Priority 2: Standard output directories
        for model in models:
            if any(dir_name in model for dir_name in ['saved_models', 'outputs', 'checkpoints', 'models']):
                print(f"📁 Selected from standard directory: {os.path.basename(model)}")
                return os.path.relpath(model, folder_path)

        # Priority 3: Largest file
        largest = max(models, key=lambda f: os.path.getsize(f))
        size_mb = os.path.getsize(largest) / (1024 * 1024)
        print(f"📏 Selected largest file: {os.path.basename(largest)} ({size_mb:.2f} MB)")
        return os.path.relpath(largest, folder_path)

    async def upload_model_to_server(self, training_id, file_path, original_path):
        """Upload trained model file to server"""
        try:
            # Extract model name from training ID (format: "ModelName_timestamp")
            model_name = training_id.split('_')[0] if '_' in training_id else training_id

            print(f"\n📤 Uploading model to server...")
            print(f"   Model: {model_name}")
            print(f"   File: {file_path}")

            # Convert WebSocket URL to HTTP URL
            http_url = self.server_url.replace('ws://', 'http://').replace('wss://', 'https://')
            upload_url = f"{http_url}/v1/agent/upload-model"

            # Get file size
            file_size_mb = os.path.getsize(file_path) / (1024 * 1024)
            print(f"   Size: {file_size_mb:.2f} MB")

            # Prepare form data
            data = aiohttp.FormData()
            data.add_field('model_name', model_name)
            data.add_field('original_path', original_path)
            data.add_field('model_file',
                          open(file_path, 'rb'),
                          filename=os.path.basename(file_path))

            # Upload with progress
            async with aiohttp.ClientSession() as session:
                headers = {'Authorization': f'Bearer {self.api_key}'}
                async with session.post(upload_url, data=data, headers=headers) as response:
                    if response.status == 200:
                        result = await response.json()
                        server_path = result.get('server_path')
                        print(f"✅ Upload successful!")
                        return server_path
                    else:
                        error_text = await response.text()
                        print(f"❌ Upload failed: {response.status} - {error_text}")
                        return None

        except Exception as e:
            print(f"❌ Error uploading model: {str(e)}")
            return None

    async def run(self):
        """Main run loop"""
        while True:
            if await self.connect():
                try:
                    await self.listen()
                except Exception as e:
                    print(f"❌ Error: {str(e)}")

            print("🔄 Reconnecting in 5 seconds...")
            await asyncio.sleep(5)

def main():
    parser = argparse.ArgumentParser(
        description='Training Agent - Connects your machine to the platform'
    )
    parser.add_argument('--api-key', type=str, required=True,
                        help='Your API key from the platform')
    parser.add_argument('--server-url', type=str,
                        default='ws://localhost:8081',
                        help='Server URL (default: ws://localhost:8081)')

    args = parser.parse_args()

    print("="*60)
    print("🤖 AI Training Agent")
    print("="*60)
    print("This agent allows you to train models using your own")
    print("computer's resources while managing them from the platform.")
    print("="*60)

    # Check PyTorch
    if torch.cuda.is_available():
        print(f"✅ GPU Available: {torch.cuda.get_device_name(0)}")
    else:
        print("ℹ️  No GPU detected - will use CPU")

    print("\n")

    agent = TrainingAgent(args.api_key, args.server_url)

    try:
        asyncio.run(agent.run())
    except KeyboardInterrupt:
        print("\n\n👋 Shutting down agent...")
        sys.exit(0)

if __name__ == "__main__":
    main()
