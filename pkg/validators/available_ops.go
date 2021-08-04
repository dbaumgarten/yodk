package validators

import (
	"fmt"
	"regexp"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
)

const (
	ChipTypeAuto         = "auto"
	ChipTypeBasic        = "basic"
	ChipTypeAdvanced     = "advanced"
	ChipTypeProfessional = "professional"
)

var unavailableBinaryOps = map[string][]string{
	ChipTypeBasic:        {"^", "%"},
	ChipTypeAdvanced:     {},
	ChipTypeProfessional: {},
}

var unavailableUnaryOps = map[string][]string{
	ChipTypeBasic:        {"!", "sqrt", "sin", "cos", "tan", "asin", "acos", "atan", "abs"},
	ChipTypeAdvanced:     {"sin", "cos", "tan", "asin", "acos", "atan"},
	ChipTypeProfessional: {},
}

var unavailableAssignments = map[string][]string{
	ChipTypeBasic:        {"^=", "%="},
	ChipTypeAdvanced:     {},
	ChipTypeProfessional: {},
}

func contains(list []string, element string) bool {
	for _, el := range list {
		if el == element {
			return true
		}
	}
	return false
}

var filenameChiptypeRegex = regexp.MustCompile(".*_(basic|advanced|professional).(?:n|y)olol")

// AutoChooseChipType chooses a chip-type based on the provided type and the filename of the source-file
func AutoChooseChipType(choosen string, filename string) (string, error) {

	if choosen != ChipTypeBasic && choosen != ChipTypeAdvanced && choosen != ChipTypeProfessional && choosen != ChipTypeAuto {
		return "", fmt.Errorf("Unknown chip-type. Possible options are professional, advanced, basic or auto")
	}

	if choosen != ChipTypeAuto {
		return choosen, nil
	}

	match := filenameChiptypeRegex.FindStringSubmatch(filename)
	if match != nil {
		return match[1], nil
	}

	return ChipTypeProfessional, nil
}

// ValidateAvailableOperations checks if all used operations are available on the given chip-type
func ValidateAvailableOperations(program ast.Node, chiptype string) error {

	if chiptype == ChipTypeProfessional {
		return nil
	}

	errors := make(parser.Errors, 0)

	logError := func(op string, node ast.Node) {
		errors = append(errors, &parser.Error{
			Message:       fmt.Sprintf("Operator '%s' is not available on %s-chips", op, chiptype),
			StartPosition: node.Start(),
			EndPosition:   node.End(),
		})
	}

	f := func(node ast.Node, visitType int) error {
		if visitType == ast.SingleVisit || visitType == ast.PreVisit {
			switch n := node.(type) {
			case *ast.UnaryOperation:
				if contains(unavailableUnaryOps[chiptype], n.Operator) {
					logError(n.Operator, n)
				}
			case *ast.BinaryOperation:
				if contains(unavailableBinaryOps[chiptype], n.Operator) {
					logError(n.Operator, n)
				}
			case *ast.Assignment:
				if contains(unavailableAssignments[chiptype], n.Operator) {
					logError(n.Operator, n)
				}
			case *nast.FuncCall:
				if contains(unavailableUnaryOps[chiptype], n.Function) {
					logError(n.Function, n)
				}
			}
		}
		return nil
	}
	err := program.Accept(ast.VisitorFunc(f))
	if err != nil {
		return err
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
