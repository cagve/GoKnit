package main

import ( 
	"bufio"
	"io"
	"regexp"
	"unicode"
)

type Token int

const (
	EOF = iota
	ILLEGAL
	INT
	IDENT
	BROPEN
	BRCLOSE
	PAROPEN
	PARCLOSE

	//Stiches
	KNIT
	PURL
	SSK
	YO
	KTOG

	REP
	SECTION
	REPBLOCK

	SEMICOLON
	PLACEMARKER
	REMOVEMARKER
	NEG
)

var tokens = []string{
	EOF:		"EOF",
	ILLEGAL:	"ILLEGAL",

	BROPEN:		"BROPEN",
	BRCLOSE:	"BRCLOSE",
	PAROPEN:	"PAROPEN",
	PARCLOSE:	"PARCLOSE",

	KNIT:		"KNIT",
	PURL:		"PURL",
	SSK:		"SSK",
	YO:			"YO",
	KTOG:		"KTOG",

	PLACEMARKER: "PLACEMARKER",
	REMOVEMARKER: "REMOVEMARKER",
	IDENT:		"IDENT",
	INT:		"INT",
	REP:		"REP",
	SECTION:	"SECTION",
	SEMICOLON:	"SEMICOLON",
	NEG:		"NEG",
	REPBLOCK:	"REPBLOCK",
}

func (t Token) String() string{
	return tokens[t]
}

func (t Token) isStitch() bool{
	if (t == KNIT || t == PURL || t == YO || t == SSK || t == KTOG){
		return true
	} else{
		return false
	}
}

type Position struct {
	line int
	column int
}

type Lexer struct {
	pos Position
	reader *bufio.Reader
}

func NewLexer(reader io.Reader) *Lexer {
	return &Lexer {
		pos: Position {line: 1, column: 0},
		reader: bufio.NewReader(reader),
	}
}

func (l *Lexer) Lex() (Position, Token, string) {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return l.pos, EOF, ""
			}
			panic(err)
		}

		l.pos.column++

		switch r {
		case ';':
			l.resetPosition()
			return l.pos, SEMICOLON, ";"
		case '-':
			return l.pos, NEG, "-"
		case '{':
			return l.pos, BROPEN, "{"
		case '}':
			return l.pos, BRCLOSE, "}"
		case '(':
			return l.pos, PAROPEN, "("
		case ')':
			return l.pos, PARCLOSE, ")"
		case '*':
			return l.pos, REP, "*"
		default:
			if unicode.IsSpace(r){
				continue
			} else if unicode.IsDigit(r) {
				startPos := l.pos
				l.backup() // como ha consumido el token, vuelvo atrs.
				lit := l.lexInt()
				return startPos, INT, lit
			} else if unicode.IsLetter(r) {
				startPos := l.pos
				l.backup()
				lit := l.lexIdent()
				switch {
				case lit == "repeat":
					return startPos, REPBLOCK, "REPBLOCK"
				case lit == "section":
					return startPos, SECTION, "SECTION"
				case isKtog(lit):
					return startPos, KTOG, l.lexKtog(lit)
				case lit == "ssk":
					return startPos, SSK, "SSK"
				case lit == "yo":
					return startPos, YO, "YO"
				case isRemoveMarker(lit):
					return startPos, REMOVEMARKER, l.lexMarkerName(lit)
				case isPlaceMarker(lit):
					return startPos, PLACEMARKER, l.lexMarkerName(lit)
				case lit == "k":
					return l.pos, KNIT, "k"
				case lit == "p":
					return l.pos, PURL, "p"
				default:
					return startPos, IDENT, lit
				}
			} else {
				return l.pos, ILLEGAL, string(r)
			}
		}
	}
}

func (l *Lexer) resetPosition() {
	l.pos.line++
	l.pos.column = 0
}

//Vuelve un posici'on atras en el lexer
func(l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic (err)
	}
	l.pos.column--
}

func (l *Lexer) lexInt() string{
	var lit string
	for {
		r,_,err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				//end of line
				return lit
			}
		}
		l.pos.column++
		if unicode.IsDigit(r) {
			lit = lit + string(r)
		} else {
			l.backup()
			return lit
		}
	}
}

func (l *Lexer) lexIdent() string {
	var lit string 
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return lit
			}
		}

		l.pos.column ++ 
		if unicode.IsLetter(r) || r == '_' || unicode.IsDigit(r){
			lit = lit + string(r)
		}else {
			l.backup()
			return lit
		}
	}
}

func (l *Lexer) lexKtog(lit string) string{
	// Devuelvee el char en segunda posicion. kntog
	return string(lit[1])
}
func (l *Lexer) lexMarkerName(lit string) string{
	// Devuelve el ultimo char de la cadena. mA mB ...
	return string(lit[len(lit)-1:]) 
}

func isPlaceMarker(lit string) bool {
	reg, _ := regexp.Compile("^m[A-Z]")
	return reg.MatchString(lit)
}

func isRemoveMarker(lit string) bool{
	reg, _ := regexp.Compile("^rm[A-Z]")
	return reg.MatchString(lit)
}

func isKtog(lit string) bool {
	reg, _ := regexp.Compile("^k[0-9]tog")
	return reg.MatchString(lit)
}

