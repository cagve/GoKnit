package main

import (
	"fmt"
	"strconv"
	"strings"
	"io"
)

type Node interface {
	String() string // Function that all nodes implement.
}

type Expr interface {
	Node
}

type Group struct {
	Content []Expr
}

func (g *Group) isGroup() {}
func (g *Group) String() string {
	var exprs []string
	for _, expr  := range g.Content {
		exprs = append(exprs, expr.String())
	}
	return strings.Join(exprs, ", ")
}

type ParsedStitch interface {
	Expr
	isStitch()
}

type ParsedKnit struct {}
func (k *ParsedKnit) isStitch() {}
func (k *ParsedKnit) String() string {return "Knit"}

type ParsedPurl struct {}
func (p *ParsedPurl) isStitch() {}
func (p *ParsedPurl) String() string {return "Purl"}

type ParsedSsk struct {} // REDUCCION
func (s *ParsedSsk) isStitch() {}
func (s *ParsedSsk) String() string {return "Slip slip knit"}

type ParsedKtog struct { // REDUCCION
	Count int
}
func (k *ParsedKtog) isStitch() {}
func (k *ParsedKtog) String() string {return "Knit "+strconv.Itoa(k.Count) + " together"}

type ParsedYo struct {}
func (y *ParsedYo) isStitch() {}
func (y *ParsedYo) String() string {return "Yarn over"}

type ParsedRepeat interface { 
	Expr
	isParsedRepeat()
}

type ParsedRepeatExact struct {
	Content Expr
	Count int
}
func (r *ParsedRepeatExact) String() string {
	return "Rep " + strconv.Itoa(r.Count) + "(" + r.Content.String() +")"
}
func (r *ParsedRepeatExact) isParsedRepeat() {}


type ParsedRepeatNeg struct {
	Content Expr
	Count int
}
func (r *ParsedRepeatNeg) String() string {
	return "Rep until " + strconv.Itoa(r.Count) + "(" + r.Content.String() +")"
}
func (r *ParsedRepeatNeg) isParsedRepeat() {}

type ParsedRepeatBlock struct {
	Content []*ParsedRow
	Count int
}

func (r *ParsedRepeatBlock) String() string {
	var exprs []string
	for _, row  := range r.Content {
		exprs = append(exprs, row.String())
	}
	return "Repeat "+strconv.Itoa(r.Count)+" times: \n  " + strings.Join(exprs, "\n  ")
}

type ParsedAction interface{
	Expr
	IsParsedAction()
}
type PlaceMarker struct {
	Name string
}
func (p *PlaceMarker) IsParsedAction() {}
func (p *PlaceMarker) String() string{ return "Place Marker "+p.Name }

type RemoveMarker struct {
	Name string
}
func (r *RemoveMarker) IsParsedAction() {}
func (r *RemoveMarker) String() string{ return "Remove Marker "+r.Name }

type ParsedRow struct {
	Content []Expr
}
func (r *ParsedRow) String() string {
	var exprs []string
	for _, expr  := range r.Content {
		exprs = append(exprs, expr.String())
	}
	return strings.Join(exprs, ", ")

}


type Section struct {
	Name string
	Content []Node
}

type Parser struct {
	l *Lexer  
	buf struct {
		pos Position
		n int		// 0 si no hay guardado, 1 si hay. i{}
		tok Token	//lst read token
		lit string	//last read literal
	}
}

// NewParser constructor
func NewParser(f io.Reader) *Parser{
	return &Parser {l: NewLexer(f)}
}

func (p *Parser) scan() (Position, Token, string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.pos, p.buf.tok, p.buf.lit
	}

	// Devuelve el siguiente toquen
	pos, tok, lit := p.l.Lex()
	p.buf.pos = pos
	p.buf.tok = tok 
	p.buf.lit = lit

	return pos, tok, lit
}

func (p *Parser) unscan()  {
	if p.buf.n != 0 {
		panic("unscan called twice without scan")
	}
	p.buf.n = 1
}

