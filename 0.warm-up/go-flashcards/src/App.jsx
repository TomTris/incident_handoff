import { useState } from "react";

const cards = [
  // Level 1
  {
    level: 1, label: "🟢 Level 1",
    q: "What command do you run to initialize a Go module?",
    a: `go mod init <module-name>\n\nExample:\ngo mod init github.com/yourname/project`
  },
  {
    level: 1, label: "🟢 Level 1",
    q: "What file does `go mod init` create?",
    a: "go.mod\n\nIt contains:\n- module path\n- Go version\n- dependencies (added later)"
  },
  {
    level: 1, label: "🟢 Level 1",
    q: "What are the two values `range` returns when iterating over a slice?",
    a: "1. i — index (position)\n2. v — value (the element)\n\nfor i, v := range mySlice { }"
  },
  {
    level: 1, label: "🟢 Level 1",
    q: "Name the 3 things a slice stores internally.",
    a: "1. Pointer — points to underlying array\n2. Length — how many elements visible\n3. Capacity — how many elements available"
  },
  {
    level: 1, label: "🟢 Level 1",
    q: "What format verb prints the TYPE of a variable?\nWhat about the MEMORY ADDRESS?",
    a: "%T → prints the type\n   fmt.Printf(\"%T\", x) → []int\n\n%p → prints the memory address\n   fmt.Printf(\"%p\", x) → 0x1400007e0c0"
  },
  {
    level: 1, label: "🟢 Level 1",
    q: "What is the zero value for:\n- int\n- string\n- bool\n- byte",
    a: "int    → 0\nstring → \"\" (empty string)\nbool   → false\nbyte   → 0"
  },
  {
    level: 1, label: "🟢 Level 1",
    q: "What are the only 3 types that make() works with?",
    a: "1. slice  → make([]int, 5)\n2. map    → make(map[string]int)\n3. channel → make(chan int)"
  },
  {
    level: 1, label: "🟢 Level 1",
    q: "What is the difference between fmt.Println and fmt.Printf?",
    a: "fmt.Println → does NOT format\n  prints everything as-is\n  fmt.Println(\"%T\", x) → prints literally \"%T\"\n\nfmt.Printf → DOES format\n  fmt.Printf(\"%T\", x) → prints the actual type"
  },
  {
    level: 1, label: "🟢 Level 1",
    q: "What character is used to write struct tags? Name the character.",
    a: "Backtick  `\n\nExample:\nName string `json:\"name\" validate:\"required\"`\n\nOn Mac: Option + key near backspace\nOn US keyboard: top-left key under ESC"
  },
  {
    level: 1, label: "🟢 Level 1",
    q: "In Go, does a string stop at a \\0 null byte like in C?",
    a: "NO ❌\n\nGo strings are LENGTH-BASED, not null-terminated.\n\nA zero byte is just an invisible character — Go keeps printing.\n\nThat's why we use buf[:n] to print only the real content."
  },
  // Level 2
  {
    level: 2, label: "🟡 Level 2",
    q: "Explain the difference:\na := an_slice[0:3]\nb := an_slice[0:3:3]\n\nWhat is the same? What is different?",
    a: "SAME:\n- Both have len = 3\n- Both point to the same underlying array\n\nDIFFERENT:\n- a: cap = original cap (e.g. 5)\n- b: cap = 3 (limited by 3rd index)\n\nEffect on append:\n- append(a, x) → writes into original array ⚠️\n- append(b, x) → cap exceeded → new memory ✅"
  },
  {
    level: 2, label: "🟡 Level 2",
    q: "Explain why this code is wrong:\nfmt.Println(\"%T\\n\", myVar)",
    a: "fmt.Println does NOT format % verbs.\nIt prints everything literally.\n\nOutput would be:\n%T\n <value of myVar>\n\nFix:\nfmt.Printf(\"%T\\n\", myVar)"
  },
  {
    level: 2, label: "🟡 Level 2",
    q: "Explain what `defer` does.\nWhy do we write `defer f.Close()` right after `os.Open()`?",
    a: "defer → runs the statement when the FUNCTION EXITS\n\nWhy right after Open():\n- If we put f.Close() at the end, it might never run if the code returns early or crashes\n- defer GUARANTEES Close() always runs no matter what\n\nCommon pattern:\nf, err := os.Open(\"file.txt\")\ndefer f.Close() // ← registered now, runs at end"
  },
  {
    level: 2, label: "🟡 Level 2",
    q: "Explain why `validate:\"required\"` does nothing by itself. What else is needed?",
    a: "Struct tags are just METADATA — passive labels.\nGo itself completely ignores them.\n\nYou need a library to READ and ACT on them:\n\nvalidate := validator.New()\nerr := validate.Struct(p) // ← this reads the tags\n\nSame for all tags:\n`json:\"name\"` → needs json.Marshal()\n`db:\"name\"`   → needs a DB library"
  },
  {
    level: 2, label: "🟡 Level 2",
    q: "What is the difference between a METHOD and a REGULAR FUNCTION in Go?",
    a: "Regular function — belongs to nobody:\nfunc PickUpBeer() bool { }\nPickUpBeer() // called standalone\n\nMethod — belongs to a type via receiver:\nfunc (p Person) PickUpBeer() bool { }\np.PickUpBeer() // called on a variable\n\nThe receiver (p Person) is like 'self' or 'this'\nin other languages."
  },
  {
    level: 2, label: "🟡 Level 2",
    q: "Explain why this module path is preferred:\n  module github.com/yourname/project\nover:\n  module myproject",
    a: "Because Go uses the module path to FETCH dependencies.\n\nWhen someone does:\ngo get github.com/yourname/project\n→ Go knows exactly where to download it ✅\n\nWith just 'myproject':\n→ Go has no idea where to find it on the internet ❌\n\nRule:\n- Learning/local → module myproject ✅\n- Publishing/sharing → module github.com/... ✅"
  },
  {
    level: 2, label: "🟡 Level 2",
    q: "Explain why Go strings are NOT null-terminated like C strings.",
    a: "C strings end at \\0 — the null byte signals the end.\n\nGo strings carry their LENGTH explicitly:\n- The string header stores: pointer + length\n- Go always knows where the string ends\n- \\0 is just a regular invisible character\n\nThis is safer — no buffer overflows from missing \\0!"
  },
  {
    level: 2, label: "🟡 Level 2",
    q: "A classmate says:\n\"When I do b := a[0:3], Go copies the data from a into b.\"\n\nIs this correct? Explain why or why not.",
    a: "INCORRECT ❌\n\nb := a[0:3] does NOT copy data.\nb is just a new slice HEADER (pointer + len + cap)\nthat points to the SAME underlying array.\n\nProof:\na[0] = 999\nfmt.Println(b[0]) // also 999!\n\nData is only copied when append exceeds capacity\nand Go allocates new memory."
  },
  // Level 3
  {
    level: 3, label: "🟠 Level 3",
    q: "Write a struct `Car` with fields:\n- Brand (string)\n- Speed (int)\n- Electric (bool)\n\nAdd JSON tags using lowercase keys.",
    a: `type Car struct {
    Brand    string \`json:"brand"\`
    Speed    int    \`json:"speed"\`
    Electric bool   \`json:"electric"\`
}`
  },
  {
    level: 3, label: "🟠 Level 3",
    q: "Write a method on Car that returns true if speed > 200.\n\nfunc ??? IsFast() bool { ... }",
    a: `func (c Car) IsFast() bool {
    return c.Speed > 200
}

// Usage:
myCar := Car{Brand: "Ferrari", Speed: 250}
fmt.Println(myCar.IsFast()) // true`
  },
  {
    level: 3, label: "🟠 Level 3",
    q: "Fix ALL problems in this code:\n\nfunc main() {\n    f, err := os.Open(\"data.txt\")\n    b := make([]byte, 512)\n    n, err := f.Read()\n    fmt.Println(string(b))\n}",
    a: `Problems:
1. err is never checked after os.Open()
2. f.Read() missing buffer argument
3. No defer f.Close()
4. Should use b[:n] not b

Fixed:
func main() {
    f, err := os.Open("data.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    b := make([]byte, 512)
    n, err := f.Read(b)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(b[:n]))
}`
  },
  {
    level: 3, label: "🟠 Level 3",
    q: "What is the output?\n\ns := []int{10, 20, 30, 40, 50}\na := s[0:3]\na[0] = 999\nfmt.Println(s[0])",
    a: "Output: 999\n\nWhy:\n- a := s[0:3] does NOT copy — a points to same memory as s\n- a[0] = 999 writes to the shared underlying array\n- s[0] also reads from that same array\n- So s[0] is now 999"
  },
  {
    level: 3, label: "🟠 Level 3",
    q: "Write a for loop using range that prints only the VALUES (not indexes) of:\nnums := []int{5, 10, 15, 20}",
    a: `nums := []int{5, 10, 15, 20}

for _, v := range nums {
    fmt.Println(v)
}

// Output:
// 5
// 10
// 15
// 20

// _ discards the index`
  },
  {
    level: 3, label: "🟠 Level 3",
    q: "Write method IsAdult() on Person that returns true if age >= 18.\nUse the SHORTEST possible code (one line return).\n\ntype Person struct {\n    Name string\n    Age  int\n}",
    a: `func (p Person) IsAdult() bool {
    return p.Age >= 18
}

// p.Age >= 18 is already a bool expression
// No if/else needed!
// Age 18 → true ✅
// Age 17 → false ✅`
  },
  {
    level: 3, label: "🟠 Level 3",
    q: "What is the output and why?\n\ns := []string{\"a\", \"b\", \"c\", \"d\"}\nx := s[0:2:2]\nx = append(x, \"Z\")\nfmt.Println(s)",
    a: `Output: [a b c d]

Why:
- x := s[0:2:2] → len=2, cap=2
- append(x, "Z") → cap exceeded!
- Go allocates NEW memory for x
- s is NOT affected ✅

If it were s[0:2] (cap=4):
- append would write into s[2]
- s would become [a b Z d] ⚠️

The 3rd index :2 protected s!`
  },
  // Level 4
  {
    level: 4, label: "🔴 Level 4",
    q: "Predict the output and explain each line:\n\na := []int{1, 2, 3, 4, 5}\nb := a[1:3:4]\nc := append(b, 99)\n\nfmt.Printf(\"a = %v\\n\", a)\nfmt.Printf(\"b = %v\\n\", b)\nfmt.Printf(\"c = %v\\n\", c)\nfmt.Printf(\"same memory? %v\\n\", &a[0] == &c[0])",
    a: `a = [1 2 3 99 5]
b = [2 3]
c = [2 3 99]
same memory? false

Why:
- b = a[1:3:4] → points to a[1], len=2, cap=3
- append(b, 99): len=2, cap=3 → fits!
  → writes 99 into a[3] (shared memory!)
  → a[3] changes from 4 to 99
- c is a new slice header: [2, 3, 99] len=3, cap=3
- &a[0] vs &c[0]: a[0]=1, c[0]=2 → different addresses → false`
  },
  {
    level: 4, label: "🔴 Level 4",
    q: "Find 2 problems with this struct design and suggest fixes:\n\ntype User struct {\n    Username string\n    Password string `json:\"password\"`\n    Age      int    `validate:\"min=13\"`\n}",
    a: `Problem 1: Password has json:"password" tag
→ Password will be included in JSON output!
→ NEVER expose passwords in JSON
Fix: json:"-" to exclude it
  Password string \`json:"-"\`

Problem 2: Age has min=13 but no "required"
→ Age=0 would pass validation (0 is zero value)
Fix: add required
  Age int \`validate:"required,min=13"\`

Bonus: Username has no validation at all
  Username string \`validate:"required,min=3"\``
  },
  {
    level: 4, label: "🔴 Level 4",
    q: "Which approach is better and why?\n\n// A\nvar buf [512]byte\nf.Read(buf[:])\nfmt.Println(string(buf[:]))\n\n// B\nbuf := make([]byte, 512)\nn, _ := f.Read(buf)\nfmt.Println(string(buf[:n]))",
    a: `Approach B is better ✅

Problems with A:
1. [512]byte is an ARRAY — fixed forever, can't grow
2. buf[:] passes whole array awkwardly
3. string(buf[:]) prints ALL 512 bytes including zeros
   → output has ugly invisible \0 chars at the end
4. Ignores how many bytes were actually read

Why B is better:
1. []byte is a SLICE — can grow with append
2. n, _ := f.Read(buf) captures actual bytes read
3. buf[:n] prints ONLY real content, no zeros
4. Cheaper to pass to other functions (pointer, not copy)`
  },
  {
    level: 4, label: "🔴 Level 4",
    q: "Design a BankAccount struct with:\n- Fields: Owner (string), Balance (float64)\n- Deposit(amount float64)\n- Withdraw(amount float64) error\n- IsRich() bool (true if balance > 1,000,000)",
    a: `type BankAccount struct {
    Owner   string
    Balance float64
}

func (b *BankAccount) Deposit(amount float64) {
    b.Balance += amount
}

func (b *BankAccount) Withdraw(amount float64) error {
    if b.Balance < amount {
        return errors.New("insufficient balance")
    }
    b.Balance -= amount
    return nil
}

func (b BankAccount) IsRich() bool {
    return b.Balance > 1_000_000
}

Note: Deposit and Withdraw use *BankAccount
(pointer receiver) so they actually modify
the balance. IsRich only reads, so no pointer needed.`
  },
  {
    level: 4, label: "🔴 Level 4",
    q: "Find the subtle bug. What prints and why?\n\nfunc main() {\n    files := []string{\"a.txt\", \"b.txt\", \"c.txt\"}\n    s := make([]string, 3)\n\n    for i, name := range files {\n        s[i] = name\n    }\n\n    a := s[0:2]\n    b := s[0:2]\n\n    a[0] = \"CHANGED\"\n    fmt.Println(b[0])\n}",
    a: `Output: "CHANGED"

The bug: a and b are NOT independent!

Both a := s[0:2] and b := s[0:2]
point to the SAME underlying array s.

So when a[0] = "CHANGED":
→ it writes into s[0]
→ b[0] reads from the same s[0]
→ b[0] is now "CHANGED"!

Fix — if you want b to be independent:
b := make([]string, 2)
copy(b, s[0:2]) // copy() creates a real separate copy`
  },
];

