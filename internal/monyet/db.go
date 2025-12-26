package monyet

import (
	"bufio"
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

	f, _ := os.Open(db.path)
	defer f.Close()
	f.Seek(offset, 0)

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			// JIKA nilainya __DELETED__, kembalikan string kosong
			if parts[1] == "__DELETED__" {
				return ""
			}
			return parts[1]
		}
	}
	return ""
}

// Tambahkan fungsi ini di db.go
func (db *MonyetDB) Delete(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Tulis baris baru yang menandakan key ini sudah dihapus
	line := fmt.Sprintf("%s:__DELETED__\n", key)

	info, _ := db.file.Stat()
	offset := info.Size()

	_, err := db.file.WriteString(line)
	if err == nil {
		// Update index di RAM: pindahkan offset ke baris DELETED ini
		db.index[key] = offset
		db.file.Sync()
	}
	return err
}

func (db *MonyetDB) Drop() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.file.Close()                   // Tutup koneksi file
	db.index = make(map[string]int64) // Kosongkan index di RAM
	return os.Remove(db.path)         // Hapus file dari folder
}
