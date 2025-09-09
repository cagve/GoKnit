package main

import (
	"fmt"
	"strconv"
	"strings"
)

type CompilePosition struct {
	RowPos int // row
	ColPos int // st
}

type CompileUnit interface {
	String() string
}

type Expr interface {
	CompileUnit
	isExpr()
}

type Stitch interface {
	Expr
	weight() int
	advance() int
}

// Implementación de isExpr() para todos los stitches
func (k *Knit) isExpr()         {}
func (p *Purl) isExpr()         {}
func (s *Ssk) isExpr()          {}
func (k *Ktog) isExpr()         {}
func (y *Yo) isExpr()           {}
func (c *Co) isExpr()           {}
func (c *Bo) isExpr()           {}
func (c *CableBkd) isExpr()     {}
func (c *CableFwd) isExpr()     {}
func (c *PurlCableBkd) isExpr() {}
func (c *PurlCableFwd) isExpr() {}
func (g *Group) isExpr()        {}

type Knit struct{}

func (k *Knit) String() string { return "Knit" }
func (k *Knit) weight() int    { return 1 }
func (k *Knit) advance() int   { return 1 }

type Purl struct{}

func (p *Purl) String() string { return "Purl" }
func (p *Purl) weight() int    { return 1 }
func (p *Purl) advance() int   { return 1 }

type Ssk struct{}             // REDUCCION
func (s *Ssk) String() string { return "Slip slip knit" }
func (s *Ssk) weight() int    { return 1 }
func (s *Ssk) advance() int   { return 2 }

type Ktog struct { // REDUCCION
	Count int
}

func (k *Ktog) String() string { return "K" + strconv.Itoa(k.Count) + "TOG" }
func (k *Ktog) weight() int    { return 1 }
func (k *Ktog) advance() int   { return k.Count }

type Co struct {
	Count int
}

func (c *Co) String() string { return "Cast on" + strconv.Itoa(c.Count) }
func (c *Co) weight() int    { return c.Count }
func (c *Co) advance() int   { return 0 }

type Bo struct {
	Count int
}

func (b *Bo) String() string { return "Bind off" + strconv.Itoa(b.Count) }
func (b *Bo) weight() int    { return 0 }
func (b *Bo) advance() int   { return b.Count }

type Yo struct{}

func (y *Yo) String() string { return "Yo" }
func (y *Yo) weight() int    { return 1 }
func (y *Yo) advance() int   { return 0 }

// CABLES
type CableFwd struct {
    FrontCount     int
    BackCount int
}

func (c *CableFwd) String() string { return fmt.Sprintf("C%d/%dF", c.FrontCount, c.BackCount) }
func (c *CableFwd) weight() int  { return c.FrontCount + c.BackCount }
func (c *CableFwd) advance() int { return c.weight() }

type CableBkd struct {
    FrontCount     int
    BackCount int
}

func (c *CableBkd) String() string { return fmt.Sprintf("C%d/%dB", c.FrontCount, c.BackCount) }
func (c *CableBkd) weight() int    { return c.FrontCount + c.BackCount }
func (c *CableBkd) advance() int   { return c.weight() }

type PurlCableFwd struct {
    FrontCount     int
    BackCount int
}

func (c *PurlCableFwd) String() string { return fmt.Sprintf("P%d/%dF", c.FrontCount, c.BackCount) }
func (c *PurlCableFwd) weight() int    { return c.FrontCount + c.BackCount }
func (c *PurlCableFwd) advance() int   { return c.weight() }

type PurlCableBkd struct {
    FrontCount     int
    BackCount int
}

func (c *PurlCableBkd) String() string { return fmt.Sprintf("P%d/%dB", c.FrontCount, c.BackCount) }
func (c *PurlCableBkd) weight() int  { return c.FrontCount + c.BackCount }
func (c *PurlCableBkd) advance() int { return c.weight() }

// Grupo - análogo a Group del parser
type Group struct {
	Content []Expr
}

