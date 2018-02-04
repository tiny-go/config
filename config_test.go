package config

import (
	"flag"
	"reflect"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// system flags should be removed first otherwise they will break the tests
var toBeRemoved = []string{
	"-convey",
	"-test.",
}

func init() {
	args = func(arguments []string) (filtered []string) {
		for _, argument := range arguments {
			for _, deprecated := range toBeRemoved {
				if strings.Contains(argument, deprecated) {
					goto SKIP
				}
			}
			filtered = append(filtered, argument)
		SKIP:
		}
		return
	}(args)
}

func Test_JoinStrings(t *testing.T) {
	type testCase struct {
		title string
		in    []string
		out   string
	}
	var cases = []testCase{
		{"not empty strings", []string{"-", "a", "b", "c"}, "a-b-c"},
		{"with empty string", []string{"_", "", "b", "c"}, "b_c"},
		{"all empty strings", []string{"#", "", "", ""}, ""},
	}
	Convey("Join Strings", t, func() {
		for _, c := range cases {
			Convey(c.title, func() {
				So(joinStrings(c.in[0], c.in[1:]...), ShouldEqual, c.out)
			})
		}
	})
}

func Test_NestedPrefix(t *testing.T) {
	type testCase struct {
		title string
		in    []string
		out   string
	}
	var cases = []testCase{
		{"empty base string", []string{"", "abc"}, "abc"},
		{"not empty base string", []string{"abc", "def"}, "abc def"},
	}
	Convey("Nested Prefix", t, func() {
		for _, c := range cases {
			Convey(c.title, func() {
				So(nestedPrefix(c.in[0], c.in[1]), ShouldEqual, c.out)
			})
		}
	})
}

func Test_FlagName(t *testing.T) {
	type In struct {
		field  reflect.StructField
		prefix string
	}
	type testCase struct {
		title string
		in    In
		out   string
	}
	var cases = []testCase{
		{
			"default flag name",
			In{
				reflect.StructField{
					Name: "Test",
					Tag:  "",
					Type: reflect.TypeOf(""),
				},
				"Db",
			},
			"db-test",
		},
		{
			"with provided flag",
			In{
				reflect.StructField{
					Name: "Test",
					Tag:  "default:\"10\" " + keyFlagTag + ":\"test-flag\"",
					Type: reflect.TypeOf(""),
				},
				"Db",
			},
			"test-flag",
		},
	}
	Convey("Flag Name", t, func() {
		for _, c := range cases {
			Convey(c.title, func() {
				So(flagName(c.in.field, c.in.prefix), ShouldEqual, c.out)
			})
		}
	})
}

func Test_EnvName(t *testing.T) {
	EnvPrefix = "TEST"

	type In struct {
		field  reflect.StructField
		prefix string
	}
	type testCase struct {
		title string
		in    In
		out   string
	}
	var cases = []testCase{
		{
			"with default tags",
			In{
				reflect.StructField{
					Name: "Test",
					Tag:  "",
					Type: reflect.TypeOf(""),
				},
				"Db",
			},
			EnvPrefix + "_DB_TEST",
		},
		{
			"with provided tags",
			In{
				reflect.StructField{
					Name: "Test",
					Tag:  "default:\"10\" " + keyEnvTag + ":\"TEST_ENV\"",
					Type: reflect.TypeOf(""),
				},
				"Db",
			},
			"TEST_ENV",
		},
	}
	Convey("Environment values", t, func() {
		for _, c := range cases {
			Convey(c.title, func() {
				So(envName(c.in.field, c.in.prefix), ShouldEqual, c.out)
			})
		}
	})
}

func Test_SetValue(t *testing.T) {
	type testStruct struct {
		D   time.Duration
		I   int
		I64 int64
		U   uint
		U64 uint64
		S   string
		B   bool
		F32 float32
		F64 float64
	}
	type in struct {
		field         reflect.Value
		flgKey, value string
	}
	type testCase struct {
		title string
		in    in
		out   interface{}
		err   error
	}
	var reflectStruct = reflect.Indirect(reflect.ValueOf(new(testStruct)))
	var cases = []testCase{
		{
			title: "time.Duration",
			in: in{
				reflectStruct.FieldByName("D"),
				"flag-test",
				"3h",
			},
			out: time.Duration(10800000000000),
		},
		{
			title: "int value",
			in: in{
				reflectStruct.FieldByName("I"),
				"flag-test",
				"123",
			},
			out: int(123),
		},
		{
			title: "int64 value",
			in: in{
				reflectStruct.FieldByName("I64"),
				"flag-test",
				"234",
			},
			out: int64(234),
		},
		{
			title: "uint value",
			in: in{
				reflectStruct.FieldByName("U"),
				"flag-test",
				"345",
			},
			out: uint(345),
		},
		{
			title: "uint64 value",
			in: in{
				reflectStruct.FieldByName("U64"),
				"flag-test",
				"456",
			},
			out: uint64(456),
		},
		{
			title: "float64 value",
			in: in{
				reflectStruct.FieldByName("F64"),
				"flag-test",
				"567.89",
			},
			out: float64(567.89),
		},
		{
			title: "bool value",
			in: in{
				reflectStruct.FieldByName("B"),
				"flag-test",
				"true",
			},
			out: true,
		},
		{
			title: "string value",
			in: in{
				reflectStruct.FieldByName("S"),
				"flag-test",
				"test string",
			},
			out: "test string",
		},
		{
			title: "unsupported float32 value",
			in: in{
				reflectStruct.FieldByName("F32"),
				"flag-test",
				"3.14159",
			},
			out: 0,
			err: errUnsupportedType(reflectStruct.FieldByName("F32").Kind().String()),
		},
		{
			title: "wrong time.Duration",
			in: in{
				reflectStruct.FieldByName("D"),
				"flag-test",
				"wrong",
			},
			out: time.Duration(10800000000000),
			err: errCantUse("wrong", *new(time.Duration)),
		},
		{
			title: "wrong int value",
			in: in{
				reflect.Indirect(
					reflect.ValueOf(new(testStruct)),
				).FieldByName("I"),
				"flag-test",
				"wrong",
			},
			out: 0,
			err: errCantUse("wrong", *new(int)),
		},
		{
			title: "wrong uint value",
			in: in{
				reflect.Indirect(
					reflect.ValueOf(new(testStruct)),
				).FieldByName("U"),
				"flag-test",
				"wrong",
			},
			out: 0,
			err: errCantUse("wrong", *new(uint)),
		},
		{
			title: "wrong int64 value",
			in: in{
				reflect.Indirect(
					reflect.ValueOf(new(testStruct)),
				).FieldByName("I64"),
				"flag-test",
				"wrong",
			},
			out: 0,
			err: errCantUse("wrong", *new(int64)),
		},
		{
			title: "wrong uint64 value",
			in: in{
				reflect.Indirect(
					reflect.ValueOf(new(testStruct)),
				).FieldByName("U64"),
				"flag-test",
				"wrong",
			},
			out: 0,
			err: errCantUse("wrong", *new(uint64)),
		},
		{
			title: "wrong float64 value",
			in: in{
				reflect.Indirect(
					reflect.ValueOf(new(testStruct)),
				).FieldByName("F64"),
				"flag-test",
				"wrong",
			},
			out: 0,
			err: errCantUse("wrong", *new(float64)),
		},
		{
			title: "wrong bool value",
			in: in{
				reflect.Indirect(
					reflect.ValueOf(new(testStruct)),
				).FieldByName("B"),
				"flag-test",
				"wrong",
			},
			out: false,
		},
	}
	Convey("Setting values", t, func() {
		for _, c := range cases {
			Convey(c.title, func() {
				flagSet := flag.NewFlagSet("config", flag.ContinueOnError)
				err := setValue(c.in.field, flagSet, c.in.flgKey, c.in.value)
				So(c.in.field.Interface(), ShouldEqual, c.out)
				So(err, ShouldResemble, c.err)
			})
		}
	})
}

func Test_Init(t *testing.T) {
	Convey("InitConfig", t, func() {
		type testCase struct {
			title  string
			config interface{}
			prefix string
			error  interface{}
		}
		var tc = []testCase{
			{
				title:  "only pointer to struct is supported",
				config: new(int64),
				prefix: emptyPrefix,
				error:  errInvalidReceiver,
			},
			{
				title: "settability of unexported fields",
				config: &struct {
					value int
				}{},
				prefix: emptyPrefix,
				error:  errCantSet,
			},
			{
				title: "settability of nested unexported fields",
				config: &struct {
					Nested struct {
						value int
					}
				}{},
				prefix: emptyPrefix,
				error:  errCantSet,
			},
			{
				title: "unsupported type",
				config: &struct {
					Value float32 `default:"3.14159"`
				}{},
				prefix: emptyPrefix,
				error:  errUnsupportedType("float32"),
			},
			{
				title: "nested struct unsupported type",
				config: &struct {
					Struct struct {
						Value float32 `default:"3.14159"`
					}
				}{},
				prefix: emptyPrefix,
				error:  errUnsupportedType("float32"),
			},
			{
				title: "check required value",
				config: &struct {
					Value int `required:"true"`
				}{},
				prefix: emptyPrefix,
				error:  errMissingRequired("value"),
			},
			{
				title: "set struct default value",
				config: &struct {
					Value int `default:"42"`
				}{},
				prefix: emptyPrefix,
				error:  nil,
			},
			{
				title: "set nested struct default value",
				config: &struct {
					Struct struct {
						Value int `default:"42"`
					}
				}{},
				prefix: emptyPrefix,
				error:  nil,
			},
		}
		for _, c := range tc {
			Convey(c.title, func() {
				err := Init(c.config, c.prefix)
				So(err, ShouldResemble, c.error)
			})
		}
	})
}