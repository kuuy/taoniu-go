package tdx

import (
  "github.com/alecthomas/participle/v2/lexer"
)

type Program struct {
  Pos       lexer.Position
  Statement []*Statement `@@*`
}

type Statement struct {
  Pos        lexer.Position
  Assignment *Assignment `@@ ";"`
}

type Operator string

type Func struct {
  Pos  lexer.Position
  Name string        `@Ident`
  Arg  []*Expression `"(" ( @@ ( "," @@ )* )? ")"`
}

type Value struct {
  Pos           lexer.Position
  Variable      string      `  @Ident`
  String        string      `| @String`
  Number        float64     `| @Number`
  Nagetive      float64     `| "-"? @Number`
  Subexpression *Expression `| "(" @@ ")"`
}

type Factor struct {
  Pos      lexer.Position
  Func     *Func  `( @@`
  Base     *Value `| @@ )`
  Exponent *Value `( "^" @@ )?`
}

type OpFactor struct {
  Pos      lexer.Position
  Operator Operator `@("*" | "/" | "<" | ">" | "!" | "=" | "AND")`
  Equal    Operator `@( "="? )`
  Factor   *Factor  `@@`
}

type Term struct {
  Pos   lexer.Position
  Left  *Factor     `@@`
  Right []*OpFactor `@@*`
}

type OpTerm struct {
  Pos      lexer.Position
  Operator Operator `@("+" | "-" | "OR")`
  Term     *Term    `@@`
}

type Cmp struct {
  Pos   lexer.Position
  Left  *Term     `@@`
  Right []*OpTerm `@@*`
}

type OpCmp struct {
  Pos      lexer.Position
  Operator Operator `@Operator`
  Cmp      *Cmp     `@@`
}

type Expression struct {
  Pos   lexer.Position
  Left  *Cmp     `@@`
  Right []*OpCmp `@@*`
}

type Assignment struct {
  Variable string      `@Ident`
  Value    *Expression `":" "="? @@`
}
