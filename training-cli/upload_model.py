#!/usr/bin/env python3
"""
Upload trained model to the platform
"""
import argparse
import requests
import os
from pathlib import Path

def upload_model(args):
    """Upload trained model to server"""
    print("="*60)
    print("📤 Uploading Model to Platform")
    print("="*60)

    model_path = Path(args.model_path)
    if not model_path.exists():
        print(f"❌ Error: Model file not found at {model_path}")
        return False

    # Prepare files for upload
    files = {
        'model_file': open(model_path, 'rb'),
    }

    # Check for training log
    log_path = model_path.parent / "training_log.json"
    if log_path.exists():
        files['training_log'] = open(log_path, 'rb')
        print("✅ Including training log")

    data = {
        'model_name': args.model_name or model_path.stem,
        'description': args.description or '',
    }

    headers = {
        'Authorization': f'Bearer {args.api_key}'
    }

    print(f"📡 Uploading to: {args.server_url}/v1/insert")
    print(f"📦 Model size: {model_path.stat().st_size / (1024*1024):.2f} MB")

    try:
        response = requests.post(
            f"{args.server_url}/v1/insert",
            files=files,
            data=data,
            headers=headers
        )

        if response.status_code == 200:
            print("✅ Upload successful!")
            print(f"🎉 Your model is now available on the platform")
            return True
        else:
            print(f"❌ Upload failed: {response.status_code}")
            print(f"   {response.text}")
            return False

    except Exception as e:
        print(f"❌ Upload error: {str(e)}")
        return False
    finally:
        # Close files
        for f in files.values():
            f.close()

def main():
    parser = argparse.ArgumentParser(
        description='Upload trained model to the platform'
    )
    parser.add_argument('--model-path', type=str, required=True,
                        help='Path to the trained model file')
    parser.add_argument('--api-key', type=str, required=True,
                        help='Your API key from the platform')
    parser.add_argument('--server-url', type=str,
                        default='http://localhost:8081',
                        help='Server URL (default: http://localhost:8081)')
    parser.add_argument('--model-name', type=str,
                        help='Model name (default: filename)')
    parser.add_argument('--description', type=str, default='',
                        help='Model description')

    args = parser.parse_args()

    if upload_model(args):
        print("\n✨ Done!")
    else:
        print("\n💡 Tip: Make sure you're logged in and have the correct API key")
        exit(1)

if __name__ == "__main__":
    main()
