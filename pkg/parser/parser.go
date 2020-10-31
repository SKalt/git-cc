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

func trimWhitespace(s string) string {
	return strings.Trim(s, "\n\r\t ")
}

func (cc *CC) Ingest(r Result) *CC {
	switch r.Type {
	case "CommitType":
		cc.Type = r.Value
	case "Scope":
		cc.Scope = r.Value
	case "BreakingChangeBang":
		cc.BreakingChange = true
	case "Description":
		cc.Description = trimWhitespace(r.Value)
	case "Body":
		cc.Body = trimWhitespace(r.Value)
	case "Footers":
		footers := []string{}
		for _, footer := range r.Children {
			for _, footerPart := range footer.Children {
				if footerPart.Type == "BreakingChange" {
					cc.BreakingChange = true
				}
			}
			footers = append(footers, trimWhitespace(footer.Value))
		}
		cc.Footers = footers
	}
	return cc
}

func (cc *CC) ToString() string {
	s := strings.Builder{}
	s.WriteString(cc.Type)
	if cc.Scope != "" {
		s.WriteString(fmt.Sprintf("(%s)", cc.Scope))
	}
	if cc.BreakingChange {
		s.WriteString("!")
	}
	s.WriteString(": ")
	s.WriteString(cc.Description)
	s.WriteString("\n\n")
	body := trimWhitespace(cc.Body)
	if body != "" {
		s.WriteString(body)
		s.WriteString("\n\n")
	}
	for _, footer := range cc.Footers {
		s.WriteString(trimWhitespace(footer) + "\n")
	}
	return s.String()
}

func (cc *CC) MinimallyValid() bool {
	return cc.Type != "" && cc.Description != ""
}

func (cc *CC) ValidCommitType(commitTypes []map[string]string) bool {
	for _, commitType := range commitTypes {
		_, matched := commitType[cc.Type]
		if matched {
			return true
		}
	}
	return false
}

func (cc *CC) ValidScope(knownScopes []map[string]string) bool {
	for _, scope := range knownScopes {
		_, matched := scope[cc.Scope]
		if matched {
			return true
		}
	}
	return len(knownScopes) == 0 && len(cc.Scope) == 0
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

var CommitType Parser = Marked("CommitType")(
	TakeUntil(Any(BreakingChangeBang, Tag(":"), Tag("("), Empty)),
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
var Scope Parser = Marked("Scope")(Delimeted(Tag("("), TakeUntil(Tag(")")), Tag(")")))
var BreakingChangeBang Parser = Marked("BreakingChangeBang")(Tag("!"))
var ShortDescription Parser = Marked("Description")(TakeUntil(Any(Empty, Newline)))

var Context = Sequence(CommitType, Opt(Scope), Opt(BreakingChangeBang))

var BreakingChange = Any(Tag("BREAKING CHANGE"), Tag("BREAKING-CHANGE"))

var KebabWord = Regex(`[\w-]+`)
var FooterToken = Any(
	Marked("BreakingChange")(Sequence(BreakingChange, ColonSep)),
	Sequence(KebabWord, Any(ColonSep, Tag(" #"))),
)

var Body = Marked("Body")(TakeUntil(Any(Empty, FooterToken)))
var Footer = Marked("Footer")(
	Sequence(FooterToken, TakeUntil(Any(Empty, FooterToken))),
)
var Footers = Marked("Footers")(Many0(Footer))

func ParseAsMuchOfCCAsPossible(fullCommit string) (*CC, error) {
	parsed, err := Some(
		CommitType, Opt(Scope), Opt(BreakingChangeBang), ColonSep, ShortDescription,
		Opt(Newline), Opt(Newline),
		Opt(Body),
		Opt(Footers),
	)([]rune(fullCommit))
	result := &CC{}
	if parsed != nil && parsed.Children != nil {
		for _, token := range parsed.Children {
			result = result.Ingest(token)
		}
	}
	return result, err
}
