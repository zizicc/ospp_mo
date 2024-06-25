# ospp_mo
## preparations
* To use unipdf, you can sign up and create a free key at https://cloud.unidoc.io
* To use llama3, install ollama
* To use PgVector, install postgresql and vector extension
## getting started
* run go main.go input.pdf
## workflow
* read your pdf path from cli
* extract texts from pdf and converts them to chunks using unipdf, then remove empty ones. Sample outputs:
  * ---------Chunk 0-----------
  * My most popular learning sessions, speeches, and TED-like talks
  * ----------Chunk 0 ---------
  * ---------Chunk 1-----------
  * Are you looking for speakers at your learning day, team meeting, or offsite? Read on for
  * talks I’d be thrilled to deliver to your team, based on popular writings on Mind The
  * Beet.
  * ---------Chunk 1-----------
* generate embeddings from chunks by llama3. Sample outputs(only prints the first 4 dimensions):
  * Got 2 embeddings:
  * 0: len=4096; first few=[0.8820618 1.7218097 1.4975904 1.0243285]
  * 1: len=4096; first few=[0.44342858 -3.340699 1.2843275 0.8755332]
* store embeddings in PgVector and query the id and content. Sample outputs:
  * 2 Are you looking for speakers at your learning day, team meeting, or offsite? Read on for
  * talks I’d be thrilled to deliver to your team, based on popular writings on Mind The
  * Beet.
## references
* https://github.com/unidoc/unipdf
* https://github.com/tmc/langchaingo
* https://eli.thegreenplace.net/2023/using-ollama-with-langchaingo/
* https://github.com/pgvector/pgvector
* https://github.com/pgvector/pgvector-go
