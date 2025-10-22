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
	for _, row := range(c.Rows) {
		fmt.Println(row)
	}
}

func main() {
	// readSession("lib/example.json")
	app()
}

