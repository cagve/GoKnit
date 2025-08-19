package main

import (
	"strconv"
	"fmt"
)

type CompilePosition struct {
	RowPos int	//row
	ColPos int  //st
}

type CompileUnit interface {
	String() string
}

// Interface Stitches: Knit, purl, ssk, ktog, yo.
type Stitch interface {
	CompileUnit
	weight() int 
	advance() int
}

type Knit struct {}
func (k *Knit) isStitch() {}
func (k *Knit) String() string {return "Knit"}
func (k *Knit) weight() int {return 1}
func (k *Knit) advance() int {return 1}

type Purl struct {}
func (p *Purl) isStitch() {}
func (p *Purl) String() string {return "Purl"}
func (p *Purl) weight() int {return 1}
func (p *Purl) advance() int {return 1}

type Ssk struct {} // REDUCCION
func (s *Ssk) isStitch() {}
func (s *Ssk) String() string {return "Slip slip knit"}
func (s *Ssk) weight() int {return 1}
func (s *Ssk) advance() int {return 2}

type Ktog struct { // REDUCCION
	Count int
}
func (k *Ktog) isStitch() {}
func (k *Ktog) String() string {return "Knit "+strconv.Itoa(k.Count) + " together"}
func (k *Ktog) weight() int {return 1}
func (k *Ktog) advance() int {return 2}

type Yo struct {}
func (y *Yo) isStitch() {}
func (y *Yo) String() string {return "Yarn over"}
func (y *Yo) weight() int {return 1}
func (y *Yo) advance() int {return 0}

type Marker struct {
	Name string
	Pos CompilePosition
}

type Row struct {
	Stitches []Stitch
	Number int
}
func (r *Row) weight() int {
	w := 0
	for _, st  := range r.Stitches {
		w += st.weight()	
	}
	return w
}

type Compiler struct {
	LastRow  *Row
	Markers  []Marker
	Rows     []*Row
	Errors   []error
	buf      struct {
		Pos        CompilePosition
		CurrentRow *Row
	}
}

func NewCompiler() *Compiler {
	return &Compiler{
		LastRow: nil,
		Markers: make([]Marker, 0),
		Rows:    make([]*Row, 0),
		Errors:  make([]error, 0),
		buf: struct {
			Pos        CompilePosition
			CurrentRow *Row
		}{
			Pos:        CompilePosition{RowPos: 0, ColPos: 1},
			CurrentRow: nil,
		},
	}
}

func (c *Compiler) startNewRow() {
	// Empieza en 0, ppor lo que cuando creo mi primera row, tengo que sumarle uno.
	c.buf.Pos.RowPos ++
	newRow := &Row {
		Stitches: make([]Stitch, 0),
		Number: c.buf.Pos.RowPos,
	}
	if c.buf.CurrentRow != nil {
		c.Rows = append (c.Rows, c.buf.CurrentRow)
		c.LastRow = c.buf.CurrentRow
	}
	c.buf.CurrentRow = newRow
	c.buf.Pos.ColPos = 1
}

func (c *Compiler) addStitch(st Stitch) error {
	if c.buf.CurrentRow == nil {
		return fmt.Errorf("No active row to add sts")
	}

	newPos := c.buf.Pos.ColPos + st.advance()
	// Primero chequeamos si se puede.	
	if c.LastRow != nil {
		lastWeight := c.LastRow.weight()
		if newPos > lastWeight {
			return fmt.Errorf("Exceded number of sts (max %d, got %d)", lastWeight, newPos) //TODO: manejo de errores
		}
	}

	//Updateamos todo.
	c.buf.CurrentRow.Stitches = append(c.buf.CurrentRow.Stitches, st)
	c.buf.Pos.ColPos = newPos
	return nil
}

func (c *Compiler) addMarker(mk Marker) {
	c.Markers = append(c.Markers, mk)
}

func (c *Compiler) removeMarker(markerName string) error{
	for i := range c.Markers {
		if c.Markers[i].Name == markerName {
			c.Markers = append(c.Markers[:i], c.Markers[i+1:]...)
		}
	}
	return fmt.Errorf("Marker %d not found", markerName)
}

func (c *Compiler) compileRow(parsedRow *ParsedRow) error {
	c.startNewRow()
	for _, parsedExpr := range(parsedRow.Content){
		if stitch, ok := isParsedStitch(parsedExpr); ok {
			c.compileStitch(stitch)
		} else if act, ok := isParsedAction(parsedExpr); ok {
			c.compileParsedAction(act)
		} else if rep, ok := isParsedRepeat(parsedExpr); ok{
			c.compileParsedRepeat(rep)
		} else {
			return fmt.Errorf("Unknown expresion %d", parsedExpr)
		}
	}
	return nil
}

func (c *Compiler) compileExpr(parsedExpr Node) error {
	if stitch, ok := isParsedStitch(parsedExpr); ok {
		c.compileStitch(stitch)
	} else if act, ok := isParsedAction(parsedExpr); ok {
		c.compileParsedAction(act)
	} else if rep, ok := isParsedRepeat(parsedExpr); ok{
		c.compileParsedRepeat(rep)
	} else {
		return fmt.Errorf("Unknown expresion %d", parsedExpr)
	}
	return nil
}


func (c *Compiler) compileStitch(stitch ParsedStitch) (Stitch, error) {
	var st Stitch
	switch s := stitch.(type) {
	case *ParsedKnit:
		st = &Knit{}
	case *ParsedPurl:
		st = &Purl{}
	case *ParsedSsk:
		st = &Ssk{}
	case *ParsedKtog:
		st = &Ktog{Count: s.Count}
	case *ParsedYo:
		st = &Yo{}
	default:
		return nil, fmt.Errorf("Unknown stitch type: %T", stitch)
	}

	c.addStitch(st)
	return st, nil
}

func (c *Compiler) compileParsedAction(act ParsedAction) (error) {
	switch a := act.(type) {
	case *PlaceMarker:
		marker := Marker{Name: a.Name, Pos: c.buf.Pos} 
		c.addMarker(marker)
	case *RemoveMarker:
		c.removeMarker(a.Name)
	default:
		return fmt.Errorf("Unknown action type: %T", act)
	}
	return fmt.Errorf("Unknown action type: %T", act)
}

func (c *Compiler) compileParsedRepeat(rep ParsedRepeat) (error) {
	switch r := rep.(type) {
	case *ParsedRepeatNeg:
		n := r.Count
		expr := r.Content
		reps := (c.LastRow.weight() - c.buf.Pos.CurrentCol) / expr.advance()
	case *ParsedRepeatExact:
		c.removeMarker(a.Name)
	default:
		return fmt.Errorf("Unknown action type: %T", act)
	}
	return fmt.Errorf("Unknown action type: %T", act)
}


