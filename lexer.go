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
	PTOG
	CO
	BO

	//REPEAT STITCHES
	KNIT_REPEAT
	PURL_REPEAT

	//Cables
	CABLE_RC 		// Cable hacia adelante/derecha (e.g., 2/2 RC)
    CABLE_LC 		// Cable hacia atrás/izquierda (e.g.,/ 2/2 LC)
    PURL_CABLE_RC 	// Cable de revés a la derecha
    PURL_CABLE_LC 	// Cable de revés a la izquierda

	REP
	SECTION
	REPBLOCK

	SEMICOLON
	PLACEMARKER
	REMOVEMARKER
	NEG
	COMMENT
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
	PTOG:		"PTOG",
	CO:			"CO",
	BO:			"BO",
	
	KNIT_REPEAT:		"KNIT_REPEAT",
	PURL_REPEAT:		"PURL_REPEAT",

	// CABLES
	CABLE_RC: "CABLE_RC",
	CABLE_LC: "CABLE_LC",
	PURL_CABLE_RC: "PURL_CABLE_RC",
	PURL_CABLE_LC: "PURL_CABLE_LC",

	PLACEMARKER: "PLACEMARKER",
	REMOVEMARKER: "REMOVEMARKER",
	IDENT:		"IDENT",
	INT:		"INT",
	REP:		"REP",
	SECTION:	"SECTION",
	SEMICOLON:	"SEMICOLON",
	NEG:		"NEG",
	REPBLOCK:	"REPBLOCK",
	COMMENT: 	"COMMENT",
}

func (t Token) String() string{
	return tokens[t]
}

func (t Token) isStitch() bool{
	if (t == KNIT || t == PURL || t == YO || t == SSK || t == KTOG || t == PTOG || t == CO || t == BO || t == CABLE_LC || t == CABLE_RC || t == PURL_CABLE_LC || t == PURL_CABLE_RC || t == KNIT_REPEAT || t == PURL_REPEAT){
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
		case '/':
			next, _, err := l.reader.ReadRune()
			if err == nil {
				if next == '/' {
					_ = l.lexComment() 
					continue          
				}
				l.backup()
				return l.pos, ILLEGAL, "/"
			}
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
				case isCableFwd(lit):
					return startPos, CABLE_RC, l.lexCableParams(lit) 
				case isCableBkd(lit):
					return startPos, CABLE_LC, l.lexCableParams(lit)
				case isPurlCableFwd(lit):
					return startPos, PURL_CABLE_RC, l.lexCableParams(lit) 
				case isPurlCableBkd(lit):
					return startPos, PURL_CABLE_LC, l.lexCableParams(lit)
				case isPurlCableBkd(lit):
					return startPos, PURL_CABLE_LC, l.lexCableParams(lit)
				case lit == "repeat":
					return startPos, REPBLOCK, "REPBLOCK"
				case lit == "section":
					return startPos, SECTION, "SECTION"
				case isKtog(lit):
					return startPos, KTOG, l.lexKtog(lit)
				case isPtog(lit):
					return startPos, PTOG, l.lexKtog(lit)
				case isCo(lit):
					return startPos, CO, l.lexCo(lit)
				case isBo(lit):
					return startPos, BO, l.lexBo(lit)
				case lit == "ssk":
					return startPos, SSK, "SSK"
				case lit == "yo":
					return startPos, YO, "YO"
				case isRemoveMarker(lit):
					return startPos, REMOVEMARKER, l.lexMarkerName(lit)
				case isPlaceMarker(lit):
					return startPos, PLACEMARKER, l.lexMarkerName(lit)
				case isKnitRepeat(lit):
					return startPos, KNIT_REPEAT, l.lexRepeatCount(lit)
				case isPurlRepeat(lit):
					return startPos, PURL_REPEAT, l.lexRepeatCount(lit)
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

func (l *Lexer) lexComment() string {
    var lit string
    for {
        r, _, err := l.reader.ReadRune()
        if err != nil || r == '\n' {
            l.resetPosition()
            break
        }
        lit += string(r)
        l.pos.column++
    }
    return lit
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
		if unicode.IsLetter(r) || r == '_' || r == '/' || unicode.IsDigit(r){
			lit = lit + string(r)
		}else {
			l.backup()
			return lit
		}
	}
}

func (l *Lexer) lexKtog(lit string) string{
	return string(lit[1])
}

func (l *Lexer) lexCo(lit string) string {
	return lit[2:]
}

func (l *Lexer) lexBo(lit string) string {
	return lit[2:]
}

func (l *Lexer) lexRepeatCount(lit string) string {
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

func isPtog(lit string) bool {
	reg, _ := regexp.Compile("^p[0-9]tog")
	return reg.MatchString(lit)
}

func isCo(lit string) bool {
	reg, _ := regexp.Compile(`^co[0-9]+$`)
	return reg.MatchString(lit)
}

func isBo(lit string) bool {
	reg, _ := regexp.Compile(`^bo[0-9]+$`)
	return reg.MatchString(lit)
}

func isCableFwd(lit string) bool {
    reg, _ := regexp.Compile(`^c[0-9]+r$|^c[0-9]+/[0-9]+r$`)
    return reg.MatchString(lit)
}

func isCableBkd(lit string) bool {
    reg, _ := regexp.Compile(`^c[0-9]+l$|^c[0-9]+/[0-9]+l$`)
    return reg.MatchString(lit)
}

func isPurlCableFwd(lit string) bool {
    reg, _ := regexp.Compile(`^p[0-9]+r$|^p[0-9]+/[0-9]+r$`)
    return reg.MatchString(lit)
}

func isPurlCableBkd(lit string) bool {
    reg, _ := regexp.Compile(`^p[0-9]+l$|^p[0-9]+/[0-9]+l$`)
    return reg.MatchString(lit)
}

func (l *Lexer) lexCableParams(lit string) string {
    re := regexp.MustCompile(`[0-9]+`)
    matches := re.FindAllString(lit, -1)
    
    if len(matches) == 1 {
        return matches[0] + "," + matches[0]
    } else if len(matches) >= 2 {
        return matches[0] + "," + matches[1]
    }
    return "1,0" 
}

func isKnitRepeat(lit string) bool {
    reg, _ := regexp.Compile(`^k[0-9]+$`)
    return reg.MatchString(lit)
}

func isPurlRepeat(lit string) bool {
    reg, _ := regexp.Compile(`^p[0-9]+$`)
    return reg.MatchString(lit)
}

