package monyet

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

type MonyetDB struct {
	file  *os.File
	path  string
	mu    sync.RWMutex
	index map[string]int64 // Menyimpan offset byte terakhir untuk setiap key
}

func NewMonyetDB(path string) *MonyetDB {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	db := &MonyetDB{
		file:  f,
		path:  path,
		index: make(map[string]int64),
	}
	db.loadIndex() // Bangun index saat database dinyalakan
	return db
}

func (db *MonyetDB) loadIndex() {
	f, _ := os.Open(db.path)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var offset int64 = 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			db.index[parts[0]] = offset
		}
		// Tambahkan panjang karakter + 1 (untuk \n) ke offset
		offset += int64(len(line) + 1)
	}
}

func (db *MonyetDB) Set(key string, value interface{}) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	valStr := fmt.Sprintf("%v", value)
	line := fmt.Sprintf("%s:%s\n", key, valStr)

	// Ambil posisi sebelum menulis (ini adalah offset baris baru)
	info, _ := db.file.Stat()
	offset := info.Size()

	_, err := db.file.WriteString(line)
	if err == nil {
		db.index[key] = offset // Update index di RAM
		db.file.Sync()
	}
	return err
}

func (db *MonyetDB) Get(key string) string {
	db.mu.RLock()
	offset, exists := db.index[key]
	db.mu.RUnlock()

	if !exists {
		return ""
	}

	// Langsung lompat ke posisi byte yang dicatat
	f, _ := os.Open(db.path)
	defer f.Close()

	f.Seek(offset, 0)

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return ""
}

func (db *MonyetDB) SetJSON(key string, value interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// 1. Baca seluruh file JSON yang ada ke dalam Map Go
	data := make(map[string]interface{})
	content, _ := os.ReadFile(db.path)
	json.Unmarshal(content, &data)

	// 2. Tambah/Update data baru
	data[key] = value

	// 3. Tulis ulang seluruh file sebagai JSON yang rapi
	newContent, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(db.path, newContent, 0644)
}
