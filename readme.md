# 🐵 MonyetLang & MonyetDB
**Lightweight Scripting Engine with Native Web Server & Log-Structured Database.**

MonyetLang is an interpreted programming language built entirely from scratch using Go. This project is specifically designed for **Edge Computing** scenarios and resource-constrained hardware, where you need a fast web server and data storage without heavy overhead.

## 🌟 Key Features

### 1. MonyetDB (Custom Storage Engine)
Unlike traditional databases, MonyetDB utilizes a **Log-Structured Storage** architecture:
* **Append-Only Logging**: Ensures data integrity. Existing data is never overwritten; all changes are recorded as new log entries.
* **In-Memory Hash Indexing**: Maps every *key* to its *byte offset* on the disk. Data retrieval is $O(1)$ because the system performs a direct `Seek` to the data position without scanning the entire file.
* **Crash Resilience**: Thanks to its *append-only* nature, historical data remains safe even if a system failure occurs during a write operation.

### 2. Built-in Web Server
MonyetLang features an internal HTTP server that supports:
* **RESTful Routing**: Effortlessly handle `$PATH` and `$METHOD` (GET, POST, PATCH, DELETE).
* **Superglobal Variables**: Easy access to request data via `$_GET`, `$_POST`, `$_PATCH`, and `$_DELETE`.
* **Native JSON Support**: Built-in integration for `json_encode` and `json_decode`.

## 🛠️ Installation & Usage

### Build the Runner
Compile the Go source code into a binary file:
```bash
go build -o monyet.exe ./cmd/monyet
```

### Run a Script
Use the runner to execute a .nyet file:
```bash
./monyet.exe examples/test.nyet
```

### Start the Web Server
The server will run on port 8080 (default) and is ready to serve API requests and database operations:
```bash
./monyet.exe examples/server.nyet
```

### 📝 Script Example (server.nyet)
```PHP
function router() {
    // 1. Endpoint Homepage
    if ($PATH == "/") {
        return "<h1>Welcome to MonyetLang Server</h1>";
    }

    // 2. API Endpoint with Database Integration
    if ($PATH == "/api/data") {
        if ($METHOD == "POST") {
            $val = $_POST["nama"];
            set_data("user_terakhir", $val);
            return json_encode(["status" => "saved", "data" => $val]);
        }
        
        if ($METHOD == "GET") {
            $saved = get_data("user_terakhir");
            return json_encode(["key" => "user_terakhir", "value" => $saved]);
        }
    }

    return "404 Not Found";
}

serve(8080, router);
```

### 🏗️ Project Structure
- `/cmd/monyet`: Application entry point.
- `/internal/monyet`: Core engine (Lexer, Parser, Interpreter, DB).
- `examples/`: Collection of example scripts.

### 📷 Screenshoot
![Screenshoot](screenshoot.png)