package monyet

type Node interface{}

type Program struct {
	Statements []Node
}

type Number struct {
	Value float64
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

type IndexAccess struct {
	Left  Node
	Index Node
}

type Include struct {
	Path string
}

type Render struct {
	Path Node
}

type JsonEncode struct {
	Data Node
}

type JsonDecode struct {
	Value Node
}

type MapLiteral struct {
	Pairs map[Node]Node
}
type ForeachStatement struct {
	Iterable Node
	Key      string
	Value    string
	Body     []Node
}

//func (f ForeachStatement) nodeSigil() {}

// func (m MapLiteral) nodeSig() {}