func (p *Parser) parseKtog() (*ParsedKtog, error){
	pos, tok, lit := p.scan()
	if tok != KTOG {
		return &ParsedKtog{}, fmt.Errorf("Expected KTOG, received %q in %q", tok, pos)
	}
	i, err  := strconv.Atoi(lit)
	if err!= nil  {
		return nil, fmt.Errorf("invalid epeition count: %q", lit)
	}
	return &ParsedKtog{Count: i}, nil
}

func (p *Parser) parseExpr() (Expr, error) {
	pos, tok, _ := p.scan()
	switch {
	case tok.isStitch():
		p.unscan()
		st, err := p.parseStitch()
		if err != nil {
			return nil, err
		}
		_, nextTok, _ := p.scan()
		if nextTok == REP {
			return p.parseParsedRepeat(st)
		}
		p.unscan()
		return st, nil
	case tok == PLACEMARKER:
		p.unscan()
		st, err := p.parsePlaceMarker()
		if err != nil {
			return nil, err
		}
		return st, nil
	case tok == REMOVEMARKER:
		p.unscan()
		st, err := p.parseRemoveMarker()
		if err != nil {
			return nil, err
		}
		return st, nil
	case tok == PAROPEN:
		p.unscan()
		group, err := p.parseGroup()
		if err != nil {
			return nil, err
		}
		_, newTok, _ := p.scan()
		if newTok == REP {
			return p.parseParsedRepeat(group)
		}
		p.unscan()
		return group, nil

	default:
		return nil, fmt.Errorf("unexpected token %q at %v", tok, pos)
	}
}


func (p *Parser) parseStitch() (ParsedStitch, error){
	pos, tok, _ := p.scan()
	if !tok.isStitch() {
		return nil, fmt.Errorf("extected stitch, got %q in %q", tok, pos)
	}
	switch tok {
	case KNIT:
		return &ParsedKnit{}, nil
	case PURL:
		return &ParsedPurl{}, nil
	case SSK:
		return &ParsedSsk{}, nil
	case YO:
		return &ParsedYo{}, nil
	case KTOG:
		p.unscan()
		return p.parseKtog()
	default:
		return nil, fmt.Errorf("extected stitch, got %q in %q", tok, pos)
	}
}


func (p *Parser) parseGroup() (Expr, error){
	var exprs []Expr
	pos, tok, _ := p.scan()
    if tok != PAROPEN {
        return nil, fmt.Errorf("expected '(', got %q at %v", tok, pos)
    }
	for {
		_, tok, _ := p.scan()
		if tok == PARCLOSE {
			break
		}
		p.unscan()

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}
	return &Group{Content: exprs}, nil
}

func (p *Parser) parseParsedRepeat(content Expr) (Expr, error) {
	pos, tok, lit := p.scan()
	switch tok {
	case INT:
		i, err  := strconv.Atoi(lit)
		if err!= nil  {
			return nil, fmt.Errorf("invalid epeition count: %q", lit)
		}
		return &ParsedRepeatExact{Content: content, Count: i}, nil
	case NEG:
		pos, tok, lit = p.scan()
		if tok != INT{
			return nil, fmt.Errorf("Expected an integer, received %q in pos %q", tok, pos)
		}
		i, err  := strconv.Atoi(lit)
		if err!= nil  {
			return nil, fmt.Errorf("invalid epeition count: %q", lit)
		}
		return &ParsedRepeatNeg{Content: content, Count: i}, nil
	default:
		return nil, fmt.Errorf("After * expected an integer or a -, received %q in pos %q", tok, pos)
	}
}


func (p *Parser) parseRow() (*ParsedRow, error){
	var exprs []Expr
	for {
		_, tok, _ := p.scan()
		if tok == SEMICOLON {
			break
		}
		if tok == EOF || tok == BRCLOSE {
			return &ParsedRow{}, fmt.Errorf("unexpected end of sequence, expected ';'")
		}
		p.unscan() // porque parseExpr va a escanear el siguiente token
		expr, err := p.parseExpr()
		if err != nil {
			return &ParsedRow{}, err
		}
		exprs = append(exprs, expr)
		if len(exprs) == 0 {
			panic("empty row")
		}
	}
	return &ParsedRow{Content: exprs}, nil
}

