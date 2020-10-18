package parser

import (
	"fmt"
	"strings"
)

type CC struct {
	Type           string
	Scope          string
	Description    string
	Body           string
	Footers        []string
	BreakingChange bool
}

type CCHeader struct {
	Type           string
	Scope          string
	Description    string
	BreakingChange bool
}
type CCRest struct {
	Body           string
	Footers        []string
	BreakingChange bool
}

// import contsants?
// https://www.conventionalcommits.org/en/v1.0.0/#specification
var Newline = Marked("Newline")(Any(LiteralRune('\n'), Tag("\r\n")))

var DoubleNewline = Sequence(Newline, Newline)
var ColonSep = Tag(": ")

// The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in RFC 2119.

// Commits MUST be prefixed with a type, which consists of a noun, feat, fix, etc., followed by the OPTIONAL scope, OPTIONAL !, and REQUIRED terminal colon and space.
// The type feat MUST be used when a commit adds a new feature to your application or library.
// The type fix MUST be used when a commit represents a bug fix for your application.

// A description MUST immediately follow the colon and space after the type/scope prefix. The description is a short summary of the code changes, e.g., fix: array parsing issue when multiple spaces were contained in string.

var CommitType = Marked("CommitType")(
	TakeUntil(Any(BreakingChangeBang, Tag(":"), Tag("("))),
)

// func CommitTypeParser(extraTypes ...string) Parser {
// 	// TODO: considuer using TakeUntil(Any(BreakingBang, Tag(":")))
// 	commitTypes := []Parser{Tag("feat"), Tag("fix")}
// 	for _, commitType := range extraTypes {
// 		commitTypes = append(commitTypes, Tag(commitType))
// 	}
// 	return Marked("CommitType")(Any(commitTypes...))
// }

// A scope MAY be provided after a type. A scope MUST consist of a noun describing a section of the codebase surrounded by parenthesis, e.g., fix(parser):
var Scope = Marked("Scope")(Delimeted(Tag("("), TakeUntil(Tag(")")), Tag(")")))

var BreakingChangeBang = Marked("BreakingChangeBang")(Tag("!"))
var Context = Sequence(CommitType, Opt(Scope), Opt(BreakingChangeBang))

func ParseHead(head []rune) (*CCHeader, error) {
	header := CCHeader{}
	ctx, ctxErr := Context(head)
	if ctxErr != nil {
		return &header, ctxErr
	}
	switch true {
	case len(ctx.Children) > 0:
		header.Type = ctx.Children[0].Value
	case len(ctx.Children) > 1:
		header.Scope = ctx.Children[1].Value
	case len(ctx.Children) > 2:
		header.BreakingChange = ctx.Children[2].Type == "BreakingChangeBang"
	}
	//
	desc, descErr := Sequence(ColonSep, TakeUntil(Empty))(ctx.Remaining)
	if descErr == nil {
		header.Description = desc.Children[1].Value
	}
	return &header, descErr
}

var BreakingChange = Tag("BREAKING CHANGE: ")

var KebabWord = Regex(`[\w-]+`)
var FooterPrefix = Any(
	BreakingChange,
	Sequence(KebabWord, Any(ColonSep, Tag(" #"))),
)

var Body = Sequence(Newline, TakeUntil(Any(Empty, FooterPrefix)))
var Footer = Sequence(FooterPrefix, TakeUntil(Any(Empty, FooterPrefix)))
var Footers = Many0(Footer)

func ParseRest(input []rune) (*CCRest, error) {
	rest := &CCRest{}
	result, err := Body(input)
	if err != nil {
		return rest, err
	}
	rest.Body = result.Children[1].Value
	result, err = Footers(result.Remaining)
	if err != nil {
		return rest, err
	}
	footers := make([]string, len(result.Children))
	breakingChange := false
	for i := range result.Children {
		child := result.Children[i]
		if child.Type == "BreakingChange" {
			breakingChange = true
		}
		footers[i] = child.Value
	}
	rest.BreakingChange = breakingChange
	rest.Footers = footers
	return rest, err
}

func firstLine(s string) (string, string) {
	result := strings.SplitN(s, "\r\n", 2)
	if len(result) == 1 {
		result = strings.SplitN(s, "\n", 2)
	}
	if len(result) == 1 {
		return result[0], ""
	} else {
		return result[0], result[1]
	}
}

func ParseCC(fullCommit string) (*CC, error) {
	cc := &CC{}
	firstLine, otherLines := firstLine(fullCommit)
	if len(firstLine) == 0 {
		return cc, fmt.Errorf("empty commit")
	}

	header, headerErr := ParseHead([]rune(firstLine))
	if headerErr != nil {
		panic(headerErr)
	}
	cc.Type = header.Type
	cc.Scope = header.Scope
	cc.BreakingChange = header.BreakingChange
	otherLines = strings.TrimRight(otherLines, "\n\r\t ")
	rest, restErr := ParseRest([]rune(otherLines))
	if restErr != nil {
		panic(restErr)
	}
	cc.Body = rest.Body
	cc.Footers = rest.Footers
	cc.BreakingChange = cc.BreakingChange || rest.BreakingChange
	return cc, nil
}
