package ast

// Visit() is called multiple times per node. The visitType tells the handler which one
// of the multiple calls the current one is
const (
	// This visit represents the beginning of the node
	PreVisit = -1
	// This visit represents the end of the node
	PostVisit = -2
	// This visit is the only one for this node
	SingleVisit = -3
	// This visit is the first intermediate visit of the node
	InterVisit1 = -4
	// This visit is the second intermediate visit of the node
	InterVisit2 = -5
)

// Acceptor MUST be implemented by EVERY AST-Node. The node must do the following things:
// - call Visit(node,Previsit) on the visitor
// - call Accept() on every one of it's children
// - if it has a list of children it MUST call Visit(node,i) before calling accept() on the i-th element of the list
// - if the node has multiple children (e.g. if.statement(condition,ifblock,elseblock))
// it must call Visit(node,InterVisitX) between the different kinds of children
// - it must call Visit(node,PostVisit) after accept() has been called on all children
// if the node has no children is MUST call Visit(node,SingleVisit) and NOTHING ELSE
type Acceptor interface {
	Accept(v Visitor) error
}

// Visitor defines an interface that allows easy node-agnostic traversal of the AST
type Visitor interface {
	Visit(node Node, visitType int) error
}

// VisitorFunc allows simple functions to be used as Visitor
type VisitorFunc func(node Node, visitType int) error

// Visit is called by the currently visited AST-Node
// The Visitor has to use a type-assertion to find out what kind of node it is
// visitType tells the visitor which one of the multiple-visits for this node the current one is.
// If the Visitor is used as argument to Acceptor.Accept() visit is called (multiple times) for
// every AST-Node in the sub-tree starting at the acceptor. It can then modify these nodes.
// By return the NodeReplacement type as the 'error' the currently visited node can be replaced by another node.
func (f VisitorFunc) Visit(node Node, visitType int) error {
	return f(node, visitType)
}

// NodeReplacement special error type. If this type is returned by an AST-child during Accept()
// the parent-node MUST react to this by replacing the child with the nodes in Replacement[] and
// then discard the error (and NOT relay it to its parent)
type NodeReplacement struct {
	Replacement []Node
}

// NewNodeReplacement is used to replace the current node
func NewNodeReplacement(replacement ...Node) NodeReplacement {
	return NodeReplacement{
		Replacement: replacement,
	}
}

// Error() must be implemented for the error-interface
// Should NEVER be called
func (e NodeReplacement) Error() string {
	return "SHOULD NEVER HAPPEN"
}

// Accept is used to implement Acceptor
func (p *Program) Accept(v Visitor) error {
	err := v.Visit(p, PreVisit)
	if err != nil {
		return err
	}
	for i := 0; i < len(p.Lines); i++ {
		err = v.Visit(p, i)
		if err != nil {
			return err
		}
		err = p.Lines[i].Accept(v)
		// if the child wants to be replaced, replace it
		if repl, is := err.(NodeReplacement); is {
			p.Lines = PatchLines(p.Lines, i, repl)
			i += len(repl.Replacement) - 1
			err = nil
		}
		if err != nil {
			return err
		}
	}
	return v.Visit(p, PostVisit)
}

// Accept is used to implement Acceptor
func (l *Line) Accept(v Visitor) error {
	err := v.Visit(l, PreVisit)
	if err != nil {
		return err
	}
	for i := 0; i < len(l.Statements); i++ {
		err = v.Visit(l, i)
		if err != nil {
			return err
		}
		err = l.Statements[i].Accept(v)
		if repl, is := err.(NodeReplacement); is {
			l.Statements = PatchStatements(l.Statements, i, repl)
			i += len(repl.Replacement) - 1
			err = nil
		}
		if err != nil {
			return err
		}
	}
	return v.Visit(l, PostVisit)
}

// Accept is used to implement Acceptor
func (c *StringConstant) Accept(v Visitor) error {
	return v.Visit(c, SingleVisit)
}

// Accept is used to implement Acceptor
func (c *NumberConstant) Accept(v Visitor) error {
	return v.Visit(c, SingleVisit)
}

// Accept is used to implement Acceptor
func (d *Dereference) Accept(v Visitor) error {
	return v.Visit(d, SingleVisit)
}

// Accept is used to implement Acceptor
func (u *UnaryOperation) Accept(v Visitor) error {
	err := v.Visit(u, PreVisit)
	if err != nil {
		return err
	}
	err = u.Exp.Accept(v)
	if repl, is := err.(NodeReplacement); is {
		u.Exp = repl.Replacement[0].(Expression)
		err = nil
	}
	if err != nil {
		return err
	}
	return v.Visit(u, PostVisit)
}

