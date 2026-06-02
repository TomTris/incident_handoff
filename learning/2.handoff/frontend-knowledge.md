# Frontend Knowledge — Reference

The textbook for [`handoff-frontend-curriculum.md`](./handoff-frontend-curriculum.md). That file is the challenges; this file is what you read before each one. Each section names the phase it supports.

Contents: The model · HTML (F1) · CSS (F1) · JavaScript (F2) · The DOM & events (F3) · Async & fetch (F4) · TypeScript (F5) · What frameworks do (after F6).

---

## The model

Frontend is three languages in one runtime (the browser), each owning one concern:

| Layer | Language | Concern |
|---|---|---|
| Structure | HTML | What exists on the page |
| Presentation | CSS | How it looks and is laid out |
| Behavior | JavaScript | What happens when the user acts |

The browser parses HTML into a tree (the DOM), applies CSS to style it, then hands JavaScript a live API to that tree. Every framework (Vue, React) is an abstraction over this cycle:

```
HTML → DOM tree → CSS applied → layout calculated → painted
   → JS executes → JS mutates DOM → re-layout/repaint
   → user interacts → event fires → JS handles → DOM mutates → repaint
```

Frameworks optimize the "JS mutates DOM" step. That is their entire value proposition. You do that step by hand first (F3–F4), so the framework stops being magic.

---

## HTML — supports F1

A markup language, not a programming language. It declares a tree of elements. Nesting creates parent-child relationships — that nesting *is* the tree the browser builds.

```html
<article id="post-1" class="featured">
  <h2>Title</h2>
  <p>Body with a <a href="/link">link</a>.</p>
</article>
```

**Elements:** opened (`<p>`) and closed (`</p>`). Void elements self-close (`<img>`, `<input>`, `<br>`).

**Semantic HTML.** Tags carry meaning beyond appearance. `<nav>`, `<main>`, `<button>` announce their role to screen readers, search engines, and other developers. `<div>` and `<span>` mean nothing — generic containers. A `<div>` styled to look like a button is not a button: no keyboard focus, no click semantics, no screen-reader announcement, unless you reimplement all of it. Use the real tag.

**Attributes.** Key-value metadata. Global (`id`, `class`, `data-*`) or element-specific (`href`, `src`, `type`). Boolean attributes are true by presence (`disabled`, `required`, `checked`).

**Document skeleton:**

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Page Title</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <!-- visible content -->
  <script src="app.js"></script>
</body>
</html>
```

**Forms.** Native controls (`<input>`, `<textarea>`, `<select>`, `<button type="submit">`) have built-in validation (`required`, `pattern`, `min`, `max`) and submission behavior. Learn the native layer before any framework rebuilds it. Your F1 login form is native HTML validation.

**Tags to know:** structural (`<div>`, `<span>`); semantic (`<header>`, `<footer>`, `<nav>`, `<main>`, `<section>`, `<article>`, `<aside>`); text (`<h1>`–`<h6>`, `<p>`, `<ul>`/`<ol>`/`<li>`, `<code>`, `<pre>`); media (`<img>`, `<video>`, `<svg>`); interactive (`<a>`, `<button>`, `<input>`, `<form>`, `<label>`, `<select>`); table (`<table>`, `<thead>`, `<tbody>`, `<tr>`, `<th>`, `<td>`).

---

## CSS — supports F1

Rule-based styling. Selectors match elements; declarations set property-value pairs.

```css
article.featured {
  background-color: #f5f5f5;
  padding: 1.5rem;
  border-left: 4px solid #333;
}
```

**Selectors:** `p` (tag), `.featured` (class), `#post-1` (id), `article p` (descendant), `article > p` (direct child), `a:hover` (pseudo-class), `input[type="email"]` (attribute).

**Specificity.** When rules conflict, the most specific wins. Low→high: element < class/attribute/pseudo-class < id < inline `style` < `!important`. Equal specificity → last rule wins. The most common CSS bug. If a style isn't applying, check specificity first.

