package tdx

import (
  "fmt"
  "log"
  "os"

  "github.com/alecthomas/participle/v2"
  "github.com/alecthomas/participle/v2/lexer"
  "github.com/alecthomas/repr"
  "github.com/urfave/cli/v2"
)

type TdxHandler struct{}

func NewTdxCommand() *cli.Command {
  var h TdxHandler
  return &cli.Command{
    Name:  "tdx",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TdxHandler{}
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

func (h *TdxHandler) test() error {
  log.Println("tdx test...")
  lex := lexer.MustSimple([]lexer.SimpleRule{
    {"String", `'(\\'|[^'])*'|"(\\"|[^"])*"`},
    {"Number", `([0-9]+\.)?[0-9]+`},
    {"Ident", fmt.Sprintf(`[a-zA-Z%v-%v_%v.']+\w*`, "\u4e00", "\u9fa5", "％")},
    {"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
    {"Operator", `!=|<=|>=|[-+*/%()=<>]`},
    {"whitespace", `[ \t\n\r]+`},
  })
  parser := participle.MustBuild[Program](
    participle.Lexer(lex),
    participle.CaseInsensitive("Ident"),
    participle.Unquote("String"),
    participle.UseLookahead(2),
  )
  file, err := os.Open("example.tdx")
  if err != nil {
    return err
  }
  program, err := parser.Parse("", file)
  if err != nil {
    return err
  }
  repr.Println(program, repr.Indent("  "), repr.OmitEmpty(true))
  return nil
}
