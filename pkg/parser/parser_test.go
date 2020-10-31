package parser

import (
	"fmt"
	"testing"
)

// these example _s are copied from https://www.conventionalcommits.org/en/v1.0.0/, which is licensed under
// CC-BY 3.0 (https://creativecommons.org/licenses/by/3.0/).  I've added escape sequences to record then as go multiline string literals.
var validCCwithBreakingChangeFooter = `feat: allow provided config object to extend other configs

BREAKING CHANGE: ` + "`extends`" + ` key in config file is now used for extending other config files
`
var validCCWithBreakingChangeBang = `refactor!: drop support for Node 6`
var validCCwithBothBreakingChangeBangAndFooter = `refactor!: drop support for Node 6

BREAKING CHANGE: refactor to use JavaScript features not available in Node 6.

`
var validCCWithOnlyHeader = `docs: correct spelling of CHANGELOG`
var validCCWithScope = `feat(lang): add polish language`
var validCCWithFooters = `fix: correct minor typos in code

see the issue for details

on typos fixed.

Reviewed-by: Z
Refs #133`

var validCCreversion = `revert: let us never again speak of the noodle incident

Refs: 676104e, a215868`

// end of copied material

func TestParsingCommitTypes(t *testing.T) {
	var testExpectedMatch = func(
		input string,
		expected string,
		remainder string,
	) func(t *testing.T) {
		result, err := CommitType([]rune(input))
		return func(t *testing.T) {
			if err != nil {
				fmt.Printf("%v", err)
				t.Fail()
				return
			}
			if result.Value != expected {
				fmt.Printf(
					"expected output %v, not %v\n", expected, result.Value,
				)
				t.Fail()
			}
			if string(result.Remaining) != remainder {
				fmt.Printf(
					"expected remainder %v, not %v\n", expected, result.Value,
				)
				t.Fail()
			}
		}
	}

	t.Run(
		"accepts valid commit types [fix]",
		testExpectedMatch("fix: ", "fix", ": "),
	)

	t.Run(
		"accepts valid commit types [feat]",
		testExpectedMatch("feat!", "feat", "!"),
	)
	t.Run(
		"watch out: only matches the first runes",
		testExpectedMatch("fix(", "fix", "("),
	)
}

func TestOpt(t *testing.T) {
	t.Run("When match is present", func(t *testing.T) {
		result, err := Opt(Tag("("))([]rune("(scope)"))
		if err != nil {
			t.Fail()
		}
		if result.Value != "(" {
			t.Fail()
		}
		if string(result.Remaining) != "scope)" {
			t.Fail()
		}
	})
	t.Run("When match is missing", func(t *testing.T) {
		result, err := Opt(Tag("("))([]rune(": desc"))
		if err != nil {
			t.Fail()
		}
		if result.Value != "" {
			t.Fail()
		}
		if string(result.Remaining) != ": desc" {
			t.Fail()
		}
	})
}

func TestTakeUntil(t *testing.T) {
	var test = func(input string, until Parser, output string, remaining string) func(t *testing.T) {
		return func(t *testing.T) {
			result, err := TakeUntil(until)([]rune(input))
			if err != nil {
				t.Fail()
			}
			if result.Value != output {
				fmt.Printf("unexpected output %v (should be %v)\n", result.Value, output)
				t.Fail()
			}
			if string(result.Remaining) != remaining {
				fmt.Printf("unexpected remaining \"%v\" (should be %v)\n", string(result.Remaining), remaining)
				t.Fail()
			}
		}
	}
	t.Run(
		"matching in the middle of the input works",
		test("abcdef", LiteralRune('c') /* output: */, "ab" /* remaining: */, "cdef"),
	)
	t.Run(
		"matching at the start of the input works",
		test("abcdef", LiteralRune('a') /* output: */, "" /* remaining: */, "abcdef"),
	)
	t.Run(
		"matching at the end of the input works",
		test("abcdef", LiteralRune('f') /* output: */, "abcde" /* remaining: */, "f"),
	)
}

