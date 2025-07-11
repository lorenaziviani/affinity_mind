import requests
import numpy as np

def test_insert_and_query():
    url_insert = "http://localhost:8001/insert"
    url_query = "http://localhost:8001/query"
    vec = np.random.rand(8).tolist()
    resp = requests.post(url_insert, json={"id": "test", "vector": vec})
    assert resp.status_code == 200
    resp2 = requests.post(url_query, json={"vector": vec, "k": 1})
    assert resp2.status_code == 200
    result = resp2.json()
    assert "ids" in result and "distances" in result
    assert result["ids"][0] == "test" 