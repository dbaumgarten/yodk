package nolol

import (
	"github.com/dbaumgarten/yodk/parser"
)

func (g *GoToLabelStatement) Accept(v parser.Visitor) error {
	return v.Visit(g, parser.SingleVisit)
}

func (p *NololProgramm) Accept(v parser.Visitor) error {
	err := v.Visit(p, parser.PreVisit)
	if err != nil {
		return err
	}
	for i := 0; i < len(p.Lines); i++ {
		err = v.Visit(p, i)
		if err != nil {
			return err
		}
		err = p.Lines[i].Accept(v)
		if repl, is := err.(parser.NodeReplacement); is {
			p.Lines = patchLines(p.Lines, i, repl)
			i += len(repl.Replacement) - 1
			err = nil
		}
		if err != nil {
			return err
		}
	}
	return v.Visit(p, parser.PostVisit)
}

func (l *StatementLine) Accept(v parser.Visitor) error {
	err := v.Visit(l, parser.PreVisit)
	if err != nil {
		return err
	}
	for i := 0; i < len(l.Statements); i++ {
		err = v.Visit(l, i)
		if err != nil {
			return err
		}
		err = l.Statements[i].Accept(v)
		if repl, is := err.(parser.NodeReplacement); is {
			l.Statements = parser.PatchStatements(l.Statements, i, repl)
			i += len(repl.Replacement) - 1
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
		l.Value = repl.Replacement[0].(parser.Expression)
		err = nil
	}
	if err != nil {
		return err
	}
	return v.Visit(l, parser.PostVisit)
}

func (s *MultilineIf) Accept(v parser.Visitor) error {
	err := v.Visit(s, parser.PreVisit)
	if err != nil {
		return err
	}
	err = s.Condition.Accept(v)
	if repl, is := err.(parser.NodeReplacement); is {
		s.Condition = repl.Replacement[0].(parser.Expression)
		err = nil
	}
	if err != nil {
		return err
	}
	err = v.Visit(s, parser.InterVisit1)
	if err != nil {
		return err
	}
	for i := 0; i < len(s.IfBlock); i++ {
		err = v.Visit(s, i)
		if err != nil {
			return err
		}
		err = s.IfBlock[i].Accept(v)
		if repl, is := err.(parser.NodeReplacement); is {
			s.IfBlock = patchExecutableLines(s.IfBlock, i, repl)
			i += len(repl.Replacement) - 1
			err = nil
		}
		if err != nil {
			return err
		}
	}
	if s.ElseBlock != nil {
		err = v.Visit(s, parser.InterVisit2)
		if err != nil {
			return err
		}
		for i := 0; i < len(s.ElseBlock); i++ {
			err = v.Visit(s, i)
			if err != nil {
				return err
			}
			err = s.ElseBlock[i].Accept(v)
			if repl, is := err.(parser.NodeReplacement); is {
				s.ElseBlock = patchExecutableLines(s.ElseBlock, i, repl)
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
	return v.Visit(s, parser.PostVisit)
}

func patchLines(old []NololLine, position int, repl parser.NodeReplacement) []NololLine {
	newv := make([]NololLine, 0, len(old)+len(repl.Replacement)-1)
	newv = append(newv, old[:position]...)
	for _, elem := range repl.Replacement {
		if line, is := elem.(NololLine); is {
			newv = append(newv, line)
		} else {
			panic("Could not patch slice. Wrong type.")
		}
	}
	newv = append(newv, old[position+1:]...)
	return newv
}

func patchExecutableLines(old []ExecutableLine, position int, repl parser.NodeReplacement) []ExecutableLine {
	newv := make([]ExecutableLine, 0, len(old)+len(repl.Replacement)-1)
	newv = append(newv, old[:position]...)
	for _, elem := range repl.Replacement {
		if line, is := elem.(ExecutableLine); is {
			newv = append(newv, line)
		} else {
			panic("Could not patch slice. Wrong type.")
		}
	}
	newv = append(newv, old[position+1:]...)
	return newv
}
