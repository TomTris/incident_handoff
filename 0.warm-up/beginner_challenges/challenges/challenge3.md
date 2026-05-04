# The File Word Counter
## Problem
You have a text file. You need to read it and count how many bytes were actually read — then print both the raw bytes count and the content.

## Scope
- Create a file called story.txt or choose any existing file with at least 2 lines of any text
- Define a const for your buffer size (e.g. BufferSize = 1) — use a small number so the loop actually runs multiple times
- Write a Go program that opens the file safely
- Use a for loop to read the file chunk by chunk until it's fully read
- Count the total bytes read across all chunks
- Print the total bytes and the full content (no extra null bytes!)
- Handle ALL errors properly

## Expected Output (Example)
- Bytes read: 43
- Content: "Hello,
my name is Tom.
I am learning Go"

## Hints
os.Open() to open the file
make([]byte, 1024) for the buffer
f.Read(buf) returns n (bytes read)
string(buf[:n]) to convert to readable text
defer f.Close() to close safely
log.Fatal(err) to handle errors