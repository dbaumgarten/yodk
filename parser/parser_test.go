package parser

import (
	"fmt"
	"testing"

	"github.com/dbaumgarten/yodk/generators"
)

func TestParser(t *testing.T) {
	programm := `abc = 123
	def = "hallo"
	a = (-b + c * d) == 1
	if a == 1 then b=2 else goto 1 end
	if a==b and a+b==2 then c=x end
	a = b
	if a==b then if b==c then goto 2 end end
	`

	p := NewParser()

	result, err := p.Parse(programm)

	if err != nil {
		t.Fatal(err)
	}

	printer := &generators.YololGenerator{}

	fmt.Println(printer.Generate(result))
}
