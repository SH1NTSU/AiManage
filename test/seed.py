
from pymongo import MongoClient

# Connect to MongoDB
client = MongoClient("mongodb://localhost:27017")  # Adjust if needed
db = client["AiManage"]  # Database name
models_collection = db["Models"]  # Collection name

mock_models = [
    {
        "name": "GPT-4 Turbo",
        "picture": "gpt4.png",
        "folder": ["model.bin", "config.json"],
    },
    {
        "name": "Claude 3 Opus",
        "picture": "claude3.png",
        "folder": ["model.bin", "config.json"],
    },
    {
        "name": "Llama 3 70B",
        "picture": "llama3.png",
        "folder": ["weights.bin"],
    },
]

# Optional: clear old data first
models_collection.delete_many({})

# Insert mock data
result = models_collection.insert_many(mock_models)

print(f"Inserted {len(result.inserted_ids)} mock models into MongoDB.")

# Close connection
client.close()
