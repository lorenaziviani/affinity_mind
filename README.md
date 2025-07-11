# AffinityMind

## Motivação

O AffinityMind nasce da necessidade de oferecer recomendações inteligentes e personalizadas, utilizando técnicas modernas de embeddings e similaridade vetorial. O objetivo é proporcionar experiências mais relevantes para os usuários, conectando-os a conteúdos, produtos ou pessoas de acordo com seus interesses reais.

## Overview Técnico

O sistema é composto por um backend em Go, responsável por orquestrar as requisições dos usuários, uma API de embeddings desenvolvida em Python (utilizando Sentence Transformers), e um banco de dados vetorial para armazenamento e busca eficiente dos embeddings. A comunicação entre os componentes é feita via HTTP/REST.

### Principais Componentes

- **Go Backend** (`cmd/backend`): expõe a API principal para o usuário e integra os demais serviços.
- **Recommender** (`pkg/recommender`): lógica de recomendação e integração com o banco vetorial.
- **Embedding Server** (`ml/embedding-server`): API Python para geração de embeddings.
- **Vector DB** (`infra/vector-db`): banco de dados vetorial (ex: Milvus, Qdrant, Pinecone).

## Arquitetura

O fluxo principal do sistema é:

1. **User Action**: Usuário faz uma requisição de recomendação.
2. **Go Backend**: Recebe a requisição e solicita o embedding ao serviço Python.
3. **Embedding API (Python)**: Gera o embedding do item/usuário.
4. **Vector DB**: Consulta os itens mais similares.
5. **Recommendations**: Backend retorna as recomendações ao usuário.

O diagrama detalhado está disponível em `docs/arquitetura.drawio`.

## Como rodar o projeto (primeiros passos)

1. Clone o repositório
2. Configure as variáveis de ambiente conforme `env-sample`
3. Siga as instruções específicas em cada subdiretório

---

## Estrutura de Diretórios

```
cmd/backend           # Backend Go
pkg/recommender       # Lógica de recomendação
infra/vector-db       # Banco vetorial
ml/embedding-server   # API Python de embeddings
docs                  # Documentação e diagramas
```
