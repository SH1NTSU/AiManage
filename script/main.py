import os
import subprocess
import json
import sys
from datetime import datetime

UPLOADS_DIR = "../server/uploads/"
RESULTS_FILE = "./training_results.json"

def run_training(model_name):
    model_dir = os.path.join(UPLOADS_DIR, model_name)
    train_file = os.path.join(model_dir, "train.py")

    if not os.path.exists(model_dir):
        print(f"❌ Model '{model_name}' does not exist in {UPLOADS_DIR}")
        return None
    if not os.path.exists(train_file):
        print(f"❌ train.py not found in {model_dir}")
        return None

    log_file = os.path.join(model_dir, "train.log")
    print(f"▶ Running training for model: {model_name}")

    # Run training
    with open(log_file, "w") as log:
        process = subprocess.Popen(
            ["python3", train_file],
            cwd=model_dir,
            stdout=log,
            stderr=log
        )
        process.wait()

    # Collect stats
    stats_path = os.path.join(model_dir, "stats.json")
    stats = {}
    if os.path.exists(stats_path):
        with open(stats_path, "r") as f:
            stats = json.load(f)

    # Detect trained model file
    model_file = None
    for fname in os.listdir(model_dir):
        if fname.endswith((".pt", ".h5", ".pkl")):
            model_file = os.path.join(model_dir, fname)
            break

    result = {
        "model_name": model_name,
        "train_file": train_file,
        "stats": stats,
        "trained_model": model_file,
        "log_file": log_file,
        "finished_at": datetime.now().isoformat()
    }

    # Save result to JSON
    with open(RESULTS_FILE, "w") as f:
        json.dump(result, f, indent=4)

    print(f"✅ Training complete for '{model_name}'. Results saved in {RESULTS_FILE}")
    return result

def main():
    if len(sys.argv) < 2:
        print("❌ Usage: python3 train_one.py <ModelName>")
        sys.exit(1)

    model_name = sys.argv[1]
    run_training(model_name)

if __name__ == "__main__":
    main()