func (g *Group) String() string {
	var exprs []string
	for _, expr := range g.Content {
		exprs = append(exprs, expr.String())
	}
	return "(" + strings.Join(exprs, ", ") + ")"
}

type RepeatBlock interface {
	Expr
	isRepeatBlock()
}

type RepeatBlockExact struct {
	Content []Row
	Count   int
}

func (r RepeatBlockExact) isExpr()        {}
func (r RepeatBlockExact) isRepeatBlock() {}
func (r RepeatBlockExact) String() string { return "not implementted" }

type Repeat interface {
	Expr
	isRepeat()
}

type RepeatExact struct {
	Content Expr
	Count   int
}

func (r *RepeatExact) isExpr()   {}
func (r *RepeatExact) isRepeat() {}
func (r *RepeatExact) String() string {
	return "Rep " + strconv.Itoa(r.Count) + "(" + r.Content.String() + ")"
}

type RepeatNeg struct {
	Content Expr
	Count   int
}

func (r *RepeatNeg) isExpr()   {}
func (r *RepeatNeg) isRepeat() {}
func (r *RepeatNeg) String() string {
	return "Rep until " + strconv.Itoa(r.Count) + "(" + r.Content.String() + ")"
}

// Funciones helper para verificar tipos
func isStitch(expr Expr) (Stitch, bool) {
	st, ok := expr.(Stitch)
	return st, ok
}

func isRepeat(expr Expr) (Repeat, bool) {
	rep, ok := expr.(Repeat)
	return rep, ok
}

func isGroup(expr Expr) (*Group, bool) {
	group, ok := expr.(*Group)
	return group, ok
}

type Row struct {
	Stitches []Stitch
	Number   int
}

func (r *Row) weight() int {
	w := 0
	for _, st := range r.Stitches {
		w += st.weight()
	}
	return w
}
func (r *Row) advance() int {
	w := 0
	for _, st := range r.Stitches {
		w += st.advance()
	}
	return w
}

func (r *Row) String() string {
	stitches := []string{}
	for _, st := range r.Stitches {
		stitches = append(stitches, st.String())
	}
	return fmt.Sprintf("Row %d: [%s]", r.Number, strings.Join(stitches, ", "))
}

type Compiler struct {
	LastRow    *Row
	Rows       []*Row
	Errors     []error
	Pos        CompilePosition
	CurrentRow *Row
}

func NewCompiler() *Compiler {
	return &Compiler{
		LastRow:    nil,
		Rows:       make([]*Row, 0),
		Errors:     make([]error, 0),
		Pos:        CompilePosition{RowPos: 0, ColPos: 1},
		CurrentRow: nil,
	}
}

func (c *Compiler) startNewRow() {
	c.Pos.RowPos++
	newRow := &Row{
		Stitches: make([]Stitch, 0),
		Number:   c.Pos.RowPos,
	}
	if c.CurrentRow != nil {
		c.Rows = append(c.Rows, c.CurrentRow)
		c.LastRow = c.CurrentRow
	}
	c.CurrentRow = newRow
	c.Pos.ColPos = 1
}

func (c *Compiler) addStitch(st Stitch) error {
	if c.CurrentRow == nil {
		return fmt.Errorf("No active row to add sts")
	}

	currentPos := c.Pos.ColPos
	if c.LastRow != nil {
		lastWeight := c.LastRow.weight()
		if currentPos > lastWeight {
			return fmt.Errorf("Exceded number of sts (max %d, got %d) in Row %d", lastWeight, currentPos, c.Pos.RowPos)
		}
	}

	newPos := c.Pos.ColPos + st.advance()
	c.CurrentRow.Stitches = append(c.CurrentRow.Stitches, st)
	c.Pos.ColPos = newPos
	return nil
}

