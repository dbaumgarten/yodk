package validators

import (
	"strings"

	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

// ValidateCodeLength checks if the given code (in text-format) does fit yolol's boundaries
func ValidateCodeLength(code string) error {
	lines := strings.Split(code, "\n")
	if len(lines) > 20 {
		return &parser.Error{
			Message: "Program has more than 20 lines",
			StartPosition: ast.Position{
				Line:    21,
				Coloumn: 1,
			},
			EndPosition: ast.Position{
				Line:    len(lines),
				Coloumn: 70,
			},
		}
	}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 70 {
			return &parser.Error{
				Message: "The line has more than 70 characters",
				StartPosition: ast.Position{
					Line:    i + 1,
					Coloumn: 70,
				},
				EndPosition: ast.Position{
					Line:    i + 1,
					Coloumn: len(line),
				},
			}
		}
	}
	return nil
}
