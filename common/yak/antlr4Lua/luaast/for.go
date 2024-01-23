package luaast

import "github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"

func (l *LuaTranslator) enterForContext(start int) {
	l.forDepthStack.Push(&whileContext{
		startCodeIndex: start,
	})
}

func (l *LuaTranslator) exitForContext(end int) {
	start := l.peekWhileStartIndex()
	if start <= 0 {
		return
	}

	for _, c := range l.codes[start:] {
		if c.Opcode == yakvm.OpBreak && c.Unary <= 0 {
			// Set the jump value of the Break Code of all statements from the beginning to the end of while
			c.Unary = end
		}
	}
	l.forDepthStack.Pop()
}