func (c *Compiler) expandGroup(compiledExpr Expr) ([]Stitch, error) {
	var sts []Stitch
	switch expr := compiledExpr.(type) {
	case *Group:
		for _, expr := range expr.Content {
			expanded, _ := c.expandExpr(expr)
			sts = append(sts, expanded...)
		}
	default:
		return nil, fmt.Errorf("Expected group expression, received: %T", expr)
	}
	return sts, nil
}

func (c *Compiler) expandRepeat(compiledExpr Expr) ([]Stitch, error) {
	var sts []Stitch
	switch expr := compiledExpr.(type) {
	case *RepeatExact:
		times := expr.Count
		if times == 0 {
			if c.LastRow != nil {
				remaining := c.LastRow.weight() - (c.Pos.ColPos - 1)
				perRepeat := c.exprAdvance(expr.Content)
				if perRepeat == 0 {
					return nil, fmt.Errorf("repeat content has zero advance, cannot calculate repetitions")
				}
				times = remaining / perRepeat
			} else {
				return nil, fmt.Errorf("cannot infer repeat count: no previous row")
			}
		}
		for range times {
			expanded, err := c.expandExpr(expr.Content)
			if err != nil {
				return nil, err
			}
			sts = append(sts, expanded...)

		}
	case *RepeatNeg:
		if c.LastRow == nil {
			return nil, fmt.Errorf("cannot expand RepeatNeg: no previous row to infer remaining stitches")
		}

		total := c.LastRow.weight()
		remaining := total - (c.Pos.ColPos - 1)
		perRepeat := c.exprAdvance(expr.Content)

		if perRepeat == 0 {
			return nil, fmt.Errorf("repeat content has zero advance, cannot calculate repetitions")
		}

		times := max((remaining-expr.Count)/perRepeat, 0)
		for range times {
			expanded, err := c.expandExpr(expr.Content)
			if err != nil {
				return nil, err
			}
			sts = append(sts, expanded...)
		}
	default:
		return nil, fmt.Errorf("Expected repeat expression, received: %T", expr)
	}
	return sts, nil
}

func (c *Compiler) expandExpr(compiledExpr Expr) ([]Stitch, error) {
	var sts []Stitch
	switch expr := compiledExpr.(type) {
	case Stitch:
		sts = append(sts, expr)
		return sts, nil
	case *Group:
		sts, err := c.expandGroup(expr)
		if err != nil {
			return nil, err
		}
		return sts, nil
	case Repeat:
		sts, err := c.expandRepeat(expr)
		if err != nil {
			return nil, err
		}
		return sts, nil
	default:
		return nil, fmt.Errorf("unsupported parsed expression type: %T", compiledExpr)
	}
}

func (c *Compiler) compileRow(parsedRow *ParsedRow) error {
	c.startNewRow()
	var sts []Stitch
	for _, parsedExpr := range parsedRow.Content {
		e, err := c.compileExpr(parsedExpr)
		if err != nil {
			return err
		}
		expandedSts, err := c.expandExpr(e)
		if err != nil {
			return err
		}
		sts = append(sts, expandedSts...)
		advance := 0
		for _, st := range expandedSts {
			advance += st.advance()
		}
		c.Pos.ColPos += advance
	}
	c.CurrentRow.Stitches = sts
	fmt.Printf(">%s\n", sts)
	if c.LastRow != nil && c.LastRow.weight() != c.CurrentRow.advance() {
		return fmt.Errorf("Unmatch number of stitches. Expected: %d, Received: %d",
			c.LastRow.weight(), c.CurrentRow.advance())
	}

	return nil
}

func (c *Compiler) compileExpr(parsedExpr ParsedExpr) (Expr, error) {
	switch expr := parsedExpr.(type) {
	case ParsedStitch:
		st, err := c.compileStitch(expr)
		if err != nil {
			return nil, err
		}
		return st, nil
	case *ParsedGroup:
		group, err := c.compileGroup(expr)
		if err != nil {
			return nil, err
		}
		return group, nil
	case ParsedRepeat:
		repeat, err := c.compileRepeat(expr)
		if err != nil {
			return nil, err
		}
		return repeat, err
	default:
		return nil, fmt.Errorf("unsupported parsed expression type: %T", parsedExpr)
	}
}

