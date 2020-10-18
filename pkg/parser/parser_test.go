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
var validCCWithOnlyDescription = `docs: correct spelling of CHANGELOG`
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
			actual, err := ParseCC(fullValidCommit)
			if err != nil {
				fmt.Printf("%+v", err)
				t.FailNow()
			}
			if actual.Type != expected.Type {
				fmt.Printf("expected: %+v actual: %+v", expected.Type, actual.Type)
				t.FailNow()
			}
			if actual.Scope != expected.Scope {
				fmt.Printf("expected: %+v actual: %+v", expected.Type, actual.Type)
				t.FailNow()
			}
			if actual.BreakingChange != expected.BreakingChange {
				fmt.Printf("expected: %+v actual: %+v", expected.BreakingChange, actual.BreakingChange)
				t.FailNow()
			}
			if actual.Body != expected.Body {
				fmt.Printf("expected: %+v actual: %+v", expected.Body, actual.Body)
				t.FailNow()
			}
			if len(actual.Footers) != len(expected.Footers) {
				fmt.Printf("expected: %+v actual: %+v", expected.Footers, actual.Footers)
				t.FailNow()
			}
			for i := range actual.Footers {
				if actual.Footers[i] != expected.Footers[i] {
					fmt.Printf("expected: '%+v' actual: '%+v'", expected.Footers[i], actual.Footers[i])
					t.FailNow()
				}
			}
		}
	}
	t.Run("can parse a valid cc with a breaking change footer",
		test(validCCwithBreakingChangeFooter, CC{
			Type:        "feat",
			Scope:       "",
			Description: "allow provided config object to extend other configs",
			Body:        "",
			Footers: []string{
				"BREAKING CHANGE: `extends` key in config file is now used for extending other config files",
			},
		}))
}
