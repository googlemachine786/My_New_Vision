"""Test Vertex AI embeddings with service account impersonation."""
from google.oauth2 import service_account
from google.auth import impersonated_credentials
import google.auth
from google.cloud import aiplatform
from vertexai.language_models import TextEmbeddingModel

print("Getting default credentials...")
creds, project = google.auth.default()
print(f"Default creds: {type(creds).__name__}")

print("Creating impersonated credentials...")
target_sa = "ragpipeline-dev@visionary-research.iam.gserviceaccount.com"
scopes = ["https://www.googleapis.com/auth/cloud-platform"]

imp_creds = impersonated_credentials.Credentials(
    source_credentials=creds,
    target_principal=target_sa,
    target_scopes=scopes,
    lifetime=3600
)

print("Initializing Vertex AI...")
aiplatform.init(
    project="visionary-research",
    location="us-central1",
    credentials=imp_creds
)
print("Vertex AI initialized")

print("Loading text-embedding-005...")
model = TextEmbeddingModel.from_pretrained("text-embedding-005")
print("Model loaded")

texts = [
    "What is force and pressure?",
    "Explain photosynthesis process",
    "Newton's laws of motion"
]

print(f"\nGenerating embeddings for {len(texts)} texts...")
embeddings = model.get_embeddings(texts)

for i, emb in enumerate(embeddings):
    print(f"\n  Text: {texts[i]}")
    print(f"  Dimension: {len(emb.values)}")
    print(f"  First 5 values: {[round(v, 6) for v in emb.values[:5]]}")

print(f"\nSUCCESS: Vertex AI embeddings working ({len(embeddings)}/{len(texts)})")
