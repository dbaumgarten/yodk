package ast

import (
	"fmt"
)

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
	// This visit is the third intermediate visit of the node
	InterVisit3 = -6
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
// Also, the new replacement node must be visited again, if Skip==false. If this new visit also reslults in a replacement,
// the new replacement must be visited again and so on, until no replacement is received anymore.
type NodeReplacement struct {
	Replacement []Node
	// if skip is true, do not re-visit the replaced node
	Skip bool
}

// NewNodeReplacement is used to replace the current node
func NewNodeReplacement(replacement ...Node) NodeReplacement {
	return NodeReplacement{
		Replacement: replacement,
	}
}

// NewNodeReplacementSkip is used to replace the current node and signals not to re-visit the new node
func NewNodeReplacementSkip(replacement ...Node) NodeReplacement {
	return NodeReplacement{
		Replacement: replacement,
		Skip:        true,
	}
}

// Error() must be implemented for the error-interface
// Should NEVER be called
func (e NodeReplacement) Error() string {
	if len(e.Replacement) > 0 {
		return fmt.Sprintf("INTERNAL COMPILER ERROR. Node-replacement for %T bubbeled up the ast", e.Replacement[0])
	}
	return fmt.Sprintf("INTERNAL COMPILER ERROR. Node-deletion bubbeled up the ast")
}

// Accept is used to implement Acceptor
func (p *Program) Accept(v Visitor) error {
	err := v.Visit(p, PreVisit)
	if err != nil {
		return err
	}

	if p == nil {
		return v.Visit(p, PostVisit)
	}

	p.Lines, err = AcceptChildLines(p, v, p.Lines)
	if err != nil {
		return err
	}

	return v.Visit(p, PostVisit)
}

// Accept is used to implement Acceptor
func (l *Line) Accept(v Visitor) error {
	err := v.Visit(l, PreVisit)
	if err != nil {
		return err
	}

	l.Statements, err = AcceptChildStatements(l, v, l.Statements)
	if err != nil {
		return err
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
	u.Exp, err = AcceptChild(v, u.Exp)
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
	o.Exp1, err = AcceptChild(v, o.Exp1)
	if err != nil {
		return err
	}
	err = v.Visit(o, InterVisit1)
	o.Exp2, err = AcceptChild(v, o.Exp2)
	if err != nil {
		return err
	}
	return v.Visit(o, PostVisit)
}

// Accept is used to implement Acceptor
func (a *Assignment) Accept(v Visitor) error {
	err := v.Visit(a, PreVisit)
	if err != nil {
		return err
	}
	a.Value, err = AcceptChild(v, a.Value)
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
	s.Condition, err = AcceptChild(v, s.Condition)
	if err != nil {
		return err
	}
	err = v.Visit(s, InterVisit1)
	if err != nil {
		return err
	}

	s.IfBlock, err = AcceptChildStatements(s, v, s.IfBlock)
	if err != nil {
		return err
	}

	if s.ElseBlock != nil {
		err = v.Visit(s, InterVisit2)
		if err != nil {
			return err
		}
		s.ElseBlock, err = AcceptChildStatements(s, v, s.ElseBlock)
		if err != nil {
			return err
		}
	}
	return v.Visit(s, PostVisit)
}

// Accept is used to implement Acceptor
func (g *GoToStatement) Accept(v Visitor) error {
	err := v.Visit(g, PreVisit)
	if err != nil {
		return err
	}

	g.Line, err = AcceptChild(v, g.Line)
	if err != nil {
		return err
	}

	return v.Visit(g, PostVisit)
}

// AcceptChild calls node.Accept(v) and processes any returned NodeReplacements
// The returned node is either the input node or an appropriate replacement
func AcceptChild(v Visitor, node Node) (Node, error) {
	replaced := node
	err := replaced.Accept(v)
	repl, is := err.(NodeReplacement)
	for is {
		err = nil
		replaced = repl.Replacement[0]
		if repl.Skip {
			err = nil
			break
		}
		err = replaced.Accept(v)
		repl, is = err.(NodeReplacement)

	}
	return replaced, err
}

// AcceptChildLines calles Accept for ever element of old and handles node-replacements
func AcceptChildLines(parent Node, v Visitor, old []*Line) ([]*Line, error) {
	for i := 0; i < len(old); i++ {
		err := v.Visit(parent, i)
		if err != nil {
			return nil, err
		}
		err = old[i].Accept(v)
		repl, is := err.(NodeReplacement)
		if is {
			new := make([]*Line, 0, len(old)+len(repl.Replacement)-1)
			new = append(new, old[:i]...)
			for _, el := range repl.Replacement {
				new = append(new, el.(*Line))
			}
			new = append(new, old[i+1:]...)
			old = new
			err = nil
			if repl.Skip {
				i += len(repl.Replacement) - 1
			} else {
				i--
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return old, nil
}

// AcceptChildStatements calles Accept for ever element of old and handles node-replacements
func AcceptChildStatements(parent Node, v Visitor, old []Statement) ([]Statement, error) {
	for i := 0; i < len(old); i++ {
		err := v.Visit(parent, i)
		if err != nil {
			return nil, err
		}
		err = old[i].Accept(v)
		repl, is := err.(NodeReplacement)
		if is {
			new := make([]Statement, 0, len(old)+len(repl.Replacement)-1)
			new = append(new, old[:i]...)
			for _, el := range repl.Replacement {
				new = append(new, el.(Statement))
			}
			new = append(new, old[i+1:]...)
			old = new
			err = nil
			if repl.Skip {
				i += len(repl.Replacement) - 1
			} else {
				i--
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return old, nil
}
