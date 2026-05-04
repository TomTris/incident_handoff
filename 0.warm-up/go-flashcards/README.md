# Go Flashcards

Learn by using flashcards — a React app with 26 questions about Go fundamentals.

## Design

Inspired by Vietnam Math University Entrance Exams, the questions are structured in 4 levels based on Bloom's Taxonomy — not only to recall, but to test understanding and connect the dots.

### 4 Levels of Questions

| Level | Focus | Examples |
|-------|-------|---------|
| 🟢 Level 1 — Remembering | Recall previously learned information | Define, list, name, identify |
| 🟡 Level 2 — Understanding | Explain or interpret in your own words | Explain why, distinguish, summarize |
| 🟠 Level 3 — Applying | Use knowledge in specific situations | Calculate, solve, apply, demonstrate |
| 🔴 Level 4 — Higher-Order Thinking | Analyze, evaluate, or create in new contexts | Compare, prove, propose, multi-step |

### Topics Covered

- Go modules and project setup
- Slices (internals, capacity, shared memory)
- Structs, methods, and struct tags
- `fmt.Println` vs `fmt.Printf`
- File I/O with `os.Open`, `defer`, and buffers
- Zero values, `make()`, and `range`

## How to Run

```bash
cd go-flashcards
npm install
npm run dev
```

Then open http://localhost:5173

## Flashcard Generation

- Claude generates both the code and content
- Content is based on the YouTube video transcription and additional questions asked during the learning process

Enjoy!