// Accept is used to implement Acceptor
func (o *BinaryOperation) Accept(v Visitor) error {
	err := v.Visit(o, PreVisit)
	if err != nil {
		return err
	}
	err = o.Exp1.Accept(v)
	if repl, is := err.(NodeReplacement); is {
		o.Exp1 = repl.Replacement[0].(Expression)
		err = nil
	}
	if err != nil {
		return err
	}
	err = v.Visit(o, InterVisit1)
	err = o.Exp2.Accept(v)
	if repl, is := err.(NodeReplacement); is {
		o.Exp2 = repl.Replacement[0].(Expression)
		err = nil
	}
	if err != nil {
		return err
	}
	return v.Visit(o, PostVisit)
}

// Accept is used to implement Acceptor
func (f *FuncCall) Accept(v Visitor) error {
	err := v.Visit(f, PreVisit)
	if err != nil {
		return err
	}
	if f.Argument != nil {
		err = f.Argument.Accept(v)
		if repl, is := err.(NodeReplacement); is {
			f.Argument = repl.Replacement[0].(Expression)
			err = nil
		}
		if err != nil {
			return err
		}
	}
	return v.Visit(f, PostVisit)
}

// Accept is used to implement Acceptor
func (a *Assignment) Accept(v Visitor) error {
	err := v.Visit(a, PreVisit)
	if err != nil {
		return err
	}
	err = a.Value.Accept(v)
	if repl, is := err.(NodeReplacement); is {
		a.Value = repl.Replacement[0].(Expression)
		err = nil
	}
	if err != nil {
		return err
	}
	return v.Visit(a, PostVisit)
}

// Accept is used to implement Acceptor
func (s *IfStatement) Accept(v Visitor) error {
	err := v.Visit(s, PreVisit)
	if err != nil {
		return err
	}
	err = s.Condition.Accept(v)
	if repl, is := err.(NodeReplacement); is {
		s.Condition = repl.Replacement[0].(Expression)
		err = nil
	}
	if err != nil {
		return err
	}
	err = v.Visit(s, InterVisit1)
	if err != nil {
		return err
	}
	for i := 0; i < len(s.IfBlock); i++ {
		err = v.Visit(s, i)
		if err != nil {
			return err
		}
		err = s.IfBlock[i].Accept(v)
		if repl, is := err.(NodeReplacement); is {
			s.IfBlock = PatchStatements(s.IfBlock, i, repl)
			i += len(repl.Replacement) - 1
			err = nil
		}
		if err != nil {
			return err
		}
	}
	if s.ElseBlock != nil {
		err = v.Visit(s, InterVisit2)
		if err != nil {
			return err
		}
		for i := 0; i < len(s.ElseBlock); i++ {
			err = v.Visit(s, i)
			if err != nil {
				return err
			}
			err = s.ElseBlock[i].Accept(v)
			if repl, is := err.(NodeReplacement); is {
				s.ElseBlock = PatchStatements(s.ElseBlock, i, repl)
				i += len(repl.Replacement) - 1
				err = nil
			}
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
	return v.Visit(s, PostVisit)
}

// Accept is used to implement Acceptor
func (g *GoToStatement) Accept(v Visitor) error {
	return v.Visit(g, SingleVisit)
}

// PatchLines is used to replace the element at position in old with the elements in repl
func PatchLines(old []*Line, position int, repl NodeReplacement) []*Line {
	newv := make([]*Line, 0, len(old)+len(repl.Replacement)-1)
	newv = append(newv, old[:position]...)
	for _, elem := range repl.Replacement {
		if line, is := elem.(*Line); is {
			newv = append(newv, line)
		} else {
			panic("Could not patch slice. Wrong type.")
		}
	}
	newv = append(newv, old[position+1:]...)
	return newv
}

// PatchStatements is used to replace the element at position in old with the elements in repl
func PatchStatements(old []Statement, position int, repl NodeReplacement) []Statement {
	newv := make([]Statement, 0, len(old)+len(repl.Replacement)-1)
	newv = append(newv, old[:position]...)
	for _, elem := range repl.Replacement {
		if line, is := elem.(Statement); is {
			newv = append(newv, line)
		} else {
			panic("Could not patch slice. Wrong type.")
		}
	}
	newv = append(newv, old[position+1:]...)
	return newv
}
