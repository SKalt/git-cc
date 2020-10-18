package parser

import (
	"fmt"
	"testing"
)

func TestParsingCommitTypes(t *testing.T) {
	var testExpectedMatch = func(
		input string,
		expected string,
		remainder string,
	) func(t *testing.T) {
		result, err := CommitType()([]rune(input))
		return func(t *testing.T) {
			if err != nil {
				fmt.Printf("%v", err)
				t.Fail()
				return
			}
			if result.Output != expected {
				fmt.Printf(
					"expected output %v, not %v\n", expected, result.Output,
				)
				t.Fail()
			}
			if string(result.Remaining) != remainder {
				fmt.Printf(
					"expected remainder %v, not %v\n", expected, result.Output,
				)
				t.Fail()
			}
		}
	}

	t.Run(
		"accepts valid commit types [fix]",
		testExpectedMatch("fix", "fix", ""),
	)

	t.Run(
		"accepts valid commit types [feat]",
		testExpectedMatch("feat", "feat", ""),
	)
	t.Run(
		"watch out: only matches the first runes",
		testExpectedMatch("fixing", "fix", "ing"),
	)
	t.Run("rejects invalid commit types", func(t *testing.T) {
		_, err := CommitType()([]rune("foo"))
		if err == nil {
			t.Fail()
		}
	})
}

func TestTakeUntil(t *testing.T) {
	var callback = func(chars []rune) interface{} {
		return string(chars)
	}
	var test = func(input string, until Parser, output string, remaining string) func(t *testing.T) {
		return func(t *testing.T) {
			result, err := TakeUntil(until, callback)([]rune(input))
			if err != nil {
				t.Fail()
			}
			if result.Output != output {
				fmt.Printf("unexpected output %v (should be %v)\n", result.Output, output)
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
		test("abcdef", LiteralRune('c', callback) /* output: */, "ab" /* remaining: */, "cdef"),
	)
	t.Run(
		"matching at the start of the input works",
		test("abcdef", LiteralRune('a', callback) /* output: */, "" /* remaining: */, "abcdef"),
	)
	t.Run(
		"matching at the end of the input works",
		test("abcdef", LiteralRune('f', callback) /* output: */, "abcde" /* remaining: */, "f"),
	)
}

// these test cases are copied from https://www.conventionalcommits.org/en/v1.0.0/, which is licensed under
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
