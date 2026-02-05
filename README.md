# ELP Projects (Elm • Go • JavaScript)

This repository groups several small projects built for the **ELP** course, across different languages:

- **Elm**: a Wordle clone (as a browser app)
- **Go**: image-processing (sequential vs parallel using goroutines) + a TCP client/server demo, where a client sends images and the server executes the processing and returns the image
- **JavaScript (Node.js)**: an interactive card game (Flip7-style)

## Repository layout

- `projet-elm/` — Elm Wordle (web)
- `projet-go/GO/` — Go image processing (demo + client/server)
- `JS/` — Node.js card game

---

## 1) Elm — Wordle clone

**Path:** `projet-elm/`

A Wordle-like game implemented with **The Elm Architecture** (Model / View / Update / Subscriptions). The list of valid words is loaded from `words.txt`, and a target word is picked randomly.

### Prerequisites

- Elm
- Python 3 (to execute the server)

### Build & run

```bash
cd projet-elm
elm make src/Main.elm --output=main.js
python3 -m http.server 8000
```

Then open: http://localhost:8000

### Useful files

- `projet-elm/src/Main.elm` — game logic + UI
- `projet-elm/index.html` — loads `main.js`
- `projet-elm/words.txt` — dictionary of valid 5-letter words
- `projet-elm/ARCHITECTURE.md` — technical explanation of the code structure

---

## 2) JavaScript — Card game

**Path:** `JS/`

A **Node.js** terminal game where 2–8 players play rounds until someone reaches **200 points**. The game is interactive and prompts for player count and names.

### Prerequisites

- Node.js (recommended: a recent LTS, e.g. Node 18+)

### Install & run

```bash
cd JS
npm install
npm start
```

### Notes

- Game entrypoint: `JS/index.js`
- Logs are written under `JS/logs/`

---

## 3) Go — Image processing (sequential vs parallel) + TCP server demo

**Path:** `projet-go/GO/`

This Go code explores **parallelization** (goroutines/workers) on image-processing tasks.

### Prerequisites

- Go

### A) Demo project: sequential vs parallel comparison

**Path:** `projet-go/GO/demo-project/`

Runs an image pipeline and compares sequential vs parallel implementations (and can generate output images such as `out.png`).

Run:

```bash
cd projet-go/GO/demo-project
go run .
```

Benchmarks:

```bash
go test -bench=. -benchmem
```

(See the project’s own README for details.)

### B) TCP server/client: remote image processing

- **Server:** `projet-go/GO/server/` (listens on `:9000`)
- **Client:** `projet-go/GO/client/`

The client sends an image to the server, chooses a processing mode, and receives a processed image back.

Start the server:

```bash
cd projet-go/GO/server
go run .
```

In another terminal, run the client:

```bash
cd projet-go/GO/client
go run .
```

The client will prompt for:
- server IP
- input image path
- processing choice (B&W / downscale / remap)

Output is written under `projet-go/GO/client/output/`.

---

## Status / scope

These are educational projects; the focus is clarity and experimentation (architecture in Elm, concurrency/parallelism in Go, and a CLI game in Node.js).