func TestParsingFullCommit(t *testing.T) {
	test := func(fullValidCommit string, expected CC) func(*testing.T) {
		return func(t *testing.T) {
			actual, err := ParseAsMuchOfCCAsPossible(fullValidCommit)
			if err != nil {
				fmt.Printf("%+v", err)
				t.FailNow()
			}
			if actual.Type != expected.Type {
				fmt.Printf("Type: expected: %+v actual: %+v\n", expected.Type, actual.Type)
				t.FailNow()
			}
			if actual.Scope != expected.Scope {
				fmt.Printf("Scope: expected: %+v actual: %+v\n", expected.Type, actual.Type)
				t.FailNow()
			}
			if actual.BreakingChange != expected.BreakingChange {
				fmt.Printf("BreakingChange: expected: %+v actual: %+v\n", expected.BreakingChange, actual.BreakingChange)
				t.FailNow()
			}
			if actual.Body != expected.Body {
				fmt.Printf("Body: expected: `%+v`\nactual: `%+v`\n", expected.Body, actual.Body)
				t.FailNow()
			}
			if len(actual.Footers) != len(expected.Footers) {
				fmt.Printf("Footers: expected: %+v actual: %+v\n", expected.Footers, actual.Footers)
				t.FailNow()
			}
			for i := range actual.Footers {
				if actual.Footers[i] != expected.Footers[i] {
					fmt.Printf("expected: '%+v' actual: '%+v'\n", expected.Footers[i], actual.Footers[i])
					t.FailNow()
				}
			}
		}
	}
	prefix := "can parse a valid cc with "
	t.Run(prefix+"breaking change footer", test(validCCwithBreakingChangeFooter, CC{
		Type:           "feat",
		Scope:          "",
		Description:    "allow provided config object to extend other configs",
		Body:           "",
		BreakingChange: true,
		Footers: []string{
			"BREAKING CHANGE: `extends` key in config file is now used for extending other config files",
		},
	}))
	t.Run(prefix+"a bang", test(validCCWithBreakingChangeBang, CC{
		Type:           "refactor",
		Scope:          "",
		Description:    "drop support for node 6",
		Body:           "",
		BreakingChange: true,
	}))
	t.Run(prefix+"both a bang and a breaking change footer", test(validCCwithBothBreakingChangeBangAndFooter, CC{
		Type:           "refactor",
		Scope:          "",
		Description:    "drop support for Node 6",
		BreakingChange: true,
		Body:           "",
		Footers: []string{
			"BREAKING CHANGE: refactor to use JavaScript features not available in Node 6.",
		},
	}))
	t.Run(prefix+"only a header", test(validCCWithOnlyHeader, CC{
		Type:           "docs",
		Scope:          "",
		Description:    "correct spelling of CHANGELOG",
		Body:           "",
		Footers:        []string{},
		BreakingChange: false,
	}))
	t.Run(prefix+"no body or footers but a scope", test(validCCWithScope, CC{
		Type:           "feat",
		Scope:          "lang",
		Description:    "add polish language",
		Body:           "",
		Footers:        []string{},
		BreakingChange: false,
	}))
	t.Run(prefix+"footers", test(validCCWithFooters, CC{
		Type:        "fix",
		Scope:       "",
		Description: "correct minor typos in code",
		Body: `see the issue for details

on typos fixed.`,
		Footers: []string{
			"Reviewed-by: Z",
			"Refs #133"},
		BreakingChange: false,
	}))
	t.Run(prefix+"reversion", test(validCCreversion, CC{
		Type:           "revert",
		Scope:          "",
		Description:    "let us never again speak of the noodle incident",
		Body:           "",
		Footers:        []string{"Refs: 676104e, a215868"},
		BreakingChange: false,
	}))
	// // template:
	// t.Run(prefix+"", test(,CC{
	//     Type:        "",
	//     Scope:       "",
	//     Description: "",
	//     Body:        "",
	//     Footers:     []string{},
	//     BreakingChange: false,
	// }))
}

func TestParsingPartialCommit(t *testing.T) {
	test := func(partialCommit string, expected CC) func(*testing.T) {
		return func(t *testing.T) {
			actual, _ := ParseAsMuchOfCCAsPossible(partialCommit)
			if actual.Type != expected.Type {
				fmt.Printf("Type: expected: %+v actual: %+v\n", expected.Type, actual.Type)
				t.FailNow()
			}
			if actual.Scope != expected.Scope {
				fmt.Printf("Scope: expected: %+v actual: %+v\n", expected.Type, actual.Type)
				t.FailNow()
			}
			if actual.BreakingChange != expected.BreakingChange {
				fmt.Printf("BreakingChange: expected: %+v actual: %+v\n", expected.BreakingChange, actual.BreakingChange)
				t.FailNow()
			}
			if actual.Body != expected.Body {
				fmt.Printf("Body: expected: `%+v`\nactual: `%+v`\n", expected.Body, actual.Body)
				t.FailNow()
			}
			if len(actual.Footers) != len(expected.Footers) {
				fmt.Printf("Footers: expected: %+v actual: %+v\n", expected.Footers, actual.Footers)
				t.FailNow()
			}
			for i := range actual.Footers {
				if actual.Footers[i] != expected.Footers[i] {
					fmt.Printf("expected: '%+v' actual: '%+v'\n", expected.Footers[i], actual.Footers[i])
					t.FailNow()
				}
			}
		}
	}
	t.Run("", test("feat", CC{Type: "feat"}))
	t.Run("", test("feat:", CC{Type: "feat"}))
	t.Run("", test("feat: ", CC{Type: "feat"}))
}
