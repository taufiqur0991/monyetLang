package monyet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var activeStorage *MonyetDB
var activeDBPath string

func Eval(prog *Program, env *Env) {
	for _, s := range prog.Statements {
		evalNode(s, env)
	}
}

type returnValue struct {
	value interface{}
}

func evalNode(n Node, env *Env) interface{} {
	switch v := n.(type) {

	case Number:
		return v.Value

	case String:
		return v.Value

	case Variable:
		val, ok := env.GetVar(v.Name)
		if !ok {
			prefixes := []string{"GET_", "POST_", "PATCH_", "DELETE_"}
			for _, p := range prefixes {
				if strings.HasPrefix(v.Name, p) {
					return ""
				}
			}
			panic("undefined variable: $" + v.Name)
		}
		return val

	case Assign:
		val := evalNode(v.Value, env)
		env.SetVar(v.Name, val)
		return val

	case Binary:
		if v.Op == AND || v.Op == OR {
			lVal, okL := evalNode(v.Left, env).(bool)
			rVal, okR := evalNode(v.Right, env).(bool)
			if !okL || !okR {
				panic("Operasi logika && atau || membutuhkan tipe boolean")
			}
			if v.Op == AND {
				return lVal && rVal
			}
			return lVal || rVal
		}

		// 2. Evaluasi Nilai Kiri dan Kanan untuk operasi lainnya
		left := evalNode(v.Left, env)
		right := evalNode(v.Right, env)

		// 3. Operasi Perbandingan EQ & Penggabungan String (PLUS)
		switch v.Op {
		case EQ:
			sLeft, okL := left.(string)
			sRight, okR := right.(string)
			if okL && okR {
				return strings.TrimSpace(sLeft) == strings.TrimSpace(sRight)
			}
			return left == right

		case PLUS:
			// Cek kalau salah satu string, lakukan Concat (PHP-style)
			_, lok := left.(string)
			_, rok := right.(string)
			if lok || rok {
				return fmt.Sprintf("%v%v", left, right)
			}
		}

		// 4. Operasi Matematika & Perbandingan Angka (Khusus Int)
		li, lok := left.(float64)
		ri, rok := right.(float64)
		if !lok || !rok {
			panic(fmt.Sprintf("invalid operands for binary operation: %v (%T) and %v (%T)", left, left, right, right))
		}

		switch v.Op {
		case PLUS:
			return li + ri
		case MINUS:
			return li - ri
		case STAR:
			return li * ri
		case SLASH:
			if ri == 0 {
				panic("pembagian dengan nol")
			}
			return li / ri
		case GT:
			return li > ri
		}

	case Echo:
		val := evalNode(v.Value, env)
		fmt.Println(val)
		return nil

	case Function:
		env.SetFunc(v.Name, v)
		return nil

	case Call:
		getStorage := func() *MonyetDB {
			base, _ := env.GetVar("__BASE_DIR__")
			dbPath := filepath.Join(base.(string), "monyet.db")

			if customName, ok := env.GetVar("DB_NAME"); ok {
				dbPath = filepath.Join(base.(string), fmt.Sprintf("%v", customName))
			}

			// Jika path-nya masih sama dan storage sudah ada, pakai yang lama saja
			if activeStorage != nil && activeDBPath == dbPath {
				return activeStorage
			}

			// Jika ganti nama DB atau baru pertama kali, buka koneksi baru
			activeDBPath = dbPath
			activeStorage = NewMonyetDB(dbPath)
			return activeStorage
		}
		if v.Name == "set_data" {
			k := fmt.Sprintf("%v", evalNode(v.Args[0], env))
			val := evalNode(v.Args[1], env)

			// Jika yang dikirim adalah Map atau Array, otomatis JSON-kan
			var finalVal interface{}
			switch val.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, _ := json.Marshal(val)
				finalVal = string(jsonBytes)
			default:
				finalVal = val
			}

			getStorage().Set(k, finalVal)
			return true
		}

		if v.Name == "get_data" {
			kEval := evalNode(v.Args[0], env)
			if kEval == nil {
				return ""
			}

			kStr := fmt.Sprintf("%v", kEval)
			return getStorage().Get(kStr)
		}
		if v.Name == "delete_data" {
			if len(v.Args) < 1 {
				return false
			}
			k := fmt.Sprintf("%v", evalNode(v.Args[0], env))
			if k != "" && k != "<nil>" {
				getStorage().Delete(k)
				return true
			}
			return false
		}
		if v.Name == "drop_db" {
			// Memanggil fungsi Drop() untuk menghapus file fisik database
			err := getStorage().Drop()
			if err != nil {
				fmt.Printf("Gagal menghapus database: %v\n", err)
				return false
			}
			return true
		}
		fn, ok := env.GetFunc(v.Name)
		if !ok {
			panic("undefined function: " + v.Name)
		}
		//fmt.Println("Memanggil fungsi:", v.Name, "dengan args:", v.Args)

		local := NewChildEnv(env)

		for i, p := range fn.Params {
			local.SetVar(p, evalNode(v.Args[i], env))
		}

		for _, stmt := range fn.Body {
			val := evalNode(stmt, local)
			if rv, ok := val.(returnValue); ok {
				return rv.value
			}
		}
		return nil

	case Return:
		val := evalNode(v.Value, env)
		return returnValue{value: val}
	case If:
		cond := evalNode(v.Condition, env)
		isTrue := false
		if b, ok := cond.(bool); ok {
			isTrue = b
		} else if i, ok := cond.(int); ok {
			isTrue = i > 0
		}

		if isTrue {
			for _, stmt := range v.Then {
				res := evalNode(stmt, env)
				// TAMBAHKAN INI: Jika ada return di dalam IF, teruskan ke atas
				if rv, ok := res.(returnValue); ok {
					return rv
				}
			}
		} else if v.Else != nil {
			for _, stmt := range v.Else {
				res := evalNode(stmt, env)
				// TAMBAHKAN INI: Jika ada return di dalam ELSE, teruskan ke atas
				if rv, ok := res.(returnValue); ok {
					return rv
				}
			}
		}
		return nil
	case IndexAccess:
		left := evalNode(v.Left, env)
		index := evalNode(v.Index, env)

		if m, ok := left.(map[string]interface{}); ok {
			idxStr := fmt.Sprintf("%v", index)
			return m[idxStr]
		}
		// TAMBAHKAN INI: Support untuk Array/Slice
		if s, ok := left.([]interface{}); ok {
			if idxInt, ok := index.(int); ok {
				if idxInt >= 0 && idxInt < len(s) {
					return s[idxInt]
				}
			}
		}
		return nil
	case Serve:
		portVal := evalNode(v.Port, env)
		handlerName := v.Handler
		addr := fmt.Sprintf("0.0.0.0:%v", portVal)

		// Gunakan server mux lokal agar tidak bentrok jika serve dipanggil ulang
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fn, ok := env.GetFunc(handlerName)
			if !ok {
				http.Error(w, fmt.Sprintf("Handler %s tidak ditemukan", handlerName), 404)
				return
			}

			local := NewChildEnv(env)

			// Setup variabel superglobal
			monyetGet := make(map[string]interface{})
			for key, values := range r.URL.Query() {
				monyetGet[key] = values[0]
				local.SetVar("GET_"+strings.ToUpper(key), values[0])
			}
			local.SetVar("_GET", monyetGet)
			// --- 2. HANDLE REQUEST BODY (_POST, _PATCH, _DELETE) ---
			// Kita asumsikan body yang dikirim adalah JSON
			bodyData := make(map[string]interface{})
			if r.Body != nil {
				decoder := json.NewDecoder(r.Body)
				// Kita tidak panic kalau decode gagal (misal body kosong)
				_ = decoder.Decode(&bodyData)
			}
			local.SetVar("_POST", bodyData)
			local.SetVar("_PATCH", bodyData)
			local.SetVar("_DELETE", bodyData)
			// Opsional: shortcut seperti POST_NAMA
			for k, v := range bodyData {
				prefix := r.Method + "_" // Akan jadi POST_, PATCH_, atau DELETE_
				local.SetVar(prefix+strings.ToUpper(k), v)
			}
			local.SetVar("PATH", r.URL.Path)
			local.SetVar("METHOD", r.Method)

			var result interface{} = "" // Default kosong agar tidak <nil>
			for _, stmt := range fn.Body {
				val := evalNode(stmt, local)
				if rv, ok := val.(returnValue); ok {
					result = rv.value
					break
				}
			}

			// --- HANDLING OUTPUT & HEADER ---
			resStr := fmt.Sprintf("%v", result)

			// Deteksi JSON secara otomatis
			if len(resStr) > 0 && (resStr[0] == '{' || resStr[0] == '[') {
				w.Header().Set("Content-Type", "application/json")
			} else {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
			}

			fmt.Fprint(w, resStr)
		})
		err := http.ListenAndServe(addr, mux)
		if err != nil {
			fmt.Printf("Web Server Gagal: %v\n", err) // Tambahkan log ini
		}
		return nil
	case Include:
		// Ambil base directory dari environment
		baseDirVal, _ := env.GetVar("__BASE_DIR__")
		baseDir := baseDirVal.(string)

		// Gabungkan: BaseDir Script + Path di dalam script
		targetPath := filepath.Join(baseDir, v.Path)

		content, err := os.ReadFile(targetPath)
		if err != nil {
			panic(fmt.Sprintf("Gagal include: %s", targetPath))
		}

		l := NewLexer(string(content))
		p := NewParser(l)
		newProg := p.Parse()

		for _, stmt := range newProg.Statements {
			evalNode(stmt, env)
		}
		return nil

	case Render:
		pathVal := evalNode(v.Path, env).(string)

		baseDirVal, _ := env.GetVar("__BASE_DIR__")
		baseDir := baseDirVal.(string)

		targetPath := filepath.Join(baseDir, pathVal)

		content, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Sprintf("Render Error: File %s tidak ditemukan", targetPath)
		}
		return string(content)
	case JsonEncode:
		dataVal := evalNode(v.Data, env)
		jsonBytes, err := json.Marshal(dataVal)
		if err != nil {
			return fmt.Sprintf(`{"error": "%v"}`, err)
		}
		return string(jsonBytes)

	case JsonDecode:
		strVal := evalNode(v.Value, env)
		str, ok := strVal.(string)
		if !ok {
			panic("json_decode membutuhkan input string")
		}

		var result interface{}
		err := json.Unmarshal([]byte(str), &result)
		if err != nil {
			return nil
		}
		return result
	case MapLiteral:
		res := make(map[string]interface{})
		for k, v := range v.Pairs {
			// Evaluasi key dan value menjadi tipe data Go asli (string, int, dll)
			keyEval := evalNode(k, env)
			valEval := evalNode(v, env)

			// Paksa key menjadi string agar aman untuk JSON
			keyStr := fmt.Sprintf("%v", keyEval)
			res[keyStr] = valEval
		}
		return res
	case ForeachStatement:
		iter := evalNode(v.Iterable, env)

		// Cek jika iterabel adalah Slice/Array
		if list, ok := iter.([]interface{}); ok {
			for i, item := range list {
				local := NewChildEnv(env)
				if v.Key != "" {
					local.SetVar(v.Key, float64(i)) // index sebagai angka
				}
				local.SetVar(v.Value, item)

				for _, stmt := range v.Body {
					res := evalNode(stmt, local)
					if rv, ok := res.(returnValue); ok {
						return rv
					}
				}
			}
			return nil
		}

		// Cek jika iterabel adalah Map
		if m, ok := iter.(map[string]interface{}); ok {
			for k, val := range m {
				local := NewChildEnv(env)
				if v.Key != "" {
					local.SetVar(v.Key, k)
				}
				local.SetVar(v.Value, val)

				for _, stmt := range v.Body {
					res := evalNode(stmt, local)
					if rv, ok := res.(returnValue); ok {
						return rv
					}
				}
			}
			return nil
		}
		//end of switch
	}

	return nil
}
