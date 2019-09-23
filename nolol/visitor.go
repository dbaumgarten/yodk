package nolol

import "github.com/dbaumgarten/yodk/parser"

func (g *GoToLabelStatement) Accept(v parser.Visitor) error {
	return v.Visit(g, parser.SingleVisit)
}

func (p *ExtProgramm) Accept(v parser.Visitor) error {
	err := v.Visit(p, parser.PreVisit)
	if err != nil {
		return err
	}
	for i, line := range p.ExecutableLines {
		err = v.Visit(p, i)
		if err != nil {
			return err
		}
		err = line.Accept(v)
		if repl, is := err.(parser.NodeReplacement); is {
			p.ExecutableLines[i] = repl.Replacement.(*ExecutableLine)
			err = nil
		}
		if err != nil {
			return err
		}
	}
	return v.Visit(p, parser.PostVisit)
}

func (l *ExecutableLine) Accept(v parser.Visitor) error {
	err := v.Visit(l, parser.PreVisit)
	if err != nil {
		return err
	}
	for i, stmt := range l.Statements {
		err = v.Visit(l, i)
		if err != nil {
			return err
		}
		err = stmt.Accept(v)
		if repl, is := err.(parser.NodeReplacement); is {
			l.Statements[i] = repl.Replacement.(parser.Statement)
			err = nil
		}
		if err != nil {
			return err
		}
	}
	return v.Visit(l, parser.PostVisit)
}

func (l *ConstDeclaration) Accept(v parser.Visitor) error {
	err := v.Visit(l, parser.PreVisit)
	if err != nil {
		return err
	}
	err = l.Value.Accept(v)
	if repl, is := err.(parser.NodeReplacement); is {
		l.Value = repl.Replacement.(parser.Expression)
		err = nil
	}
	if err != nil {
		return err
	}
	return v.Visit(l, parser.PostVisit)
}