**Box model.** Every element is a box: content, then padding, then border, then margin (inner to outer). By default `width` sets only content width — padding and border add to it. Fix globally:

```css
*, *::before, *::after { box-sizing: border-box; }
```

**Display:** block elements take full width and stack (`<div>`, `<p>`, `<h1>`); inline elements flow with text (`<span>`, `<a>`). Override with `display: block | inline | inline-block | none`.

**Units:** `px` (absolute — borders, fixed sizes); `rem` (relative to root font size, ~16px — fonts, spacing, most dimensions); `em` (relative to parent — compounds, use sparingly); `%` (relative to parent); `vw`/`vh` (1% of viewport); `fr` (fraction, grid only).

**Layout — the hard part. Three systems:**

*Normal flow* — the default. Block stacks vertically, inline flows and wraps. The baseline mental model.

*Flexbox* — one-dimensional (a row OR a column):

```css
.container {
  display: flex;
  flex-direction: row;       /* main axis */
  justify-content: center;   /* along main axis */
  align-items: center;       /* along cross axis */
  gap: 1rem;
}
.item { flex: 1; }           /* grow to fill equally */
```

Use for: navbars, centering, rows/columns of items, card rows. (Your F1 login form centers with flexbox.)

*Grid* — two-dimensional (rows AND columns):

```css
.grid {
  display: grid;
  grid-template-columns: 1fr 2fr 1fr;
  gap: 1rem;
}
.header { grid-column: 1 / -1; }   /* span all columns */
```

Use for: page layouts, dashboards, 2D alignment. Rule: a line of items → flexbox; a 2D arrangement → grid. Real interfaces nest both.

**Responsive (mobile-first):**

```css
.grid { grid-template-columns: 1fr; }            /* base: mobile */
@media (min-width: 768px)  { .grid { grid-template-columns: 1fr 1fr; } }
@media (min-width: 1024px) { .grid { grid-template-columns: 1fr 2fr 1fr; } }
```

**Positioning:** `relative` (offset from normal), `absolute` (relative to nearest positioned ancestor), `fixed` (relative to viewport), `sticky` (relative until a scroll threshold, then fixed). Most layout is flexbox/grid; positioning is for overlays, tooltips, sticky headers.

**Custom properties:**

```css
:root { --color-primary: #2563eb; --spacing-md: 1rem; }
.button { background: var(--color-primary); padding: var(--spacing-md); }
```

---

## JavaScript — supports F2

The browser's programming language. Dynamically typed, single-threaded, event-driven. No static types, no goroutines, no compile step in raw form.

**Variables.** `const` by default, `let` when reassigning, never `var`.

```js
const name = "Alice";          // string
const count = 42;              // number (no int/float distinction)
const active = true;           // boolean
const nothing = null;          // intentional absence
const missing = undefined;     // uninitialized
const items = [1, 2, 3];       // array
const user = { name: "Alice", age: 30 };  // object
```

**Coercion.** JS silently converts types: `"5" + 3` → `"53"`, `"5" - 3` → `2`. The classic footgun. Always use `===` (strict), never `==`.

**Functions:**

```js
const add = (a, b) => a + b;          // arrow — preferred default
function addF(a, b) { return a + b; } // traditional — when you need `this` or hoisting
```

**Objects are references.** Assigning does not copy:

```js
const a = { x: 1 };
const b = a;          // same object
b.x = 2;              // a.x is now 2
const c = { ...a };           // shallow copy
const d = structuredClone(a); // deep copy
```

Critical for understanding reactivity later.

**Destructuring:**

```js
const { name, age } = user;
const [first, second] = items;
const { name, ...rest } = user;   // rest = { age: 30 }
```

**Array methods** — replace most loops, return new arrays, do not mutate (except `forEach`):

```js
nums.map(n => n * 2);                 // transform each
nums.filter(n => n > 3);              // keep matching
nums.reduce((sum, n) => sum + n, 0);  // accumulate
nums.find(n => n > 3);                // first match
nums.some(n => n > 3);                // any?
nums.every(n => n > 0);               // all?
nums.forEach(n => console.log(n));    // side effect, no return
```

