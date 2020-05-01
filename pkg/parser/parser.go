package parser

import (
	"fmt"
)

// see https://medium.com/@armin.heller/using-parser-combinators-in-go-e63b3ad69c94
// and https://github.com/Geal/nom

type ParserResult struct {
	Output    interface{}
	Remaining []rune
}
type ParserCallback func([]rune) interface{}
type Parser func([]rune) (ParserResult, error)
type void struct{}

func ident(input []rune) interface{} {
	return []rune(input)
}
func TakeUntil(character rune, callback ParserCallback) Parser {
	return func(input []rune) (ParserResult, error) {
		for i, nextChar := range input {
			if nextChar == character {
				output := callback(input[:i])
				return ParserResult{output, input[i:]}, nil
			}
		}
		return ParserResult{nil, input}, fmt.Errorf("didn't match '%v'", character)
	}
}

func Opt(parser Parser) Parser {
	return func(input []rune) (ParserResult, error) {
		result, _ := parser(input)
		return result, nil
	}
}

func LiteralRune(match rune, callback ParserCallback) Parser {
	return func(input []rune) (ParserResult, error) {
		if len(input) > 0 {
			if input[0] == match {
				return ParserResult{callback([]rune{match}), input[1:]}, nil
			} else {
				return ParserResult{nil, input}, fmt.Errorf("%v not matched", match)
			}
		} else {
			return ParserResult{nil, input}, fmt.Errorf("no input")
		}
	}
}

func NotMatching(parser Parser) Parser {
	return func(input []rune) (ParserResult, error) {
		result, err := parser(input)
		if err == nil {
			return result, nil
		} else {
			return ParserResult{nil, input}, fmt.Errorf("wasn't expecting to match %v", parser)
		}
	}
}

func toRunes(i interface{}) []rune {
	switch i.(type) {
	case string:
		str, _ := i.(string)
		return []rune(str)
	case rune:
		r, _ := i.(rune)
		return []rune{r}
	case []rune:
		runes, _ := i.([]rune)
		return runes
	default:
		panic(fmt.Errorf("%v is not string or []rune", i))
	}
}

func Tag(tag interface{}) Parser {
	toMatch := toRunes(tag)
	return func(input []rune) (ParserResult, error) {
		if len(toMatch) > len(input) {
			return ParserResult{nil, input}, fmt.Errorf("input longer than tag")
		}
		for i, matching := range toMatch {
			if input[i] != matching {
				err := fmt.Errorf(
					"\"%v\" does not match \"%v\"",
					string(input[:i+1]),
					string(toMatch),
				)
				return ParserResult{nil, input}, err
			}
		}
		return ParserResult{string(toMatch), input[len(toMatch):]}, nil
	}
}

func Any(parsers ...Parser) Parser {
	return func(input []rune) (ParserResult, error) {
		for _, parser := range parsers {
			result, err := parser(input)
			if err == nil {
				return result, err
			}
		}
		return ParserResult{nil, input}, fmt.Errorf("expected a parser to match")
	}
}

func OneOfTheseRunes(str string) Parser {
	set := make(map[rune]void)
	var present void
	for _, char := range []rune(str) {
		set[char] = present
	}
	parsers := make([]Parser, len(set))
	for char := range set {
		parsers = append(parsers, LiteralRune(char, ident))
	}
	return Any(parsers...)
}

func asString(result []rune) interface{} {
	return string(result)
}
func clone(input []rune) []rune {
	cloned := []rune{}
	copy(cloned, input)
	return cloned
}
func Sequence(parsers ...Parser) Parser {
	return func(input []rune) (ParserResult, error) {
		var currentInput = []rune{}
		results := make([]interface{}, len(parsers))
		copy(currentInput, input)
		for _, parser := range parsers {
			result, err := parser(currentInput)
			if err != nil {
				return ParserResult{nil, input}, err
			} else {
				currentInput = result.Remaining
				results = append(results, result.Output)
			}
		}
		return ParserResult{results, currentInput}, nil
	}
}

func Delimeted(start Parser, middle Parser, end Parser) Parser {
	return func(input []rune) (ParserResult, error) {
		result, err := Sequence(start, middle, end)(input)
		if err != nil {
			return ParserResult{nil, input}, err
		}
		results, _ := result.Output.([]string)
		return ParserResult{results[1], result.Remaining}, nil
	}
}

func Many0(parser Parser) Parser {
	return func(input []rune) (ParserResult, error) {
		i := 0
		results := make([]interface{}, len(input))
		for i < len(input) { // the highest possible # of times callable
			result, err := parser(input[i:])
			if err != nil {
				break
			}
			results = append(results, result.Output)
			i += len(result.Remaining)
		}
		return ParserResult{results, input[i:]}, nil
	}
}
func Many1(parser Parser) Parser {
	return func(input []rune) (ParserResult, error) {
		result, _ := Many0(parser)(input)
		resultArr := result.Output.([]interface{})
		if len(resultArr) == 0 {
			return result, fmt.Errorf("no results")
		} else {
			return result, nil
		}
	}
}

var CommitType = Any(Tag("feat"), Tag("fix"))
var Scope = Delimeted(
	Tag('('),
	NotMatching(Any(OneOfTheseRunes(")\n\r"))),
	Tag(')'),
)
var BreakingChangeBang = Tag("!")
var CommitTypeContext = Sequence(CommitType, Opt(Scope), Opt(BreakingChangeBang))

func main() {}
