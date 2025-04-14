import json
import os
from pathlib import Path

def save_models_to_individual_files():
    # Define file paths
    current_dir = Path(__file__).parent
    input_file = current_dir / "demo_models.json"
    output_dir = current_dir / "demos"
    
    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)
    
    # Read the demo_models.json file
    with open(input_file, 'r') as f:
        models = json.load(f)
    
    # Save each model to a separate file
    for i, model in enumerate(models):
        output_file = output_dir / f"demo_{i}.json"
        with open(output_file, 'w') as f:
            json.dump(model, f, indent=2)
        print(f"Saved model {i} to {output_file}")
    
    print(f"Successfully saved {len(models)} model files to the demos directory.")

if __name__ == "__main__":
    save_models_to_individual_files()