Your F2 incident-array logic (count by severity, filter open, sort) is built from these.

**Closures** — a function keeps access to its creation scope after that scope exits:

```js
function makeCounter() {
  let count = 0;            // captured
  return () => { count += 1; return count; };
}
const c = makeCounter();
c(); // 1
c(); // 2
```

Closures are why callbacks, React hooks, and Vue's Composition API work. Not optional.

**`this`.** In traditional functions, `this` depends on *how* the function is called. In arrow functions, `this` is inherited from the enclosing scope. Rule of thumb: arrow functions for callbacks, traditional functions for object methods.

**Event loop.** JS is single-threaded. Long operations (network, timers) are non-blocking: they register a callback that runs later when the call stack is empty. Promises (microtasks) run before timer/DOM callbacks.

**ES Modules:**

```js
// math.js
export const add = (a, b) => a + b;
export default function multiply(a, b) { return a * b; }
// app.js
import multiply, { add } from './math.js';
```

In the browser: `<script type="module" src="app.js"></script>`.

---

## The DOM & events — supports F3

The DOM is the browser's live tree of JS objects, built from your HTML. JS reads and mutates it. Every framework is, at core, a DOM-manipulation library.

```js
// Query
const el = document.getElementById("post-1");
const el2 = document.querySelector(".featured");    // first match
const all = document.querySelectorAll("article");    // NodeList

// Content
el.textContent = "New text";                 // text only (safe)
el.innerHTML = "<strong>Bold</strong>";      // parses HTML (XSS risk with user input)

// Attributes / classes
el.setAttribute("data-id", "42");
el.classList.add("active");
el.classList.toggle("active");

// Create / insert / remove
const div = document.createElement("div");
div.textContent = "Hello";
document.body.appendChild(div);
el.remove();
```

**Events:**

```js
button.addEventListener("click", (event) => {
  event.preventDefault();    // stop default (e.g. form submit)
  // handle
});
```

**Propagation.** Events bubble up the tree: button → its div → parent → up to `<html>`. Stop with `event.stopPropagation()`.

**Delegation.** Instead of 100 listeners on 100 list items, attach one to the parent and check `event.target`:

```js
list.addEventListener("click", (e) => {
  if (e.target.tagName === "LI") { /* handle e.target */ }
});
```

How frameworks handle dynamic lists efficiently.

**Render-from-state — the F3 pattern.** Hold state in one array. On every change, rebuild the DOM from that array. Never treat the DOM as the source of truth.

**Why it matters:** Vue's `ref()`/`reactive()` reactivity is automated calls to these exact DOM APIs. When `count.value++` updates the page, Vue is running `textContent`/`createElement`/etc. under the hood. Do it by hand once and the abstraction is legible.

---

## Async & fetch — supports F4

Promises represent future values. `async`/`await` is syntax over them.

```js
async function loadIncidents(token) {
  const list = document.querySelector("#list");
  list.textContent = "Loading...";
  try {
    const res = await fetch("http://localhost:8080/your-incidents-endpoint", {
      headers: { "Authorization": `Bearer ${token}` },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    render(data);                 // build DOM from data
  } catch (err) {
    list.textContent = `Error: ${err.message}`;
  }
}
```

**Login POST shape:**

```js
const res = await fetch("http://localhost:8080/your-auth-endpoint", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username, password }),
});
const { token } = await res.json();   // field name = your real response
```

Hold `token` in a JS variable — that is your in-memory auth state. Send it as `Authorization: Bearer <token>` on subsequent requests.

**CORS.** Your frontend (e.g. `http://localhost:5500`) and your Go backend (e.g. `http://localhost:8080`) are different origins. The browser blocks the request unless the backend sends CORS headers. Add a CORS middleware to Phase 9 if it lacks one. The block happens *before your code runs*; the console and Network panel show it.

**DevTools Network panel** — every request, payload, status, timing. Your primary debugging surface for fetch.

