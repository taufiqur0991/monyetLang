package monyet

type Env struct {
	vars  map[string]interface{}
	funcs map[string]Function
	outer *Env
}

func NewEnv() *Env {
	return &Env{
		vars:  make(map[string]interface{}),
		funcs: make(map[string]Function),
	}
}

func NewChildEnv(outer *Env) *Env {
	return &Env{
		vars:  make(map[string]interface{}),
		funcs: outer.funcs, // share functions
		outer: outer,
	}
}

func (e *Env) GetVar(name string) (interface{}, bool) {
	if v, ok := e.vars[name]; ok {
		return v, true
	}
	if e.outer != nil {
		return e.outer.GetVar(name)
	}
	return nil, false
}

func (e *Env) SetVar(name string, val interface{}) {
	e.vars[name] = val
}

func (e *Env) GetFunc(name string) (Function, bool) {
	fn, ok := e.funcs[name]
	return fn, ok
}

func (e *Env) SetFunc(name string, fn Function) {
	e.funcs[name] = fn
}
