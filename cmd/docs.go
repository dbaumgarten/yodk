package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/nolol/nast"
	"github.com/dbaumgarten/yodk/pkg/validators"

	"github.com/spf13/cobra"
)

var name string
var regex string
var docsOutputFile string

// docsCmd represents the compile command
var docsCmd = &cobra.Command{
	Use:   "docs [file]+",
	Short: "Generate markdown-documentation for nolol-files",
	Run: func(cmd *cobra.Command, args []string) {
		if docsOutputFile != "" {
			os.Remove(docsOutputFile)
		}
		for _, file := range args {
			generateDocstring(file)
		}
	},
	Args: cobra.MinimumNArgs(1),
}

func getMacroSignature(m *nast.MacroDefinition) string {
	text := m.Name + "(" + strings.Join(m.Arguments, ", ") + ")"
	if len(m.Externals) > 0 {
		text += "<" + strings.Join(m.Externals, ", ") + ">"
	}
	text += " " + m.Type
	return text
}

func mdNewlines(in string) string {
	return strings.ReplaceAll(in, "\n", "  \n")
}

func mdCodeBlock(in string) string {
	return "```\n" + in + "\n```"
}

var defaultTemplate = template.Must(template.New("templ").Funcs(template.FuncMap{
	"macroSignature": getMacroSignature,
	"mdNewlines":     mdNewlines,
	"mdCodeBlock":    mdCodeBlock,
}).Parse(`{{ $root := . }}# {{ .Name }}

{{ mdNewlines .Report.FileDocstring }}
{{ if .Report.Definitions }}
## Definitions
{{ range $name, $element := .Report.Definitions }} 
### **{{ $name }}**
{{ mdNewlines (index $root.Report.Docstrings $name) }}
{{ end }}{{ end }}
{{ if .Report.Macros }}
## Macros
{{ range $name, $element := .Report.Macros }} 
### **{{ $name }}**
{{ mdCodeBlock (macroSignature $element) }}
{{ mdNewlines (index $root.Report.Docstrings $name) }}

{{ end }}
{{ end }}

`))

func generateDocstring(fpath string) {
	if !strings.HasSuffix(fpath, ".nolol") {
		fmt.Println("This command only works for .nolol files")
		os.Exit(1)
	}

	if chipType != validators.ChipTypeAuto {
		autotype, err := validators.AutoChooseChipType(validators.ChipTypeAuto, fpath)
		if err == nil && autotype != chipType {
			// The user specified a chip-type, but the auto-detected type of the current file does not match it
			// ignore the file
			return
		}
	}

	chip, err := validators.AutoChooseChipType(chipType, fpath)
	exitOnError(err, "choosing chip-type")

	converter := nolol.NewConverter()
	converter.SetChipType(chip)
	converteri := converter.LoadFile(fpath).ProcessIncludes()
	exitOnError(converteri.Error(), "parsing file")

	report, err := nolol.Analyse(converteri.GetIntermediateProgram())
	exitOnError(err, "analyzing file")

	var submatches []string
	if regex != "" {
		rg, err := regexp.Compile(regex)
		exitOnError(err, "compiling regex")
		submatches = rg.FindStringSubmatch(fpath)
	}

	filename := fpath

	if name != "" {
		filename = name
		for i, sm := range submatches {
			filename = strings.ReplaceAll(filename, "$"+strconv.Itoa(i), sm)
		}
	}

	var out io.Writer
	out = os.Stdout
	if docsOutputFile != "" {
		outf, err := os.OpenFile(docsOutputFile, os.O_CREATE|os.O_APPEND, 700)
		exitOnError(err, "opening output-file")
		out = outf
	}

	defaultTemplate.Execute(out, struct {
		Name   string
		Report *nolol.AnalysisReport
	}{
		Name:   filename,
		Report: report,
	})
}

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringVarP(&docsOutputFile, "out", "o", "", "The output file. Defaults to stdout.")
	docsCmd.Flags().StringVarP(&name, "name", "n", "", "Use this as filename in generated doc")
	docsCmd.Flags().StringVarP(&regex, "regex", "r", "", "A regex to extract submatches from the input-filename to be used in -n")
	docsCmd.Flags().StringVarP(&chipType, "chip", "c", "auto", "Chip-type to validate for. (auto|professional|advanced|basic)")
}
