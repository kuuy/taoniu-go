package basic

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

type BasicHandler struct{}

func NewBasicCommand() *cli.Command {
	var h BasicHandler
	return &cli.Command{
		Name:  "basic",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = BasicHandler{}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "test",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.test(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *BasicHandler) test() error {
	log.Println("participles basic")
	basicLexer := lexer.MustSimple([]lexer.SimpleRule{
		{"Comment", `(?i)rem[^\n]*`},
		{"String", `"(\\"|[^"])*"`},
		{"Number", `[-+]?(\d*\.)?\d+`},
		{"Ident", `[a-zA-Z_]\w*`},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
		{"EOL", `[\n\r]+`},
		{"whitespace", `[ \t]+`},
	})
	parser := participle.MustBuild[Program](
		participle.Lexer(basicLexer),
		participle.CaseInsensitive("Ident"),
		participle.Unquote("String"),
		participle.UseLookahead(2),
	)
	file, err := os.Open("example.bas")
	if err != nil {
		return err
	}
	program, err := parser.Parse("", file)
	if err != nil {
		return err
	}
	program.init()
	//funcs := map[string]Function{
	//	"ADD": func(args ...interface{}) (interface{}, error) {
	//		return args[0].(float64) + args[1].(float64), nil
	//	},
	//}
	//err = program.Evaluate(os.Stdin, os.Stdout, funcs)
	//if err != nil {
	//	return err
	//}
	repr.Println(program, repr.Indent("  "), repr.OmitEmpty(true))
	return nil
}
