package parser

import (
	"fmt"
	"strings"

	"github.com/skalt/git-cc/internal/utils"
)

// A parsed Conventional Commit (CC). See https://www.conventionalcommits.org/en/v1.0.0/
// for more details about what a CC consists of.
type CC struct {
	// A noun such as feat, fix, etc. that describes what kind of change this commit introduces.
	Type string
	// An optional noun describing what part of the codebase was changed.
	Scope *string
	// A short summary of the changes in the commit
	Description string
	// free-form description of the changes; possibly multiple paragraphs.
	Body           *string
	Footers        []string
	BreakingChange bool
}

func trimWhitespace(s string) string {
	return strings.Trim(s, "\n\r\t ")
}

type ResultType string

var (
	TypeCommit         ResultType = "CommitType"
	TypeScope          ResultType = "Scope"
	TypeBang           ResultType = "BreakingChangeBang"
	TypeDescription    ResultType = "Description"
	TypeBody           ResultType = "Body"
	TypeFooter         ResultType = "Footer"
	TypeFooters        ResultType = "Footers"
	TypeBreakingChange ResultType = "BreakingChange"
)

func (cc *CC) Ingest(r Result) *CC {
	switch r.Type {
	case TypeCommit:
		cc.Type = r.Value
	case TypeScope:
		cc.Scope = &r.Value
	case TypeBang:
		cc.BreakingChange = true
	case TypeDescription:
		cc.Description = trimWhitespace(r.Value)
	case TypeBody:
		b := trimWhitespace(r.Value)
		cc.Body = &b
	case TypeFooters:
		footers := []string{}
		for _, footer := range r.Children {
			for _, footerPart := range footer.Children {
				if footerPart.Type == TypeBreakingChange {
					cc.BreakingChange = true
				}
			}
			footers = append(footers, trimWhitespace(footer.Value))
		}
		cc.Footers = footers
	}
	return cc
}

// deprecated.
func (cc *CC) ToString() string {
	return cc.String()
}

func (cc *CC) BreakingChanges() (result []string) {
	if cc.BreakingChange {
		result = make([]string, len(cc.Footers))
		for _, f := range cc.Footers {
			if bc, _ := BreakingChange([]rune(f)); bc != nil {
				result = append(result, f)
			}
		}
	}
	return result
}

func (cc *CC) Render(s *strings.Builder) { // TODO: make private
	s.WriteString(cc.Type)
	if cc.Scope != nil {
		utils.Must(fmt.Fprintf(s, "(%s)", *cc.Scope))
	}
	if cc.BreakingChange {
		utils.Must(s.WriteString("!"))
	}
	utils.Must(s.WriteString(": "))
	utils.Must(s.WriteString(cc.Description))
	utils.Must(s.WriteString("\n\n"))
	if cc.Body != nil {
		utils.Must(s.WriteString(trimWhitespace(*cc.Body)))
		utils.Must(s.WriteString("\n\n"))
	}
	for _, footer := range cc.Footers {
		utils.Must(s.WriteString(trimWhitespace(footer) + "\n"))
	}
}

// String implements [strings.Stringer]
func (cc *CC) String() string { return utils.Render(cc.Render) }

// import constants?
// https://www.conventionalcommits.org/en/v1.0.0/#specification
var Newline = Marked("Newline")(Any(LiteralRune('\n'), Tag("\r\n")))

var DoubleNewline = Sequence(Newline, Newline)
var ColonSep = Regex(": ?") // accept a colon with or without a space after it

// The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in RFC 2119.

// Commits MUST be prefixed with a type, which consists of a noun, feat, fix, etc., followed by the OPTIONAL scope, OPTIONAL !, and REQUIRED terminal colon and space.
// The type feat MUST be used when a commit adds a new feature to your application or library.
// The type fix MUST be used when a commit represents a bug fix for your application.

// A description MUST immediately follow the colon and space after the type/scope prefix. The description is a short summary of the code changes, e.g., fix: array parsing issue when multiple spaces were contained in string.

var CommitType Parser = Marked(TypeCommit)(
	TakeUntil(Any(BreakingChangeBang, Tag(":"), Tag("("), Newline, Empty)),
)

// A scope MAY be provided after a type. A scope MUST consist of a noun describing a section of the codebase surrounded by parenthesis, e.g., fix(parser):
var Scope Parser = Marked(TypeScope)(Delimited(Tag("("), TakeUntil(Tag(")")), Tag(")")))
var BreakingChangeBang Parser = Marked(TypeBang)(Tag("!"))
var ShortDescription Parser = Marked(TypeDescription)(TakeUntil(Any(Empty, Newline)))

// The bit before the description, e.g. "feat", "fix(scope)", "refactor!", etc.
var Context = Sequence(CommitType, Opt(Scope), Opt(BreakingChangeBang))

// See https://www.conventionalcommits.org/en/v1.0.0/#specification:~:text=If,change,-%2E
// , https://www.conventionalcommits.org/en/v1.0.0/#specification:~:text=BREAKING%2DCHANGE%20MUST%20be%20synonymous%20with%20BREAKING%20CHANGE%2C%20when%20used%20as%20a%20token%20in%20a%20footer
var BreakingChange = Any(Tag("BREAKING CHANGE"), Tag("BREAKING-CHANGE"))

var KebabWord = Regex(`[\w-]+`)
var FooterToken = Any(
	Marked(TypeBreakingChange)(Sequence(BreakingChange, ColonSep)),
	Sequence(KebabWord, Any(ColonSep, Tag(" #"))),
)

var Body = Marked(TypeBody)(TakeUntil(Any(Empty, FooterToken)))
var Footer = Marked(TypeFooter)(
	Sequence(FooterToken, TakeUntil(Any(Empty, FooterToken))),
)
var Footers = Marked(TypeFooters)(Many0(Footer))

var asMuchOfScopeAsPossible = Marked(TypeScope)(
	Delimited(
		Tag("("),
		TakeUntil(Any(Tag(")"), Empty, Newline, Tag(":"), Tag("!"))),
		Opt(Tag(")")),
	),
)

var asMuchOfCCAsPossible = Some(
	CommitType, Opt(asMuchOfScopeAsPossible), Opt(BreakingChangeBang), ColonSep, ShortDescription,
	Opt(Newline), Opt(Newline),
	Opt(Body),
	Opt(Footers),
)

func ParseAsMuchOfCCAsPossible(fullCommit string) (result CC, err error) {
	var parsed *Result
	parsed, err = asMuchOfCCAsPossible([]rune(fullCommit))
	if parsed != nil && parsed.Children != nil {
		for _, token := range parsed.Children {
			result = *result.Ingest(token)
		}
	}

	if parsed.Remaining != nil {
		body := utils.Coalesce(result.Body, "") + string(parsed.Remaining)
		result.Body = &body
	}
	return result, err
}
