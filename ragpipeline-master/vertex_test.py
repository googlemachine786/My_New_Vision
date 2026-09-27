"""Test Vertex AI embeddings with gcloud credentials."""
from google.cloud import aiplatform
from vertexai.language_models import TextEmbeddingModel
import google.auth
import json

output = []

try:
    creds, project = google.auth.default()
    output.append(f"Using project: {project}")
    
    aiplatform.init(project=project, location="us-central1", credentials=creds)
    output.append("Vertex AI initialized")
    
    model = TextEmbeddingModel.from_pretrained("text-embedding-005")
    output.append("Model loaded: text-embedding-005")
    
    embeddings = model.get_embeddings(["What is force and pressure?"])
    for emb in embeddings:
        output.append(f"Dimension: {len(emb.values)}")
        output.append(f"First 5 values: {emb.values[:5]}")
    
    output.append("SUCCESS: Vertex AI embeddings working!")
except Exception as e:
    output.append(f"ERROR: {e}")
    import traceback
    output.append(traceback.format_exc())

with open("vertex_test_output.txt", "w") as f:
    f.write("\n".join(output))

print("\n".join(output))
