---
name: llm-judge
description: Grade a single news article's relevance to an IR benchmark query. Use when the bench tool calls claude -p for automated judging, or when asked to grade one article at a time. Returns a single JSON object with doc_id and grade.
---

# LLM Judge

Grade one news article's relevance to a given search query.

## Grading Scale

- **3** — Highly relevant: article is centrally about the query topic
- **2** — Relevant: article clearly covers the topic
- **1** — Marginal: topic is mentioned but not the focus
- **0** — Not relevant: article is about something else

When in doubt between two grades, pick the lower one.

## Input

You will receive:
- A query description (what the searcher is looking for)
- One article with title, description, and a content snippet

## Output

Respond with **only** this JSON — no markdown, no explanation:

```
{"doc_id":"<uuid>","grade":<0|1|2|3>}
```

The `doc_id` must match exactly the one provided in the input.

## Grading rules

- Base your grade on **title + description + content snippet** together
- Content is truncated — if title and description already make relevance clear, don't over-infer from the snippet
- Grade what the article **is about**, not whether the query terms appear in the text
- A political article that mentions climate in passing is grade 1, not 2