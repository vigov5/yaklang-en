package luaast

import (
	"fmt"
	lua "github.com/yaklang/yaklang/common/yak/antlr4Lua/parser"
)

func (l *LuaTranslator) VisitTableConstructor(raw lua.ITableconstructorContext) interface{} {
	if l == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*lua.TableconstructorContext)
	if i == nil {
		return nil
	}

	if fieldList := i.Fieldlist(); fieldList != nil {
		l.VisitFieldList(fieldList)
	} else {
		l.pushNewMap(0)
	}
	return nil
}

func (l *LuaTranslator) VisitFieldList(raw lua.IFieldlistContext) interface{} {
	if l == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*lua.FieldlistContext)
	if i == nil {
		return nil
	}

	indexCounter := 1 // indexCounter default to 1 in lua

	fields := i.AllField()
	// The field contains... What to do with
	// , this is a simple test. When... the variable parameter is passed in as{1,2,3,4}
	// If the table constructor is {"a",...,"b"} => {"a",1,"b"}
	// If the table constructor is {...} => {1,2,3,4}
	fieldLen := len(fields)
	if fieldLen == 1 && fields[0].GetText() == "..." {
		l.VisitVariadicField()
		l.pushNewVariadicMap(0) // The count here is meaningless. The size is determined dynamically.
		return nil
	}
	variadicFieldPos := 0
	variadicPresent := false
	for _, field := range fields {
		if field.GetText() == "..." {
			variadicPresent = true
			variadicFieldPos = indexCounter
			indexCounter++
			continue
		}
		l.VisitField(field, &indexCounter)
	}
	if !variadicPresent {
		l.pushNewMap(len(fields))
	} else {
		l.pushNewMapWithVariadicPos(len(fields), variadicFieldPos)
	}
	return nil
}

func (l *LuaTranslator) VisitField(raw lua.IFieldContext, indexCounter *int) interface{} {
	if l == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*lua.FieldContext)
	if i == nil {
		return nil
	}

	if i.LBracket() != nil {
		l.VisitExp(i.Exp(0))
		l.VisitExp(i.Exp(1))
		return nil
	}

	if i.NAME() != nil {
		l.pushString(i.NAME().GetText(), i.NAME().GetText())
		l.VisitExp(i.Exp(0))
		return nil
	}

	if len(i.AllExp()) == 1 {
		l.pushInteger(*indexCounter, fmt.Sprintf("%v", *indexCounter))
		l.VisitExp(i.Exp(0))
		*indexCounter++
		return nil

	}
	return nil
}

func (l *LuaTranslator) VisitVariadicField() interface{} {
	l.pushInteger(1, fmt.Sprintf("%v", 1))
	l.VisitVariadicEllipsis(false)
	return nil
}
