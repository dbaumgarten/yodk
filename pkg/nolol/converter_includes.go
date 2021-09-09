package nolol

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/parser/ast"
	"github.com/dbaumgarten/yodk/pkg/validators"
	"github.com/dbaumgarten/yodk/stdlib"
)

// resolveIncludes searches for include-directives and inserts the lines of the included files
func (c *Converter) convertInclude(include *nast.IncludeDirective) error {

	c.includecount++
	if c.includecount > 20 {
		return &parser.Error{
			Message:       "Error when processing includes: Include-loop detected",
			StartPosition: ast.NewPosition("", 1, 1),
			EndPosition:   ast.NewPosition("", 20, 70),
		}
	}

	filesnames := make([]string, 1)
	filesnames[0] = include.File

	file, err := c.getIncludedFile(include)
	if err != nil {
		return err
	}

	p := NewParser().(*Parser)
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

func (c *Converter) getIncludedFile(include *nast.IncludeDirective) (string, error) {

	importname := include.File
	getfunc := c.files.Get

	if stdlib.Is(importname) {
		getfunc = stdlib.Get
	} else {
		// this include is inside an included file
		if include.Position.File != "" {
			// the included file is inside another directory
			dir := path.Dir(include.Position.File)
			if dir != "." {
				dir = filepath.ToSlash(dir)
				// fix the import-path
				importname = path.Join(dir, importname)
			}
		}
	}

	filename := importname

	if !strings.HasSuffix(importname, ".nolol") {
		filename = importname + ".nolol"
	}

	// first try to import exact file
	file, origerr := getfunc(filename)
	if origerr == nil {
		return file, nil
	}

	// next try all available chip-specific imports
	switch c.targetChipType {
	case validators.ChipTypeProfessional:
		file, err := getfunc(importname + "_" + validators.ChipTypeProfessional + ".nolol")
		if err == nil {
			return file, nil
		}
		fallthrough
	case validators.ChipTypeAdvanced:
		file, err := getfunc(importname + "_" + validators.ChipTypeAdvanced + ".nolol")
		if err == nil {
			return file, nil
		}
		fallthrough
	case validators.ChipTypeBasic:
		file, err := getfunc(importname + "_" + validators.ChipTypeBasic + ".nolol")
		if err == nil {
			return file, nil
		}
	}

	return "", &parser.Error{
		Message:       fmt.Sprintf("Error when opening included file '%s': %s", importname, origerr.Error()),
		StartPosition: include.Start(),
		EndPosition:   include.End(),
	}
}
