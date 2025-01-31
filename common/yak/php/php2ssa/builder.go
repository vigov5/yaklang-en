package php2ssa

import (
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/yak/antlr4util"
	phpparser "github.com/yaklang/yaklang/common/yak/php/parser"
	"github.com/yaklang/yaklang/common/yak/ssa"
)

type builder struct {
	ast  phpparser.IHtmlDocumentContext
	prog *ssa.Program
	ir   *ssa.FunctionBuilder
}

func ParseSSA(src string, f func(builder *ssa.FunctionBuilder)) (prog *ssa.Program) {
	lex := phpparser.NewPHPLexer(antlr.NewInputStream(src))
	tokenStream := antlr.NewCommonTokenStream(lex, antlr.TokenDefaultChannel)
	parser := phpparser.NewPHPParser(tokenStream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(antlr4util.NewLegacyErrorListener())
	program := ssa.NewProgram()
	builder := &builder{
		prog: program,
		ast:  parser.HtmlDocument(),
	}
	builder.Build()
	for _, r := range builder.prog.GetErrors() {
		log.Errorf("ssa-ir program error: %v", r)
	}
	return builder.prog
}

func (y *builder) Build() {
	pkg := ssa.NewPackage("main")
	y.prog.AddPackage(pkg)
	main := pkg.NewFunction("main")
	y.ir = ssa.NewBuilder(main, nil)
	y.VisitHtmlDocument(y.ast)
	y.ir.Finish()
}
