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
func (r *Row) String() string{
	stitches := []string{}
	for _, st := range r.Stitches {
		stitches = append(stitches, st.String())
	}
	return fmt.Sprintf("Row %d: [%s]", r.Number, fmt.Sprintf("%s", stitches))
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
		c.compileExpr(parsedExpr)
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

func (c *Compiler) compileExpr(expr Expr) ([]Stitch, error) {
    var stitches []Stitch
    // Si el Expr es un grupo, iteramos sobre sus contenidos
    if group, ok := expr.(*Group); ok {
        for _, subExpr := range group.Content {
            subStitches, err := c.compileExpr(subExpr)
            if err != nil {
                return nil, err
            }
            stitches = append(stitches, subStitches...)
        }
    } else if parsedStitch, ok := expr.(ParsedStitch); ok {
        st, err := c.compileStitch(parsedStitch)
        if err != nil {
            return nil, err
        }
        stitches = append(stitches, st)
    } else if repeatExpr, ok := expr.(ParsedRepeat); ok {
        switch r := repeatExpr.(type) {
        case *ParsedRepeatExact:
            for i := 0; i < r.Count; i++ {
                subStitches, err := c.compileExpr(r.Content)
                if err != nil {
                    return nil, err
                }
                stitches = append(stitches, subStitches...)
            }
        case *ParsedRepeatNeg:
			return nil, fmt.Errorf("Not implemented yet.")
        default:
            return nil, fmt.Errorf("unknown repeat type: %T", repeatExpr)
        }
	} else if actionExpr, ok := expr.(ParsedAction); ok {
        c.compileParsedAction(actionExpr)
    } else {
        return nil, fmt.Errorf("unsupported expression type: %T", expr)
    }

    return stitches, nil
}


func (c *Compiler) printAllRows() {
	for _, row := range c.Rows {
		fmt.Printf("Row %d: %v\n", row.Number, row)
	}
}
