package parser

import (
	"fmt"
	"regexp"
)

// see https://medium.com/@armin.heller/using-parser-combinators-in-go-e63b3ad69c94,
// https://github.com/Geal/nom (v5+), and https://bodil.lol/parser-combinators/

// TODO: have parsers output a Concrete Syntax Tree (CST).
// IDEA: have parserresult interface be tree-ish:
// type ResultIdea struct {
// 	Children []ResultIdea
// 	Kind     string  // should be Token or symbol, might be empty in case of ignored chars
// 	Range    [2]int  // start, end. Empty if start == end.
// 	Original *string // pointer to the original input
// 	//? Parent *ResultIdea
// }

// func (r *ResultIdea) OriginalString() string {
// 	return (*r.Original)[r.Range[0]:r.Range[1]]
// }
// func (r *ResultIdea) Remaining() []rune {
// 	return []rune((*r.Original)[r.Range[1]:])
// }

// type ParserIdea func( /*input*/ *string, int /*rune start index*/) (*ResultIdea, error)

type Result struct {
	Children  []Result
	Value     string
	Type      string
	Remaining []rune
}

func (r *Result) CopyTyped(name string) *Result {
	return &Result{
		Children:  r.Children,
		Remaining: r.Remaining,
		Value:     r.Value,
		Type:      name,
	}
}

type ParserCallback func([]rune) interface{}
type Parser func([]rune) (*Result, error)
type void struct{}

func LastRuneOf(parser Parser) Parser {
	return func(input []rune) (*Result, error) {
		maxIndex := len(input) - 1
		if maxIndex < 0 {
			return parser([]rune{})
		} else {
			lastRune := input[maxIndex-1:]
			return parser(lastRune)
		}
	}
}

func TakeUntil(parser Parser) Parser {
	return func(input []rune) (*Result, error) {
		for i := range input {
			_, err := parser(input[i:])
			if err == nil {
				return &Result{
					Value:     string(input[:i]),
					Remaining: input[i:],
				}, nil
			}
		}
		_, err := parser([]rune{})
		if err == nil {
			return &Result{
				Value:     string(input),
				Remaining: []rune{},
			}, nil
		}
		return nil, fmt.Errorf("didn't match parser")
	}
}

func Marked(mark string) func(Parser) Parser {
	if len(mark) == 0 {
		panic("empty mark")
	}
	return func(parser Parser) Parser {
		return func(input []rune) (*Result, error) {
			result, err := parser(input)
			if err != nil {
				return nil, err
			} else {
				return result.CopyTyped(mark), nil
			}
		}
		//
	}
}

func Opt(parser Parser) Parser {
	return func(input []rune) (*Result, error) {
		result, err := parser(input)
		if err == nil {
			return result, nil
		}
		result = &Result{Remaining: input}
		return result, nil
	}
}

func LiteralRune(match rune) Parser {
	return func(input []rune) (*Result, error) {
		if len(input) > 0 {
			if input[0] == match {
				return &Result{
					Value:     string(match),
					Remaining: input[1:],
				}, nil
			} else {
				return nil, fmt.Errorf("%v not matched", match)
			}
		} else {
			return nil, fmt.Errorf("no input")
		}
	}
}

func Not(parser Parser) Parser {
	return func(input []rune) (*Result, error) {
		result, err := parser(input)
		if err == nil {
			return result, nil
		} else {
			return nil, fmt.Errorf("wasn't expecting to match %v", parser)
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

func Tag(tag string) Parser {
	toMatch := []rune(tag)
	return func(input []rune) (*Result, error) {
		if len(toMatch) > len(input) {
			return nil, fmt.Errorf("input longer than tag")
		}
		for i, matching := range toMatch {
			if input[i] != matching {
				err := fmt.Errorf(
					"`%v` does not match `%v`",
					string(input[:i+1]),
					string(toMatch),
				)
				return nil, err
			}
		}
		return &Result{
			Value: string(toMatch), Remaining: input[len(toMatch):],
		}, nil
	}
}

func Any(parsers ...Parser) Parser {
	return func(input []rune) (*Result, error) {
		for _, parser := range parsers {
			result, err := parser(input)
			if err == nil {
				return result, err
			}
		}
		return nil, fmt.Errorf("expected a parser to match")
	}
}

func Empty(input []rune) (*Result, error) {
	if len(input) == 0 {
		return nil, nil
	} else {
		return nil, fmt.Errorf("Not the end")
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
		parsers = append(parsers, LiteralRune(char))
	}
	return Any(parsers...)
}

func asString(result []rune) interface{} {
	return string(result)
}
func clone(input []rune) []rune {
	cloned := make([]rune, len(input))
	copy(cloned, input)
	return cloned
}

func Sequence(parsers ...Parser) Parser {
	return func(input []rune) (*Result, error) {
		var currentInput = make([]rune, len(input))
		children := make([]Result, len(parsers))
		copy(currentInput, input)
		for i, parser := range parsers {
			result, err := parser(currentInput)
			if err != nil {
				return nil, err
			} else {
				currentInput = result.Remaining
				if len(result.Value)+len(result.Type) > 0 {
					children[i] = *result
				}
			}
		}
		value := ""
		for i := range children {
			value = value + children[i].Value
		}
		return &Result{
			Children:  children,
			Value:     value,
			Remaining: currentInput,
		}, nil
	}
}
func Delimeted(start Parser, middle Parser, end Parser) Parser {
	return func(input []rune) (*Result, error) {
		result, err := Sequence(start, middle, end)(input)
		if err != nil {
			return nil, err
		}

		return &Result{
			Value:     result.Children[1].Value,
			Remaining: result.Remaining,
		}, nil
	}
}
func Many0(parser Parser) Parser {
	return func(input []rune) (*Result, error) {
		i := 0
		results := []Result{}
		for { // the highest possible # of times callable
			result, err := parser(input[i:])
			if err != nil {
				break
			}
			results = append(results, *result)
			if len(result.Remaining) == 0 {
				break
			}
		}
		return &Result{Children: results, Remaining: input[i:]}, nil
	}
}

func Many1(parser Parser) Parser {
	return func(input []rune) (*Result, error) {
		result, _ := Many0(parser)(input)
		if len(result.Children) == 0 {
			return nil, fmt.Errorf("no results")
		} else {
			return result, nil
		}
	}
}

func Regex(pattern string) Parser {
	re := regexp.MustCompile(`^` + pattern) // should be from the start of the bytes
	return func(input []rune) (*Result, error) {
		b := []byte(string(input))
		result := re.FindIndex(b) //Match(b)
		if result == nil {        // no match found
			return nil, fmt.Errorf("no match for /%s/", pattern)
		} else {
			// a rune can be multiple bytes, so convert the reult back to runes
			startByte := result[0]
			endByte := result[1]
			endRune := len([]rune(string(b[startByte:endByte])))
			return &Result{
				Value:     string(input[:endRune]),
				Remaining: input[endRune:],
			}, nil
		}
	}
}

// func Range(start rune, end rune) Parser {
// 	return Not(Empty)(
// 		func(input []rune) (ParserResult, error) {
// 			if input[0] >= start && input[0] <= end {
// 				// ok
// 			} else {
// 				return
// 			}
// 		},
// 	)
// }
