package monyet

type Env struct {
	vars map[string]interface{}
}

func NewEnv() *Env {
	return &Env{vars: make(map[string]interface{})}
}
