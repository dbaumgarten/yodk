package nolol

import (
	"fmt"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// resolveIncludes searches for include-directives and inserts the lines of the included files
func (c *Converter) convertInclude(include *nast.IncludeDirective) error {
	p := NewParser()

	c.includecount++
	if c.includecount > 20 {
		return &parser.Error{
			Message:       "Error when processing includes: Include-loop detected",
			StartPosition: ast.NewPosition("", 1, 1),
			EndPosition:   ast.NewPosition("", 20, 70),
		}
	}

	file, err := c.files.Get(include.File)
	if err != nil {
		return &parser.Error{
			Message:       fmt.Sprintf("Error when opening included file '%s': %s", include.File, err.Error()),
			StartPosition: include.Start(),
			EndPosition:   include.End(),
		}
	}
	p.SetFilename(include.File)
	parsed, err := p.Parse(file)
	if err != nil {
		// override the position of the error with the position of the include
		// this way the error gets displayed at the correct location
		// the message does contain the original location
		return &parser.Error{
			Message:       err.Error(),
			StartPosition: include.Start(),
			EndPosition:   include.End(),
		}
	}

	if usesTimeTracking(parsed) {
		c.usesTimeTracking = true
	}

	replacements := make([]ast.Node, len(parsed.Elements))
	for i := range parsed.Elements {
		replacements[i] = parsed.Elements[i]
	}
	return ast.NewNodeReplacement(replacements...)
}
