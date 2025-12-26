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
