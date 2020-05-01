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
		result, err := CommitType([]rune(input))
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
		_, err := CommitType([]rune("foo"))
		if err == nil {
			t.Fail()
		}
	})
}

func TestTakeUntil(t *testing.T) {
	var callback = func(chars []rune) interface{} {
		return string(chars)
	}
	var test = func(input string, until rune, output string, remaining string) func(t *testing.T) {
		return func(t *testing.T) {
			result, err := TakeUntil(until, callback)([]rune(input))
			if err != nil {
				fmt.Println("err != nil")
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
		test("abcdef", 'c' /* output: */, "ab" /* remaining: */, "cdef"),
	)
	t.Run(
		"matching at the start of the input works",
		test("abcdef", 'a' /* output: */, "" /* remaining: */, "abcdef"),
	)
	t.Run(
		"matching at the end of the input works",
		test("abcdef", 'f' /* output: */, "abcde" /* remaining: */, "f"),
	)
}
