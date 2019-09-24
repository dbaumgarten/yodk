package parser

const (
	PreVisit    = -1
	PostVisit   = -2
	SingleVisit = -3
	InterVisit1 = -4
	InterVisit2 = -5
)

type Visitor interface {
	Visit(node Node, visitType int) error
}

type VisitorFunc func(node Node, visitType int) error

func (f VisitorFunc) Visit(node Node, visitType int) error {
	return f(node, visitType)
}

type NodeReplacement struct {
	Replacement []Node
}

func NewNodeReplacement(replacement ...Node) NodeReplacement {
	return NodeReplacement{
		Replacement: replacement,
	}
}

func (e NodeReplacement) Error() string {
	return "SHOULD NEVER HAPPEN"
}

func (p *Programm) Accept(v Visitor) error {
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

func (c *StringConstant) Accept(v Visitor) error {
	return v.Visit(c, SingleVisit)
}

func (c *NumberConstant) Accept(v Visitor) error {
	return v.Visit(c, SingleVisit)
}

func (d *Dereference) Accept(v Visitor) error {
	return v.Visit(d, SingleVisit)
}

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

func (f *FuncCall) Accept(v Visitor) error {
	err := v.Visit(f, PreVisit)
	if err != nil {
		return err
	}
	err = f.Argument.Accept(v)
	if repl, is := err.(NodeReplacement); is {
		f.Argument = repl.Replacement[0].(Expression)
		err = nil
	}
	if err != nil {
		return err
	}
	return v.Visit(f, PostVisit)
}

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

func (g *GoToStatement) Accept(v Visitor) error {
	return v.Visit(g, SingleVisit)
}

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
