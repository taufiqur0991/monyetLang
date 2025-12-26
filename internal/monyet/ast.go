package monyet

type Node interface{}

type Program struct {
	Statements []Node
}

type Number struct {
	Value int
}

type Variable struct {
	Name string
}

type Binary struct {
	Left  Node
	Op    TokenType
	Right Node
}

type Assign struct {
	Name  string
	Value Node
}

type Echo struct {
	Value Node
}

type String struct {
	Value string
}

type Function struct {
	Name   string
	Params []string
	Body   []Node
}

type Call struct {
	Name string
	Args []Node
}

type Return struct {
	Value Node
}

type If struct {
	Condition Node
	Then      []Node
	Else      []Node // Bisa nil kalau tidak ada else
}

type Serve struct {
	Port    Node
	Handler string
}
