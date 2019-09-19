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

func (p *Programm) Accept(v Visitor) error {
	err := v.Visit(p, PreVisit)
	if err != nil {
		return err
	}
	for i, line := range p.Lines {
		err = v.Visit(p, i)
		if err != nil {
			return err
		}
		err = line.Accept(v)
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
	for i, stmt := range l.Statements {
		err = v.Visit(l, i)
		if err != nil {
			return err
		}
		err = stmt.Accept(v)
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
	if err != nil {
		return err
	}
	err = v.Visit(o, InterVisit1)
	err = o.Exp2.Accept(v)
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
	if err != nil {
		return err
	}
	err = v.Visit(s, InterVisit1)
	if err != nil {
		return err
	}
	for i, ifstmt := range s.IfBlock {
		err = v.Visit(s, i)
		if err != nil {
			return err
		}
		err = ifstmt.Accept(v)
		if err != nil {
			return err
		}
	}
	if s.ElseBlock != nil {
		err = v.Visit(s, InterVisit2)
		if err != nil {
			return err
		}
		for i, elsestmt := range s.ElseBlock {
			err = v.Visit(s, i)
			if err != nil {
				return err
			}
			err = elsestmt.Accept(v)
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
