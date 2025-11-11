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
            print("✅ Connected to server!")
            print("📡 Waiting for training jobs...")
            return True
        except Exception as e:
            print(f"❌ Connection failed: {str(e)}")
            return False

    async def listen(self):
        """Listen for training commands from server"""
        try:
            async for message in self.websocket:
                data = json.loads(message)
                await self.handle_message(data)
        except websockets.exceptions.ConnectionClosed:
            print("⚠️  Connection closed by server")
        except Exception as e:
            print(f"❌ Error: {str(e)}")

    async def handle_message(self, data: dict):
        """Handle messages from server"""
        msg_type = data.get("type")

        if msg_type == "ping":
            await self.send_message({"type": "pong"})

        elif msg_type == "system_info_request":
            info = self.get_system_info()
            await self.send_message({
                "type": "system_info",
                "data": info
            })

        elif msg_type == "train":
            await self.handle_training(data.get("data", {}))

        elif msg_type == "stop":
            await self.stop_training()

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

            # Run training script
            await self.run_training_script(
                training_id,
                folder_path,
                script_path,
                python_cmd
            )

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
                await self.send_message({
                    "type": "training_completed",
                    "training_id": training_id
                })
            else:
                stderr = process.stderr.read()
                print(f"\n❌ Training failed with code {return_code}")
                print(stderr)
                await self.send_message({
                    "type": "training_failed",
                    "training_id": training_id,
                    "error": stderr
                })

        except Exception as e:
            print(f"\n❌ Error running training: {str(e)}")
            await self.send_message({
                "type": "training_failed",
                "training_id": training_id,
                "error": str(e)
            })

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
