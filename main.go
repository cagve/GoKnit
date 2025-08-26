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
				c.compileRow(row)
			}
		}
	}
	c.printAllRows()
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
