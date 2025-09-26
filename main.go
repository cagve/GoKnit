package main

import ( 
	"fmt"
	"strings"
	"os"
	"io"
	"github.com/rivo/tview"
)


func test_lexer(file io.Reader) {
	fmt.Println("LEXING")
	lexer := NewLexer(file)
	for {
		pos, tok, lit := lexer.Lex()
		fmt.Printf("%d:%d\t%s\t%s\n", pos.line, pos.column, tok, lit)
		if tok == EOF {
			break
		}
	}
}

func test_parser(file io.Reader){
	fmt.Println("PARSING")
	parser := NewParser(file)
	pattern, err := parser.ParsePattern()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return 
	}
	for _, section := range pattern {
		fmt.Printf("Section: %s\n", section.Name)
		for _, node := range section.Content{
			fmt.Printf("➡ Tipo: %T, Nodo: %S\n", node, node)
		}
	}
}



func test_compile(file io.Reader) Compiler {
	parser := NewParser(file)
	pattern, err := parser.ParsePattern()

	if err != nil {
		fmt.Print(err)
	}

	c := NewCompiler()
	for _, section := range pattern {
		for _, node := range section.Content {
			if row, ok := node.(*ParsedRow); ok {
				if err := c.compileRow(row); err != nil {
					fmt.Printf("[ERROR] compiling row fila %d: %s\n >> desc: %v\n", c.Pos.RowPos, row, err)
					return Compiler{}
				}
			} else if repeatBlock, ok := node.(*ParsedRepeatBlock); ok  {
				c.compileRepeatBlock(repeatBlock)
			} 
		}
	}
	return *c
}

func launch_parser(){
	file, err := os.Open("./input.test")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	test_lexer(file)
	//Abre el archivo de nuevo.
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	// test_parser(file)
	c := test_compile(file)
	// err = compileToImg(c, "output.jpg")
	fmt.Println(c)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	app 	:= tview.NewApplication()
	form 	:= tview.NewForm().
				AddTextView("Título", "Patrón de Diseño Singleton", 0, 1, false, false).
				AddTextView("Autor", "Gamma, Helm, Johnson, Vlissides", 0, 1, false, false).
				AddTextView("Fecha inicio", "1994", 0, 1, false, false).
				AddTextView("Dificultad", "Media", 0, 1, false, false).
				AddTextView("Categoría", "Creacional", 0, 1, false, false)
	form.SetBorder(true).SetTitle("Información del Patrón")


	// Crear List en lugar de Table
	list := tview.NewList().ShowSecondaryText(false)

	// Datos de las filas del patrón
	filas := []struct {
		numero string
		puntos []string
	}{
		{"1", []string{"CO3"}},
		{"2", []string{"P", "P", "P"}},
		{"3", []string{"K", "K", "K"}},
	}

	// Añadir cada fila como item de la lista
	for _, fila := range filas {
		instrucciones := strings.Join(fila.puntos, ", ")
		list.AddItem(fmt.Sprintf("Fila %s: %s", fila.numero, instrucciones), "", 0, nil)
	}

	list.SetBorder(true).SetTitle("Filas del Patrón (↑↓ para navegar)")

	// Hacerla navegable y con feedback visual
	list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		// Aquí puedes añadir acción al seleccionar una fila
		app.SetFocus(list)
	})

	flex	:= tview.NewFlex().
			AddItem(form, 0, 30, false).
			AddItem(list, 0, 70, true)

	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}
}

