package parser

func (g *GoToLabelStatement) Accept(v Visitor) error {
	return v.Visit(g, SingleVisit)
}

func (p *ExtProgramm) Accept(v Visitor) error {
	err := v.Visit(p, PreVisit)
	if err != nil {
		return err
	}
	for i, line := range p.ExecutableLines {
		err = v.Visit(p, i)
		if err != nil {
			return err
		}
		err = line.Accept(v)
		if repl, is := err.(NodeReplacement); is {
			p.ExecutableLines[i] = repl.Replacement.(*ExtLine)
			err = nil
		}
		if err != nil {
			return err
		}
	}
	return v.Visit(p, PostVisit)
}

func (l *ExtLine) Accept(v Visitor) error {
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
		if repl, is := err.(NodeReplacement); is {
			l.Statements[i] = repl.Replacement.(Statement)
			err = nil
		}
		if err != nil {
			return err
		}
	}
	return v.Visit(l, PostVisit)
}
