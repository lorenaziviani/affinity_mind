from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
import os
import time

try:
    from sentence_transformers import SentenceTransformer
    model = SentenceTransformer(os.getenv("MODEL_NAME", "sentence-transformers/all-MiniLM-L6-v2"))
    use_local_model = True
except ImportError:
    model = None
    use_local_model = False

import requests

OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
OPENAI_MODEL = os.getenv("OPENAI_MODEL", "text-embedding-ada-002")

app = FastAPI()

class EmbedRequest(BaseModel):
    text: str

class EmbedResponse(BaseModel):
    embedding: List[float]
    elapsed_ms: float
    provider: str

@app.post("/embed", response_model=EmbedResponse)
def embed(req: EmbedRequest):
    start = time.time()
    if use_local_model and model:
        emb = model.encode(req.text).tolist()
        elapsed = (time.time() - start) * 1000
        return EmbedResponse(embedding=emb, elapsed_ms=elapsed, provider="sentence-transformers")
    elif OPENAI_API_KEY:
        headers = {"Authorization": f"Bearer {OPENAI_API_KEY}"}
        json = {"input": req.text, "model": OPENAI_MODEL}
        resp = requests.post("https://api.openai.com/v1/embeddings", headers=headers, json=json)
        if resp.status_code == 200:
            emb = resp.json()["data"][0]["embedding"]
            elapsed = (time.time() - start) * 1000
            return EmbedResponse(embedding=emb, elapsed_ms=elapsed, provider="openai")
        else:
            raise HTTPException(status_code=500, detail="OpenAI API error: " + resp.text)
    else:
        raise HTTPException(status_code=500, detail="No embedding provider available.") 