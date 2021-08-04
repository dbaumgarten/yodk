package validators_test

import (
	"testing"

	"github.com/dbaumgarten/yodk/pkg/nolol"
	"github.com/dbaumgarten/yodk/pkg/parser"
	"github.com/dbaumgarten/yodk/pkg/validators"
)

func TestAutoChoosingChiptype(t *testing.T) {
	testdata := []struct {
		choosen   string
		filename  string
		wanted    string
		expectErr bool
	}{
		{
			choosen:  "auto",
			filename: "myscript_basic.yolol",
			wanted:   validators.ChipTypeBasic,
		},
		{
			choosen:  "auto",
			filename: "C:\\test\\myscript_basic.yolol",
			wanted:   validators.ChipTypeBasic,
		},
		{
			choosen:  "auto",
			filename: "myscript_advanced.yolol",
			wanted:   validators.ChipTypeAdvanced,
		},
		{
			choosen:  "auto",
			filename: "/foo/myscript_professional.yolol",
			wanted:   validators.ChipTypeProfessional,
		},
		{
			choosen:  "basic",
			filename: "myscript_basic.yolol",
			wanted:   validators.ChipTypeBasic,
		},
		{
			choosen:  "basic",
			filename: "C:\\test\\myscript_basic.yolol",
			wanted:   validators.ChipTypeBasic,
		},
		{
			choosen:  "basic",
			filename: "myscript_advanced.yolol",
			wanted:   validators.ChipTypeBasic,
		},
		{
			choosen:  "basic",
			filename: "myscript_professional.yolol",
			wanted:   validators.ChipTypeBasic,
		},
		{
			choosen:  "auto",
			filename: "myscript.yolol",
			wanted:   validators.ChipTypeProfessional,
		},
		{
			choosen:   "foobar",
			filename:  "myscript.yolol",
			expectErr: true,
		},
	}

	for i, entry := range testdata {
		out, err := validators.AutoChooseChipType(entry.choosen, entry.filename)
		if err == nil && entry.expectErr {
			t.Fatalf("Expected an error but got none for case %d", i)
		}
		if err != nil && !entry.expectErr {
			t.Error(err)
		}
		if out != entry.wanted {
			t.Fatalf("Wrong output for %d: %s", i, out)
		}
	}
}

func TestAvailableOps(t *testing.T) {
	testdata := []struct {
		prog      string
		chiptype  string
		expectErr bool
	}{
		{
			prog:     "x=1+2",
			chiptype: validators.ChipTypeBasic,
		},
		{
			prog:     "x=1+2",
			chiptype: validators.ChipTypeAdvanced,
		},
		{
			prog:     "x=1+2",
			chiptype: validators.ChipTypeProfessional,
		},
		{
			prog:      "x=1%2",
			chiptype:  validators.ChipTypeBasic,
			expectErr: true,
		},
		{
			prog:     "x=1%2",
			chiptype: validators.ChipTypeAdvanced,
		},
		{
			prog:     "x=1%2",
			chiptype: validators.ChipTypeProfessional,
		},
		{
			prog:      "x=sin(13)",
			chiptype:  validators.ChipTypeBasic,
			expectErr: true,
		},
		{
			prog:      "x=sin(13)",
			chiptype:  validators.ChipTypeAdvanced,
			expectErr: true,
		},
		{
			prog:     "x=sin(13)",
			chiptype: validators.ChipTypeProfessional,
		},
		{
			prog:      "x=abs(13)",
			chiptype:  validators.ChipTypeBasic,
			expectErr: true,
		},
		{
			prog:     "x=abs(13)",
			chiptype: validators.ChipTypeAdvanced,
		},
	}
	for i, entry := range testdata {
		parsed, err := parser.NewParser().Parse(entry.prog)
		if err != nil {
			t.Error(err)
		}

		nololParsed, err := nolol.NewParser().Parse(entry.prog)
		if err != nil {
			t.Error(err)
		}

		err = validators.ValidateAvailableOperations(parsed, entry.chiptype)
		if err == nil && entry.expectErr {
			t.Fatalf("Expected error for test %d", i)
		}
		if err != nil && !entry.expectErr {
			t.Fatalf("Expected no error for test %d", i)
		}

		err = validators.ValidateAvailableOperations(nololParsed, entry.chiptype)
		if err == nil && entry.expectErr {
			t.Fatalf("Expected error for nolol-test %d", i)
		}
		if err != nil && !entry.expectErr {
			t.Fatalf("Expected no error for nolol-test %d", i)
		}

	}
}
