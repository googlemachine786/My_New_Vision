#!/usr/bin/env python3
"""
Visionary RAG - Terminal Client
Connects to the REAL Go Orchestrator backend via its API and streams Server-Sent Events.
"""

import os
import sys
import json
import base64
import hmac
import hashlib
import time
import requests

def base64url_encode(data: bytes) -> str:
    """Base64Url encode bytes to string."""
    return base64.urlsafe_b64encode(data).decode('utf-8').rstrip('=')

def generate_mock_jwt() -> str:
    """Generate a valid HS256 JWT using local secrets to authenticate with Go backend."""
    # Attempt to read JWT_SECRET from .env.local
    secret = "visionary-secret-key" # Default fallback
    env_paths = [".env.local", "orchestrator/.env.local"]
    for path in env_paths:
        if os.path.exists(path):
            with open(path, 'r') as f:
                for line in f:
                    if line.startswith("JWT_SECRET="):
                        val = line.split("=", 1)[1].strip()
                        secret = val.strip('"').strip("'")
                        break
    
    header = {"alg": "HS256", "typ": "JWT"}
    payload = {
        "user_id": "student-cli",
        "grade": 8,
        "subject": "Science",
        "taxonomy_id": 1,
        "session_id": "session-cli-1234",
        "exp": int(time.time()) + 3600
    }
    
    encoded_header = base64url_encode(json.dumps(header).encode('utf-8'))
    encoded_payload = base64url_encode(json.dumps(payload).encode('utf-8'))
    
    signature_input = f"{encoded_header}.{encoded_payload}"
    signature = hmac.new(
        key=secret.encode('utf-8'),
        msg=signature_input.encode('utf-8'),
        digestmod=hashlib.sha256
    ).digest()
    
    encoded_signature = base64url_encode(signature)
    return f"{signature_input}.{encoded_signature}"

class TerminalClient:
    def __init__(self, port=8080):
        self.api_url = f"http://localhost:{port}/query"
        self.session_id = f"cli-sess-{int(time.time())}"
        self.jwt_token = generate_mock_jwt()
        
    def ask(self, query: str):
        print("\n" + "="*70)
        print("🤖 Generating answer...")
        print("="*70 + "\n")
        
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.jwt_token}"
        }
        
        payload = {
            "query": query,
            "session_id": self.session_id
        }
        
        try:
            # Send POST request and start streaming
            response = requests.post(
                self.api_url, 
                json=payload, 
                headers=headers,
                stream=True,
                timeout=60
            )
            
            if response.status_code != 200:
                print(f"❌ API Error {response.status_code}: {response.text}")
                return
                
            # Parse Server-Sent Events (SSE)
            sources = []
            
            for line in response.iter_lines():
                if not line:
                    continue
                    
                line_str = line.decode('utf-8')
                
                if line_str.startswith("event: error"):
                    # The next line should contain the data
                    error_data = next(response.iter_lines()).decode('utf-8').replace('data: ', '')
                    print(f"\n❌ Error from server: {error_data}")
                    break
                    
                elif line_str.startswith("data: "):
                    content = line_str[6:] # Strip 'data: '
                    
                    if content == "[DONE]":
                        break
                        
                    try:
                        # Sometimes event complete holds JSON
                        obj = json.loads(content)
                        if "status" in obj and obj["status"] == "success":
                            break
                    except Exception:
                        pass
                        
                    # It's a token stream! Print it out cleanly without newline
                    sys.stdout.write(content.replace('\\n', '\n'))
                    sys.stdout.flush()
                    
            print("\n\n" + "-"*70)
            
        except requests.exceptions.ConnectionError:
            print(f"❌ Connection Error: Could not connect to Go Orchestrator at {self.api_url}.")
            print("   Please make sure the server is running (go run orchestrator/cmd/server/main.go).")
        except Exception as e:
            print(f"\n❌ Unexpected error: {e}")

def run_interactive():
    print("="*70)
    print("VISIONARY RAG - TERMINAL CLIENT")
    print("="*70)
    print("Connecting to REAL Go Orchestrator (AlloyDB + Gemini SSE)")
    print("Type 'quit' or 'exit' to stop.")
    print("="*70)
    
    client = TerminalClient()
    
    while True:
        try:
            query = input("\n❓ Ask: ").strip()
            if query.lower() in ['quit', 'exit', 'q']:
                print("\n👋 Goodbye!")
                break
            if not query:
                continue
                
            client.ask(query)
            
        except KeyboardInterrupt:
            print("\n\n👋 Goodbye!")
            break

if __name__ == "__main__":
    run_interactive()
