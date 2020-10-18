package parser

// import contsants?
// https://www.conventionalcommits.org/en/v1.0.0/#specification

var Newline = Any(LiteralRune('\n', ident), Tag("\r\n"))
var DoubleNewline = Sequence(Newline, Newline)
var ColonSep = Tag(": ")

// The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in RFC 2119.

// Commits MUST be prefixed with a type, which consists of a noun, feat, fix, etc., followed by the OPTIONAL scope, OPTIONAL !, and REQUIRED terminal colon and space.
// The type feat MUST be used when a commit adds a new feature to your application or library.
// The type fix MUST be used when a commit represents a bug fix for your application.

// A description MUST immediately follow the colon and space after the type/scope prefix. The description is a short summary of the code changes, e.g., fix: array parsing issue when multiple spaces were contained in string.

func CommitType(extraTypes ...string) Parser {
	commitTypes := []Parser{Tag("feat"), Tag("fix")}
	for _, commitType := range extraTypes {
		commitTypes = append(commitTypes, Tag(commitType))
	}
	return Any(commitTypes...)
}

// A scope MAY be provided after a type. A scope MUST consist of a noun describing a section of the codebase surrounded by parenthesis, e.g., fix(parser):
var Scope = Delimeted(Tag('('), TakeUntil(Tag(')'), ident), Tag(')'))
var BreakingChangeBang = Tag("!")

// TODO: -> fn                                vVv -- pass in configured commit types
var CommitTypeContext = Sequence(CommitType( /**/ ), Opt(Scope), Opt(BreakingChangeBang), ColonSep)
var BreakingChange = Tag("BREAKING CHANGE")
var ShortDescription = TakeUntil(Any(Newline, Empty), ident)

var KebabWord = Regex(`[\w-]+`)
var FooterPrefix = Any(
	Sequence(BreakingChange, ColonSep),
	Sequence(KebabWord, Any(ColonSep, Tag(" #"))),
)

var Body = TakeUntil(Any(Empty, FooterPrefix), ident)
var Footer = TakeUntil(Any(Empty, FooterPrefix), ident)
var Footers = Many0(Footer)
