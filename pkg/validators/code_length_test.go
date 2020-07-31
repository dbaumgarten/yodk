package validators

import (
	"testing"

	"github.com/dbaumgarten/yodk/pkg/parser"
)

var code1 = `if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
`

var code2 = `if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
if b==3 then c=3 end goto 1 a="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if 1!=b then goto 3 end if b==1 then c=1 end if b==2 then c=2 end
`

func TestValidateCodeLength(t *testing.T) {
	err := ValidateCodeLength(code1)
	if err == nil {
		t.Fatal("Did not find overlong line")
	}

	if err.(*parser.Error).StartPosition.Line != 4 {
		t.Fatal("Wrong line for overlong line error")
	}

	err = ValidateCodeLength(code2)
	if err == nil {
		t.Fatal("Did not find too many lines")
	}

	if err.(*parser.Error).StartPosition.Line != 21 {
		t.Fatal("Wrong line for overlong program error")
	}
}
