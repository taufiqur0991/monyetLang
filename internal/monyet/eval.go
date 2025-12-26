package monyet

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

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
			// Jika variabel diawali GET_, jangan panic, kembalikan string kosong
			if strings.HasPrefix(v.Name, "GET_") {
				return ""
			}
			panic("undefined variable: $" + v.Name)
		}
		return val

	case Assign:
		val := evalNode(v.Value, env)
		env.SetVar(v.Name, val)
		return val

	case Binary:
		left := evalNode(v.Left, env)
		right := evalNode(v.Right, env)

		// 1. Operasi Perbandingan (Bisa String atau Int)
		switch v.Op {
		case AND:
			left := evalNode(v.Left, env).(bool)
			right := evalNode(v.Right, env).(bool)
			return left && right
		case OR:
			left := evalNode(v.Left, env).(bool)
			right := evalNode(v.Right, env).(bool)
			return left || right
		case EQ:
			sLeft, okL := left.(string)
			sRight, okR := right.(string)

			if okL && okR {
				// Bersihkan spasi/newline di kedua sisi
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

		// 2. Operasi Matematika & Perbandingan Angka (Khusus Int)
		li, lok := left.(int)
		ri, rok := right.(int)
		if !lok || !rok {
			// Biar tidak bingung, kasih info lebih detail di panic
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

		// Cek apakah left adalah sebuah Map (untuk $_GET)
		if m, ok := left.(map[string]interface{}); ok {
			idxStr := fmt.Sprintf("%v", index)
			return m[idxStr]
		}
		return nil
	case Serve:
		portVal := evalNode(v.Port, env)
		handlerName := v.Handler
		addr := fmt.Sprintf("0.0.0.0:%v", portVal)

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fn, ok := env.GetFunc(handlerName)
			if !ok {
				return
			}
			monyetGet := make(map[string]interface{})
			for key, values := range r.URL.Query() {
				monyetGet[key] = values[0]
			}

			local := NewChildEnv(env)
			local.SetVar("_GET", monyetGet)
			local.SetVar("PATH", r.URL.Path)
			local.SetVar("METHOD", r.Method)

			// Memasukkan semua query params secara otomatis
			for key, values := range r.URL.Query() {
				local.SetVar("GET_"+strings.ToUpper(key), values[0])
			}

			var result interface{}
			for _, stmt := range fn.Body {
				val := evalNode(stmt, local)
				if rv, ok := val.(returnValue); ok {
					result = rv.value
					break
				}
			}
			fmt.Fprintf(w, "%v", result)
		})
		err := http.ListenAndServe(addr, nil)
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

		// PENTING: Jika file yang di-include berada di subfolder,
		// kita harus update __BASE_DIR__ untuk file tersebut agar include di dalamnya juga jalan.
		// Tapi untuk tahap awal, ini sudah cukup.

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
		//end of switch
	}

	return nil
}
