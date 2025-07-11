import requests

def test_embed():
    url = "http://localhost:5000/embed"
    data = {"text": "hello world"}
    resp = requests.post(url, json=data)
    assert resp.status_code == 200
    result = resp.json()
    assert "embedding" in result
    assert isinstance(result["embedding"], list)
    assert result["embedding"] 