---

## TypeScript — supports F5

A static type system over JavaScript. Compiles to plain JS — the browser never sees TS. It catches at compile time what JS permits silently. Typing is **structural** — two types are compatible if their shapes match, regardless of name.

**Inference:** annotate function parameters and return types; let inference handle locals.

**Interfaces and aliases:**

```ts
interface Incident {
  id: number;
  title: string;
  severity: "low" | "medium" | "high" | "critical";   // literal union
  status: "open" | "resolved";
  createdAt: string;
  resolvedAt?: string;        // optional
}

interface LoginResponse { token: string; }   // match your real response

type ID = string | number;            // union
type Handler = (e: Event) => void;    // function type
```

Use `interface` for object shapes; `type` for unions, tuples, function signatures.

**Union types + narrowing:**

```ts
function format(input: string | number): string {
  if (typeof input === "string") return input.toUpperCase();  // narrowed to string
  return input.toFixed(2);                                    // narrowed to number
}
```

Narrowing mechanisms: `typeof`, `instanceof`, `in`, equality checks, discriminated unions.

**Discriminated unions** — the pattern for modeling fetch state. Use this for your incident list in F5:

```ts
type FetchState<T> =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "success"; data: T }
  | { status: "error"; error: string };

function render(s: FetchState<Incident[]>) {
  switch (s.status) {
    case "idle":    return "";
    case "loading": return "Loading...";
    case "success": return s.data;        // TS knows data exists
    case "error":   return s.error;       // TS knows error exists
    default:
      const _exhaustive: never = s;       // compile error if a case is missing
      return _exhaustive;
  }
}
```

**Generics:**

```ts
function filter<T>(items: T[], pred: (item: T) => boolean): T[] {
  return items.filter(pred);
}
```

**Utility types:** `Partial<T>`, `Required<T>`, `Readonly<T>`, `Pick<T, K>`, `Omit<T, K>`, `Record<K, V>`. Compile-time only, zero runtime cost.

**Assertions** (escape hatches — sparingly):

```ts
const input = document.querySelector("#email") as HTMLInputElement;
const el = document.getElementById("app")!;   // non-null assertion
```

Prefer narrowing (`if (el !== null)`) over assertions.

**`any` vs `unknown`:** `any` disables all checking — avoid. `unknown` forces narrowing before use — prefer it.

**Async types:**

```ts
async function fetchIncidents(token: string): Promise<Incident[]> {
  const res = await fetch("/your-incidents-endpoint", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json() as Promise<Incident[]>;
}
try { await fetchIncidents(token); }
catch (error: unknown) {
  if (error instanceof Error) console.error(error.message);
}
```

**tsconfig — always `strict`:**

```json
{ "compilerOptions": {
  "target": "ES2022", "module": "ESNext",
  "strict": true, "noUncheckedIndexedAccess": true, "skipLibCheck": true
}}
```

`strict: true` enables `strictNullChecks`, `noImplicitAny`, and more. Without it the safety guarantees are gutted.

**Tooling — Vite (introduced at F5).** Browsers do not understand TypeScript; something must compile `.ts` → `.js` before the browser sees it. That converter is Vite, your first tooling. It does three jobs: compiles TS on the fly, runs a dev server with hot reload, and produces an optimized build for deploy. Setup: `npm create vite@latest` → vanilla-ts template → `npm install` → `npm run dev`. After this you never run `tsc` by hand.

---

## What frameworks do — read after F6

Once the minimum slice is built by hand, this is what Vue/React automate:

1. **Templating** — declarative templates compile to the `createElement` calls you wrote by hand.
2. **Reactivity** — declared state auto-updates the DOM; you stop calling re-render yourself.
3. **Components** — the UI composes from isolated, reusable pieces with their own state and styles.
4. **Minimal DOM mutation** — React diffs a virtual tree; Vue tracks individual reactive refs. Both minimize expensive real-DOM operations.

None of it is magic once you've built the same things manually. That is the point of this work, and the bridge into Phase 11.