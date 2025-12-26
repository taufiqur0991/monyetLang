package monyet

import "fmt"

func Eval(prog *Program, env *Env) {
	for _, s := range prog.Statements {
		evalNode(s, env)
	}
}

func evalNode(n Node, env *Env) interface{} {
	switch v := n.(type) {

	case Number:
		return v.Value

	case String:
		return v.Value

	case Variable:
		val, ok := env.vars[v.Name]
		if !ok {
			panic("undefined variable: $" + v.Name)
		}
		return val

	case Assign:
		val := evalNode(v.Value, env)
		env.vars[v.Name] = val
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
	}

	return nil
}