func (p *Parser) parseParsedRepeatBlock() (*ParsedRepeatBlock, error){
	var rows []*ParsedRow
	pos, tok, lit := p.scan()
	if tok != REPBLOCK {
		return nil, fmt.Errorf("expected 'repeat' got %q at %v", tok, pos)
	}
	pos, tok, lit = p.scan()
	if tok != INT {
		return nil, fmt.Errorf("expected int got %q at %v", tok, pos)
	}

	i, err  := strconv.Atoi(lit)
	if err!= nil  {
		return nil, fmt.Errorf("invalid epeition count: %q", i)
	}
	pos, tok, _ = p.scan()
	if tok != BROPEN {
		return nil, fmt.Errorf("expected '{' got %q at %v", tok, pos)
	}
	for {
		_, tok, _ := p.scan()
		if tok == BRCLOSE {
			break
		}
		if tok == EOF {
			return nil, fmt.Errorf("unexpected EOF inside repeatblock")
		}
		p.unscan()
		row, err := p.parseRow()
		if err != nil {
			return nil, fmt.Errorf("error parsing row %q at %v: %v", pos, err)
		}
		rows =  append(rows, row)
	}
	return &ParsedRepeatBlock{Content:rows, Count: i}, nil 
}


func (p *Parser) parsePlaceMarker() (*PlaceMarker, error){
	pos, tok, lit := p.scan()
	if tok != PLACEMARKER {
		return nil, fmt.Errorf("expected 'placemarker', got %q at %v", tok, pos)
	}
	return &PlaceMarker{Name: lit}, nil
}

func (p *Parser) parseRemoveMarker() (*RemoveMarker, error){
	pos, tok, lit := p.scan()
	if tok != REMOVEMARKER {
		return nil, fmt.Errorf("expected 'placemarker', got %q at %v", tok, pos)
	}
	return &RemoveMarker{Name: lit}, nil
}


func (p *Parser) parseSection() (*Section, error) {
	// Validar que la sección empieza con la palabra clave "section"
	pos, tok, _ := p.scan()
	if tok != SECTION {
		return nil, fmt.Errorf("expected 'section', got %q at %v", tok, pos)
	}

	pos, tok, lit := p.scan()
	if tok != IDENT {
		return nil, fmt.Errorf("expected section name (IDENT), got %q at %v", tok, pos)
	}
	section := &Section{Name: lit}

	pos, tok, _ = p.scan()
	if tok != BROPEN {
		return nil, fmt.Errorf("expected '{', got %q at %v", tok, pos)
	}

	for {
		pos, tok, _ := p.scan()
		if tok == BRCLOSE {
			break
		}
		if tok == EOF {
			return nil, fmt.Errorf("unexpected EOF inside section %q", section.Name)
		}

		if tok == REPBLOCK {
			p.unscan()
			repblock, err := p.parseParsedRepeatBlock()
			if err != nil {
				return nil, fmt.Errorf("error parsing row in section %q at %v: %v", section.Name, pos, err)
			}
			section.Content = append(section.Content, repblock)
		}else {
			p.unscan()
			row, err := p.parseRow()
			if err != nil {
				return nil, fmt.Errorf("error parsing row in section %q at %v: %v", section.Name, pos, err)
			}
			section.Content = append(section.Content, row)
		}
	}

	return section, nil
}

func (p *Parser) ParsePattern() ([]*Section, error) {
	var sections []*Section

	for {
		pos, tok, _ := p.scan()
		if tok == EOF {
			break
		}

		p.unscan()
		section, err := p.parseSection()
		if err != nil {
			return nil, fmt.Errorf("error parsing section at %v: %v", pos, err)
		}

		sections = append(sections, section)
	}

	return sections, nil
}

func isParsedStitch(t Node) (ParsedStitch, bool) {
	st, ok := t.(ParsedStitch)
	return st, ok
}

func isParsedAction(t Node) (ParsedAction, bool) {
	act, ok := t.(ParsedAction)
	return act, ok
}


func isParsedRepeat(t Node) (ParsedRepeat, bool){
	rep, ok := t.(ParsedRepeat)
	return rep, ok
}
