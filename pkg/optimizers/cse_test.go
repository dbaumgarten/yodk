package optimizers

import (
	"testing"
)

var cseCases = map[string]string{
	`:foo=a+b+c :bar=a+b+d`:             `_tmp0=a+b :foo=_tmp0+c :bar=_tmp0+d`,
	`:out+="abc"+"\n" :out+="def"+"\n"`: `_tmp0="\n" :out+="abc"+_tmp0 :out+="def"+_tmp0`,
	`x=:foo/1000*1000 y=:bar/1000*1000`: `_tmp0=1000 x=:foo/_tmp0*_tmp0 y=:bar/_tmp0*_tmp0`,
	"a=1+x+2 b=1+x+2":                   "a=1+x+2 b=a",
	`x="abcd" y="abcd"`:                 `x="abcd" y=x`,
	`x=2*a a=3 y=2*a`:                   `x=2*a a=3 y=2*a`,
	`x=i++ y=i++ z=i++`:                 `x=i++ y=i++ z=i++`,
	`:out=:longglobal*:longglobal`:      `_tmp0=:longglobal :out=_tmp0*_tmp0`,
}

func TestCse(t *testing.T) {
	optimizationTesting(t, &CommonSubexpressionOptimizer{}, cseCases)
}
