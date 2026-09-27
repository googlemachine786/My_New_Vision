# Visionary RAG Pipeline - Real Ollama Test Results

**Test Date**: 2026-03-27T11:00:57.771348
**Ollama URL**: http://localhost:11434
**Embed Model**: nomic-embed-text
**LLM Model**: llama3.2:3b
**Status**: COMPLETE - PASS

## Summary

- **Total Tests**: 4
- **Passed**: 4
- **Failed**: 0
- **Success Rate**: 100.0%

## Ingestion Results

- **Status**: success
- **Chunks Embedded**: 50
- **Embedding Dimensions**: 768
- **Time**: 112.22s

## Query Results

| # | Query | Answer Preview | Latency (ms) |
|---|-------|---------------|--------------|
| 1 | What is photosynthesis? | ✅ Unfortunately, the provided context doesn't explic... | 12205.56 |
| 2 | What are the parts of a cell? | ✅ In the provided context, the question is asking ab... | 13247.45 |
| 3 | What is force? | ✅ According to the context, force is defined as "a p... | 12280.54 |

## Performance Metrics

- **Average Latency**: 12577.85 ms
- **P95 Latency**: 13247.45 ms
- **P99 Latency**: 13247.45 ms
- **Success Rate**: 100.0%

## Conclusion

✅ **All tests passed!** Real Ollama embeddings and LLM working correctly.
