package monyet

import "fmt"

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

		// STRING CONCAT (PHP-style)
		if v.Op == PLUS {
			_, lok := left.(string)
			_, rok := right.(string)
			if lok || rok {
				return fmt.Sprintf("%v%v", left, right)
			}
		}

		// INTEGER OPS
		li, lok := left.(int)
		ri, rok := right.(int)
		if !lok || !rok {
			panic("invalid operands for binary operation")
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

		// Pastikan hasil Binary GT (>) adalah boolean
		isTrue := false
		if b, ok := cond.(bool); ok {
			isTrue = b
		} else if i, ok := cond.(int); ok {
			isTrue = i > 0
		}

		if isTrue {
			for _, stmt := range v.Then {
				evalNode(stmt, env) // Jalankan setiap statement di dalam { }
			}
		} else if v.Else != nil {
			for _, stmt := range v.Else {
				evalNode(stmt, env) // Jalankan blok else jika ada
			}
		}
		return nil
	}

	return nil
}