const levelColors = {
  1: { bg: "#1a2e1a", border: "#4caf50", badge: "#4caf50", text: "#a5d6a7" },
  2: { bg: "#2e2a0e", border: "#ffc107", badge: "#ffc107", text: "#fff9c4" },
  3: { bg: "#2e1a0e", border: "#ff9800", badge: "#ff9800", text: "#ffe0b2" },
  4: { bg: "#2e0e0e", border: "#f44336", badge: "#f44336", text: "#ffcdd2" },
};

export default function App() {
  const [idx, setIdx] = useState(0);
  const [flipped, setFlipped] = useState(false);
  const [filter, setFilter] = useState(0);

  const filtered = filter === 0 ? cards : cards.filter(c => c.level === filter);
  const card = filtered[idx];
  const colors = levelColors[card.level];

  const go = (dir) => {
    setFlipped(false);
    setTimeout(() => {
      setIdx(i => (i + dir + filtered.length) % filtered.length);
    }, 150);
  };

  const changeFilter = (f) => {
    setFilter(f);
    setIdx(0);
    setFlipped(false);
  };

  return (
    <div style={{ minHeight: "100vh", background: "#0d1117", color: "#e6edf3", fontFamily: "monospace", display: "flex", flexDirection: "column", alignItems: "center", padding: "24px 16px" }}>
      
      {/* Title */}
      <div style={{ fontSize: 20, fontWeight: "bold", marginBottom: 20, color: "#58a6ff" }}>
        Go Fundamentals Flashcards
      </div>

      {/* Filter buttons */}
      <div style={{ display: "flex", gap: 8, marginBottom: 24, flexWrap: "wrap", justifyContent: "center" }}>
        {[
          { f: 0, label: "All", color: "#58a6ff" },
          { f: 1, label: "🟢 L1", color: "#4caf50" },
          { f: 2, label: "🟡 L2", color: "#ffc107" },
          { f: 3, label: "🟠 L3", color: "#ff9800" },
          { f: 4, label: "🔴 L4", color: "#f44336" },
        ].map(({ f, label, color }) => (
          <button key={f} onClick={() => changeFilter(f)} style={{
            padding: "6px 14px", borderRadius: 20, border: `1px solid ${color}`,
            background: filter === f ? color : "transparent",
            color: filter === f ? "#000" : color,
            cursor: "pointer", fontSize: 13, fontWeight: "bold"
          }}>{label}</button>
        ))}
      </div>

      {/* Progress */}
      <div style={{ fontSize: 13, color: "#8b949e", marginBottom: 16 }}>
        {idx + 1} / {filtered.length}
      </div>

      {/* Card */}
      <div onClick={() => setFlipped(f => !f)} style={{
        width: "100%", maxWidth: 600, minHeight: 320,
        background: flipped ? colors.bg : "#161b22",
        border: `2px solid ${flipped ? colors.border : "#30363d"}`,
        borderRadius: 16, padding: 28, cursor: "pointer",
        transition: "all 0.2s ease", position: "relative",
        display: "flex", flexDirection: "column", justifyContent: "space-between"
      }}>
        {/* Badge */}
        <div style={{
          display: "inline-block", padding: "3px 10px", borderRadius: 12,
          background: colors.badge, color: "#000", fontSize: 11,
          fontWeight: "bold", marginBottom: 16, alignSelf: "flex-start"
        }}>{card.label}</div>

        {/* Content */}
        <div style={{ flex: 1 }}>
          {!flipped ? (
            <>
              <div style={{ fontSize: 11, color: "#8b949e", marginBottom: 12, textTransform: "uppercase", letterSpacing: 1 }}>Question</div>
              <div style={{ fontSize: 15, lineHeight: 1.7, whiteSpace: "pre-wrap", color: "#e6edf3" }}>{card.q}</div>
            </>
          ) : (
            <>
              <div style={{ fontSize: 11, color: colors.text, marginBottom: 12, textTransform: "uppercase", letterSpacing: 1 }}>Answer</div>
              <div style={{ fontSize: 14, lineHeight: 1.8, whiteSpace: "pre-wrap", color: colors.text }}>{card.a}</div>
            </>
          )}
        </div>

        {/* Flip hint */}
        <div style={{ marginTop: 20, fontSize: 11, color: "#8b949e", textAlign: "center" }}>
          {flipped ? "👆 Click to see question" : "👆 Click to reveal answer"}
        </div>
      </div>

      {/* Navigation */}
      <div style={{ display: "flex", gap: 16, marginTop: 24, alignItems: "center" }}>
        <button onClick={() => go(-1)} style={{
          padding: "10px 24px", borderRadius: 8, border: "1px solid #30363d",
          background: "#21262d", color: "#e6edf3", cursor: "pointer", fontSize: 16
        }}>← Prev</button>

        <button onClick={() => setFlipped(f => !f)} style={{
          padding: "10px 24px", borderRadius: 8, border: "1px solid #58a6ff",
          background: "transparent", color: "#58a6ff", cursor: "pointer", fontSize: 13, fontWeight: "bold"
        }}>Flip Card</button>

        <button onClick={() => go(1)} style={{
          padding: "10px 24px", borderRadius: 8, border: "1px solid #30363d",
          background: "#21262d", color: "#e6edf3", cursor: "pointer", fontSize: 16
        }}>Next →</button>
      </div>

      {/* Shuffle */}
      <button onClick={() => { setIdx(Math.floor(Math.random() * filtered.length)); setFlipped(false); }} style={{
        marginTop: 12, padding: "8px 20px", borderRadius: 8, border: "1px solid #30363d",
        background: "transparent", color: "#8b949e", cursor: "pointer", fontSize: 12
      }}>🔀 Random Card</button>
    </div>
  );
}