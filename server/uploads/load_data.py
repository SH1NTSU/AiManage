import requests
import os

# Fetch Gen 1 cards
url = "https://api.pokemontcg.io/v2/cards?q=set.id:base1 OR set.id:jungle OR set.id:fossil OR set.id:team-rocket"
headers = {"X-Api-Key": "44e5c27d-331a-4f2e-bc8c-4369ce64b459"}  # If required
response = requests.get(url, headers=headers).json()

# Create a folder
os.makedirs("./pokemon_cards_gen1", exist_ok=True)

# Download each card
for card in response["data"]:
    img_url = card["images"]["large"]
    card_name = card["name"].replace(" ", "_").lower()
    set_name = card["set"]["name"].replace(" ", "_").lower()
    filename = f"pokemon_cards_gen1/{set_name}_{card_name}.png"
    
    img_data = requests.get(img_url).content
    with open(filename, "wb") as f:
        f.write(img_data)
    print(f"Downloaded: {filename}")
