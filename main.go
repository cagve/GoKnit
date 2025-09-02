package main

import ( 
	"fmt"
	"os"
	"io"
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



func test_compile(file io.Reader) {
	parser := NewParser(file)
	pattern, _ := parser.ParsePattern()

	c := NewCompiler()
	for _, section := range pattern {
		for _, node := range section.Content {
			if row, ok := node.(*ParsedRow); ok {
				// Suponiendo que compileRow ahora retorna error
				if err := c.compileRow(row); err != nil {
					fmt.Printf("[ERROR] compiling row fila %d: %s\n >> desc: %v\n", c.Pos.RowPos, row, err)
					// Si quieres detener la ejecución al primer error:
					return
				}

				if c.LastRow != nil {
					// fmt.Printf(
					// 	"Current row token: %s; Current Compiled row: %s; Expected number of sts: %v; Current number of sts: %v\n",
					// 	row, c.CurrentRow, c.LastRow.weight(), c.CurrentRow.weight(),
					// )
				} else {
					// fmt.Printf(
					// 	"Current row token: %s; Current Compiled row: %s; Current number of sts: %v\n",
					// 	row, c.CurrentRow, c.CurrentRow.weight(),
					// )
				}
			}
		}
	}
}



func main() {
	file, err := os.Open("input.test")
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
	test_compile(file)
}
