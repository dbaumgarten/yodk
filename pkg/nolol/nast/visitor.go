package nast

import (
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// Accept is used to implement Acceptor
func (g *GoToLabelStatement) Accept(v ast.Visitor) error {
	return v.Visit(g, ast.SingleVisit)
}

// Accept is used to implement Acceptor
func (p *Program) Accept(v ast.Visitor) error {
	err := v.Visit(p, ast.PreVisit)
	if err != nil {
		return err
	}

	p.Elements, err = AcceptElementList(p, v, p.Elements)
	if err != nil {
		return err
	}

	return v.Visit(p, ast.PostVisit)
}

// Accept is used to implement Acceptor
func (l *StatementLine) Accept(v ast.Visitor) error {
	err := v.Visit(l, ast.PreVisit)
	if err != nil {
		return err
	}

	l.Statements, err = ast.AcceptChildStatements(l, v, l.Statements)
	if err != nil {
		return err
	}

	return v.Visit(l, ast.PostVisit)
}

// Accept is used to implement Acceptor
func (l *ConstDeclaration) Accept(v ast.Visitor) error {
	err := v.Visit(l, ast.PreVisit)
	if err != nil {
		return err
	}
	l.Value, err = ast.AcceptChild(v, l.Value)
	if err != nil {
		return err
	}
	return v.Visit(l, ast.PostVisit)
}

// Accept is used to implement Acceptor
func (s *Block) Accept(v ast.Visitor) error {
	err := v.Visit(s, ast.PreVisit)
	if err != nil {
		return err
	}

	s.Elements, err = AcceptNestableElementList(s, v, s.Elements)
	if err != nil {
		return err
	}

	err = v.Visit(s, ast.PostVisit)
	if err != nil {
		return err
	}
	return nil
}

// Accept is used to implement Acceptor
func (s *MultilineIf) Accept(v ast.Visitor) error {
	err := v.Visit(s, ast.PreVisit)
	if err != nil {
		return err
	}

	for i := 0; i < len(s.Conditions); i++ {
		err = v.Visit(s, i)
		if err != nil {
			return err
		}
		s.Conditions[i], err = ast.AcceptChild(v, s.Conditions[i])
		if err != nil {
			return err
		}
		err = v.Visit(s, ast.InterVisit1)
		if err != nil {
			return err
		}
		repl, err := ast.AcceptChild(v, s.Blocks[i])
		s.Blocks[i] = repl.(*Block)
		if err != nil {
			return err
		}
	}
	if s.ElseBlock != nil {
		err = v.Visit(s, ast.InterVisit2)
		if err != nil {
			return err
		}
		repl, err := ast.AcceptChild(v, s.ElseBlock)
		s.ElseBlock = repl.(*Block)
		if err != nil {
			return err
		}
	}
	return v.Visit(s, ast.PostVisit)
}

// Accept is used to implement Acceptor
func (s *WhileLoop) Accept(v ast.Visitor) error {
	err := v.Visit(s, ast.PreVisit)
	if err != nil {
		return err
	}
	s.Condition, err = ast.AcceptChild(v, s.Condition)
	if err != nil {
		return err
	}
	err = v.Visit(s, ast.InterVisit1)
	if err != nil {
		return err
	}
	repl, err := ast.AcceptChild(v, s.Block)
	s.Block = repl.(*Block)
	if err != nil {
		return err
	}
	return v.Visit(s, ast.PostVisit)
}

// Accept is used to implement Acceptor
func (s *WaitDirective) Accept(v ast.Visitor) error {
	err := v.Visit(s, ast.PreVisit)
	if err != nil {
		return err
	}
	s.Condition, err = ast.AcceptChild(v, s.Condition)
	if err != nil {
		return err
	}
	return v.Visit(s, ast.PostVisit)
}

// Accept is used to implement Acceptor
func (s *IncludeDirective) Accept(v ast.Visitor) error {
	return v.Visit(s, ast.SingleVisit)
}

// Accept is used to implement Acceptor
func (s *MacroDefinition) Accept(v ast.Visitor) error {
	err := v.Visit(s, ast.PreVisit)
	if err != nil {
		return err
	}

	rv, err := ast.AcceptChild(v, s.Block)
	if err != nil {
		return err
	}
	s.Block = rv.(*Block)

	return v.Visit(s, ast.PostVisit)
}

// Accept is used to implement Acceptor
func (s *MacroInsetion) Accept(v ast.Visitor) error {
	err := v.Visit(s, ast.PreVisit)
	if err != nil {
		return err
	}

	s.Arguments, err = AcceptExpressionList(s, v, s.Arguments)
	if err != nil {
		return err
	}

	return v.Visit(s, ast.PostVisit)
}

// AcceptElementList calles Accept for ever element of old and handles node-replacements
func AcceptElementList(parent ast.Node, v ast.Visitor, old []Element) ([]Element, error) {
	for i := 0; i < len(old); i++ {
		err := v.Visit(parent, i)
		if err != nil {
			return nil, err
		}
		err = old[i].Accept(v)
		repl, is := err.(ast.NodeReplacement)
		if is {
			new := make([]Element, 0, len(old)+len(repl.Replacement)-1)
			new = append(new, old[:i]...)
			for _, el := range repl.Replacement {
				new = append(new, el.(Element))
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

// AcceptNestableElementList calles Accept for every element of old and handles node-replacements
func AcceptNestableElementList(parent ast.Node, v ast.Visitor, old []NestableElement) ([]NestableElement, error) {
	for i := 0; i < len(old); i++ {
		err := v.Visit(parent, i)
		if err != nil {
			return nil, err
		}
		err = old[i].Accept(v)
		repl, is := err.(ast.NodeReplacement)
		if is {
			new := make([]NestableElement, 0, len(old)+len(repl.Replacement)-1)
			new = append(new, old[:i]...)
			for _, el := range repl.Replacement {
				new = append(new, el.(NestableElement))
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

// AcceptExpressionList calls Accept for every element of old and handles node-replacements
func AcceptExpressionList(parent ast.Node, v ast.Visitor, old []ast.Expression) ([]ast.Expression, error) {
	for i := 0; i < len(old); i++ {
		err := v.Visit(parent, i)
		if err != nil {
			return nil, err
		}
		err = old[i].Accept(v)
		repl, is := err.(ast.NodeReplacement)
		if is {
			new := make([]ast.Expression, 0, len(old)+len(repl.Replacement)-1)
			new = append(new, old[:i]...)
			for _, el := range repl.Replacement {
				new = append(new, el.(NestableElement))
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
