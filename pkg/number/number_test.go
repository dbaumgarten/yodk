package number

import (
	"math"
	"reflect"
	"testing"
)

func TestFromString(t *testing.T) {

	tests := []struct {
		teststr string
		want    Number
		wanterr bool
	}{
		{
			teststr: "5",
			want:    FromInt(5),
		},
		{
			teststr: "123",
			want:    FromInt(123),
		},
		{
			teststr: "1.5",
			want:    FromFloat64(1.5),
		},
		{
			teststr: "1.567",
			want:    FromFloat64(1.567),
		},
		{
			teststr: "1.5678",
			want:    FromFloat64(1.567),
		},
		{
			teststr: "1.56.7",
			wanterr: true,
		},
		{
			teststr: ".17",
			want:    FromFloat64(0.17),
		},
		{
			teststr: "-5",
			want:    FromInt(-5),
		},
		{
			teststr: "-5.5",
			want:    FromFloat64(-5.5),
		},
	}
	for _, tt := range tests {
		got, err := FromString(tt.teststr)
		if err != nil && !tt.wanterr {
			t.Errorf("FromString() error = %v", err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FromString() = %v, want %v", got, tt.want)
		}
	}
}

func TestFromInt(t *testing.T) {
	type args struct {
		in int
	}
	tests := []struct {
		name string
		args args
		want Number
	}{
		{
			"test1",
			args{
				5,
			},
			MustFromString("5"),
		},
		{
			"test2",
			args{
				-5,
			},
			MustFromString("-5"),
		},
		{
			"test4",
			args{
				12345,
			},
			MustFromString("12345"),
		},
	}
	for _, tt := range tests {
		if got := FromInt(tt.args.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. FromInt() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestFromFloat64(t *testing.T) {
	type args struct {
		in float64
	}
	tests := []struct {
		name string
		args args
		want Number
	}{
		{
			"test1",
			args{
				0.0,
			},
			MustFromString("0"),
		},
		{
			"test2",
			args{
				7.0,
			},
			MustFromString("7"),
		},
		{
			"test3",
			args{
				123.789,
			},
			MustFromString("123.789"),
		},
		{
			"test4",
			args{
				-123.789,
			},
			MustFromString("-123.789"),
		},
	}
	for _, tt := range tests {
		if got := FromFloat64(tt.args.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. FromFloat64() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNumber_Float64(t *testing.T) {
	tests := []struct {
		name string
		n    Number
		want float64
	}{
		{
			"test1",
			MustFromString("123.789"),
			123.789,
		},
		{
			"test2",
			MustFromString("123"),
			123.000,
		},
		{
			"test4",
			MustFromString("-123.789"),
			-123.789,
		},
		{
			"test5",
			MustFromString("0"),
			0,
		},
	}
	for _, tt := range tests {
		if got := tt.n.Float64(); got != tt.want {
			t.Errorf("%q. Number.Float64() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNumber_String(t *testing.T) {
	tests := []struct {
		name string
		n    Number
		want string
	}{
		{
			"test1",
			MustFromString("123"),
			"123",
		},
		{
			"test2",
			MustFromString("123.789"),
			"123.789",
		},
		{
			"test3",
			MustFromString("123.1"),
			"123.100",
		},
		{
			"test4",
			MustFromString("-123"),
			"-123",
		},
		{
			"test5",
			MustFromString("-123.789"),
			"-123.789",
		},
	}
	for _, tt := range tests {
		if got := tt.n.String(); got != tt.want {
			t.Errorf("%q. Number.String() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNumber_Int(t *testing.T) {
	tests := []struct {
		name string
		n    Number
		want int
	}{
		{
			"test1",
			MustFromString("123"),
			123,
		},
		{
			"test2",
			MustFromString("123.12"),
			123,
		},
		{
			"test3",
			MustFromString("-123"),
			-123,
		},
		{
			"test4",
			MustFromString("-123.12"),
			-123,
		},
	}
	for _, tt := range tests {
		if got := tt.n.Int(); got != tt.want {
			t.Errorf("%q. Number.Int() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNumber_Add(t *testing.T) {
	type args struct {
		m Number
	}
	tests := []struct {
		name string
		n    Number
		args args
		want Number
	}{
		{
			"test1",
			MustFromString("123"),
			args{
				MustFromString("123"),
			},
			MustFromString("246"),
		},
		{
			"test2",
			MustFromString("123"),
			args{
				MustFromString("-123"),
			},
			MustFromString("0"),
		},
		{
			"test3",
			MustFromString("123.100"),
			args{
				MustFromString("123.025"),
			},
			MustFromString("246.125"),
		},
	}
	for _, tt := range tests {
		if got := tt.n.Add(tt.args.m); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Number.Add() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNumber_Sub(t *testing.T) {
	type args struct {
		m Number
	}
	tests := []struct {
		name string
		n    Number
		args args
		want Number
	}{
		{
			"test1",
			MustFromString("123"),
			args{
				MustFromString("123"),
			},
			MustFromString("0"),
		},
		{
			"test2",
			MustFromString("123"),
			args{
				MustFromString("100"),
			},
			MustFromString("23"),
		},
		{
			"test3",
			MustFromString("50"),
			args{
				MustFromString("100.50"),
			},
			MustFromString("-50.5"),
		},
	}
	for _, tt := range tests {
		if got := tt.n.Sub(tt.args.m); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Number.Sub() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNumber_Mul(t *testing.T) {
	type args struct {
		m Number
	}
	tests := []struct {
		name string
		n    Number
		args args
		want Number
	}{
		{
			"test1",
			MustFromString("123"),
			args{
				MustFromString("1"),
			},
			MustFromString("123"),
		},
		{
			"test2",
			MustFromString("123"),
			args{
				MustFromString("2"),
			},
			MustFromString("246"),
		},
		{
			"test3",
			MustFromString("123.12"),
			args{
				MustFromString("-2"),
			},
			MustFromString("-246.24"),
		},
		{
			"test4",
			MustFromString("124"),
			args{
				MustFromString("0.5"),
			},
			MustFromString("62"),
		},
		{
			"test4",
			MustFromString("1"),
			args{
				MustFromString("0.333"),
			},
			MustFromString("0.333"),
		},
	}
	for _, tt := range tests {
		if got := tt.n.Mul(tt.args.m); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Number.Mul() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNumber_Div(t *testing.T) {
	type args struct {
		m Number
	}
	tests := []struct {
		name    string
		n       Number
		args    args
		want    Number
		wantErr bool
	}{
		{
			"test1",
			MustFromString("123"),
			args{
				MustFromString("1"),
			},
			MustFromString("123"),
			false,
		},
		{
			"test2",
			MustFromString("124.5"),
			args{
				MustFromString("2"),
			},
			MustFromString("62.25"),
			false,
		},
		{
			"test3",
			MustFromString("1"),
			args{
				MustFromString("3"),
			},
			MustFromString("0.333"),
			false,
		},
		{
			"test3",
			MustFromString("1"),
			args{
				MustFromString("-3"),
			},
			MustFromString("-0.333"),
			false,
		},
		{
			"test4",
			MustFromString("1"),
			args{
				MustFromString("0"),
			},
			MustFromString("0"),
			true,
		},
	}
	for _, tt := range tests {
		got, err := tt.n.Div(tt.args.m)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Number.Div() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Number.Div() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNumber_Abs(t *testing.T) {
	tests := []struct {
		name string
		n    Number
		want Number
	}{
		{
			"test1",
			MustFromString("159"),
			MustFromString("159"),
		},
		{
			"test2",
			MustFromString("-159"),
			MustFromString("159"),
		},
		{
			"test3",
			MustFromString("0"),
			MustFromString("0"),
		},
	}
	for _, tt := range tests {
		if got := tt.n.Abs(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Number.Abs() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNumber_Floor(t *testing.T) {
	a := MustFromString("123.1")
	d, _ := a.Div(FromInt(1000))
	if d.Mul(FromInt(1000)).String() != "123" {
		t.Fatal("Flooring failed")
	}
}

func TestNumber_Sin(t *testing.T) {
	cases := []float64{3.141, 0, 1, -1, -3.141}
	for _, c := range cases {
		if math.Abs(FromFloat64(c).Sin().Float64()-math.Sin(c*math.Pi/180)) > 0.001 {
			t.Fatal("Sin is wrong")
		}
	}
}

func TestNumber_Cos(t *testing.T) {
	cases := []float64{3.141, 0, 1, -1, -3.141}
	for _, c := range cases {
		if math.Abs(FromFloat64(c).Cos().Float64()-math.Cos(c*math.Pi/180)) > 0.001 {
			t.Fatal("Cos is wrong")
		}
	}
}

func TestNumber_Tan(t *testing.T) {
	cases := []float64{3.141, 0, 1, -1, -3.141}
	for _, c := range cases {
		if math.Abs(FromFloat64(c).Tan().Float64()-math.Tan(c*math.Pi/180)) > 0.001 {
			t.Fatal("Tan is wrong")
		}
	}
}

func TestNumber_Asin(t *testing.T) {
	cases := []float64{3.141, 0, 1, -1, -3.141}
	for _, c := range cases {
		if math.Abs(FromFloat64(c).Asin().Float64()-math.Asin(c)*180/math.Pi) > 0.001 {
			t.Fatal("Asin is wrong")
		}
	}
}

func TestNumber_Acos(t *testing.T) {
	cases := []float64{3.141, 0, 1, -1, -3.141}
	for _, c := range cases {
		if math.Abs(FromFloat64(c).Acos().Float64()-math.Acos(c)*180/math.Pi) > 0.001 {
			t.Fatal("Acos is wrong")
		}
	}
}

func TestNumber_Atan(t *testing.T) {
	cases := []float64{3.141, 0, 1, -1, -3.141}
	for _, c := range cases {
		if math.Abs(FromFloat64(c).Atan().Float64()-math.Atan(c)*180/math.Pi) > 0.001 {
			t.Fatal("Atan is wrong")
		}
	}
}
