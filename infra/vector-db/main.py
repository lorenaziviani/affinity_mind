from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
import faiss
import numpy as np

app = FastAPI()

class InsertRequest(BaseModel):
    id: str
    vector: List[float]

class QueryRequest(BaseModel):
    vector: List[float]
    k: int = 5

class QueryResponse(BaseModel):
    ids: List[str]
    distances: List[float]

# Simple in-memory vector database
vectors = []
ids = []
index = None

@app.post("/insert")
def insert(req: InsertRequest):
    global vectors, ids, index
    vec = np.array(req.vector, dtype=np.float32)
    if len(vectors) == 0:
        index = faiss.IndexFlatL2(len(vec))
    vectors.append(vec)
    ids.append(req.id)
    index.add(np.expand_dims(vec, axis=0))
    return {"status": "ok"}

@app.post("/query", response_model=QueryResponse)
def query(req: QueryRequest):
    global vectors, ids, index
    if index is None or len(vectors) == 0:
        raise HTTPException(status_code=404, detail="No vectors in database.")
    vec = np.array(req.vector, dtype=np.float32).reshape(1, -1)
    D, I = index.search(vec, req.k)
    result_ids = [ids[i] for i in I[0] if i < len(ids)]
    result_distances = D[0][:len(result_ids)]
    return QueryResponse(ids=result_ids, distances=result_distances) 