func (c *Compiler) compileStitch(parsedStitch ParsedStitch) (Stitch, error) {
	switch s := parsedStitch.(type) {
	case *ParsedKnit:
		return &Knit{}, nil
	case *ParsedPurl:
		return &Purl{}, nil
	case *ParsedSsk:
		return &Ssk{}, nil
	case *ParsedKtog:
		return &Ktog{Count: s.Count}, nil
	case *ParsedYo:
		return &Yo{}, nil
	case *ParsedCo:
		return &Co{Count: s.Count}, nil
	case *ParsedBo:
		return &Bo{Count: s.Count}, nil
	case *ParsedCableBkd:
		return &CableBkd{BackCount: s.BackCount, FrontCount: s.FrontCount}, nil
	case *ParsedCableFwd:
		return &CableFwd{BackCount: s.BackCount, FrontCount: s.FrontCount}, nil
	case *ParsedPurlCableBkd:
		return &PurlCableBkd{BackCount: s.BackCount, FrontCount: s.FrontCount}, nil
	case *ParsedPurlCableFwd:
		return &PurlCableFwd{BackCount: s.BackCount, FrontCount: s.FrontCount}, nil
	default:
		return nil, fmt.Errorf("Unknown stitch type: %T", parsedStitch)
	}
}

func (c *Compiler) compileGroup(parsedGroup *ParsedGroup) (*Group, error) {
	var exprs []Expr
	for _, subExpr := range parsedGroup.Content {
		compiledSubExprs, err := c.compileExpr(subExpr)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, compiledSubExprs)
	}
	return &Group{Content: exprs}, nil
}

func (c *Compiler) compileRepeatBlock(parsedRepeatBlock *ParsedRepeatBlock) error {
	for i := 0; i < parsedRepeatBlock.Count; i++ {
		for _, row := range parsedRepeatBlock.Content {
			c.compileRow(row)
		}
	}
	return nil
}

func (c *Compiler) compileRepeat(parsedRepeat ParsedRepeat) (Repeat, error) {
	switch r := parsedRepeat.(type) {
	case *ParsedRepeatExact:
		content, err := c.compileExpr(r.Content)
		if err != nil {
			return nil, err
		}
		return &RepeatExact{
			Content: content,
			Count:   r.Count,
		}, nil

	case *ParsedRepeatNeg:
		content, err := c.compileExpr(r.Content)
		if err != nil {
			return nil, err
		}
		return &RepeatNeg{
			Content: content,
			Count:   r.Count,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported parsed repeat type: %T", parsedRepeat)
	}
}

func (c *Compiler) exprWeight(compileExpr Expr) int {
	switch expr := compileExpr.(type) {
	case Stitch:
		return expr.weight()
	case *Group:
		w := 0
		for _, subExpr := range expr.Content {
			w += c.exprWeight(subExpr)
		}
		return w
	case *RepeatExact:
		return expr.Count * c.exprWeight(expr.Content)
	case *RepeatNeg:
		return 0
	}
	return 0
}

func (c *Compiler) exprAdvance(compileExpr Expr) int {
	switch expr := compileExpr.(type) {
	case Stitch:
		return expr.advance()
	case *Group:
		w := 0
		for _, subExpr := range expr.Content {
			w += c.exprAdvance(subExpr)
		}
		return w
	case *RepeatExact:
		if expr.Count == 0 {
			if c.LastRow != nil {
				remaining := c.LastRow.weight() - c.Pos.ColPos
				perRepeat := c.exprAdvance(expr.Content)

				if perRepeat == 0 {
					return 0
				}
				return (remaining / perRepeat) * perRepeat
			}
			return 0
		}
		return expr.Count * c.exprAdvance(expr.Content)
	case *RepeatNeg:
		return 0
	}
	return 0
}
