# 🐵 MonyetLang & MonyetDB
**Lightweight Scripting Engine with Native Web Server & Log-Structured Database.**

MonyetLang adalah bahasa pemrograman interpretatif yang dibuat murni menggunakan Go. Proyek ini dirancang untuk skenario **Edge Computing** atau hardware dengan resource terbatas, di mana kamu butuh web server dan penyimpanan data yang cepat tanpa overhead yang besar.

## 🌟 Fitur Unggulan

### 1. MonyetDB (Custom Storage Engine)
Berbeda dengan database tradisional, MonyetDB menggunakan arsitektur **Log-Structured Storage**:
* **Append-Only Logging**: Menjamin integritas data. Tidak ada data yang ditimpa, semua perubahan dicatat sebagai log baru.
* **In-Memory Hash Indexing**: Memetakan setiap *key* ke *byte offset* di disk. Pencarian data bersifat O(1) karena sistem langsung melakukan `Seek` ke posisi data tanpa scanning file.
* **Crash Resilience**: Karena sistemnya *append-only*, data lama tetap aman meskipun terjadi kegagalan sistem saat penulisan.

### 2. Built-in Web Server
MonyetLang memiliki server HTTP internal yang mendukung:
* **RESTful Routing**: Menangani `$PATH` dan `$METHOD` (GET, POST, PATCH, DELETE).
* **Superglobal Variables**: Akses mudah ke data request melalui `$_GET`, `$_POST`, `$_PATCH`, dan `$_DELETE`.
* **JSON Native**: Integrasi langsung dengan `json_encode` dan `json_decode`.

## 🛠️ Instalasi & Penggunaan

### Build Runner
Kompilasi source code Go menjadi file binary:
```bash
go build -o monyet.exe ./cmd/monyet
```

### Menjalankan Script
Gunakan runner untuk mengeksekusi file .nyet:
```bash
./monyet.exe examples/test.nyet
```

### Menjalankan Web Server
Server akan berjalan di port 8080 (default) dan siap melayani request API serta database:
```bash
./monyet.exe examples/server.nyet
```

### 📝 Contoh Script (server.nyet)
```PHP
function router() {
    // 1. Endpoint Homepage
    if ($PATH == "/") {
        return "<h1>Welcome to MonyetLang Server</h1>";
    }

    // 2. Endpoint API dengan Database
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

### 🏗️ Struktur Proyek
- `/cmd/monyet`: Entry point aplikasi.
- `/internal/monyet`: Core engine (Lexer, Parser, Interpreter, DB).
- `examples/`: Koleksi script contoh penggunaan.

### 📷 Screenshoot
![Screenshoot][screenshoot.png]