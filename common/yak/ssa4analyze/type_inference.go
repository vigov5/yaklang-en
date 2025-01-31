package ssa4analyze

import (
	"strings"

	"github.com/samber/lo"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/ssa"
)

const TITAG = "TypeInference"

type TypeInference struct {
	Finish     map[ssa.Value]struct{}
	DeleteInst []ssa.Instruction
}

func NewTypeInference(config) Analyzer {
	return &TypeInference{
		Finish: make(map[ssa.Value]struct{}),
	}
}

func (t *TypeInference) Run(prog *ssa.Program) {
	prog.EachFunction(func(f *ssa.Function) {
		t.RunOnFunction(f)
	})
}

func (t *TypeInference) RunOnFunction(fun *ssa.Function) {
	t.DeleteInst = make([]ssa.Instruction, 0)
	for _, b := range fun.Blocks {
		for _, inst := range b.Insts {
			t.InferenceOnInstruction(inst)
		}
	}
	for _, inst := range t.DeleteInst {
		ssa.DeleteInst(inst)
	}
}

func (t *TypeInference) InferenceOnInstruction(inst ssa.Instruction) {
	if iv, ok := inst.(ssa.Value); ok {
		t := iv.GetType()
		if utils.IsNil(t) {
			iv.SetType(ssa.BasicTypes[ssa.Null])
		}
	}

	switch inst := inst.(type) {
	case *ssa.Phi:
		// return t.TypeInferencePhi(inst)
	case *ssa.UnOp:
	case *ssa.BinOp:
		t.TypeInferenceBinOp(inst)
	case *ssa.Call:
		t.TypeInferenceCall(inst)
	// case *ssa.Return:
	// 	return t.TypeInferenceReturn(inst)
	// case *ssa.Switch:
	// case *ssa.If:
	case *ssa.Next:
		t.TypeInferenceNext(inst)
	case *ssa.Make:
		t.TypeInferenceMake(inst)
	case *ssa.Field:
		t.TypeInferenceField(inst)
	case *ssa.Update:
		// return t.TypeInferenceUpdate(inst)
	}
}

func collectTypeFromValues(values []ssa.Value, skip func(int, ssa.Value) bool) []ssa.Type {
	typMap := make(map[ssa.Type]struct{})
	typs := make([]ssa.Type, 0, len(values))
	for index, v := range values {
		// skip function
		if skip(index, v) {
			continue
		}
		// uniq typ
		typ := v.GetType()
		if _, ok := typMap[typ]; !ok {
			typMap[typ] = struct{}{}
			typs = append(typs, typ)
		}
	}
	return typs
}

// if all finish, return false
func (t *TypeInference) checkValuesNotFinish(vs []ssa.Value) bool {
	for _, v := range vs {
		if _, ok := t.Finish[v]; !ok {
			return true
		}
	}
	return false
}

/*
if v.Type !match typ return true
if v.Type match  typ return false
*/
func checkType(v ssa.Value, typ ssa.Type) bool {
	if v.GetType() == nil {
		v.SetType(typ)
		return false
	}
	v.SetType(typ)
	return true
}

func (t *TypeInference) TypeInferenceNext(next *ssa.Next) {
	/*
		next map[T]U

		{
			key: T
			field: U
			ok: bool
		}
	*/
	typ := ssa.NewStructType()
	typ.AddField(ssa.NewConst("ok"), ssa.BasicTypes[ssa.Boolean])
	if it, ok := next.Iter.GetType().(*ssa.ObjectType); ok {
		switch it.Kind {
		case ssa.SliceTypeKind:
			if next.InNext {
				typ.AddField(ssa.NewConst("key"), it.FieldType)
				typ.AddField(ssa.NewConst("field"), ssa.BasicTypes[ssa.Null])
			} else {
				typ.AddField(ssa.NewConst("key"), it.KeyTyp)
				typ.AddField(ssa.NewConst("field"), it.FieldType)
			}
		case ssa.StructTypeKind:
			typ.AddField(ssa.NewConst("key"), ssa.BasicTypes[ssa.String])
			typ.AddField(ssa.NewConst("field"), ssa.BasicTypes[ssa.Any])
		case ssa.ObjectTypeKind:
			typ.AddField(ssa.NewConst("key"), ssa.BasicTypes[ssa.Any])
			typ.AddField(ssa.NewConst("field"), ssa.BasicTypes[ssa.Any])
		case ssa.MapTypeKind:
			typ.AddField(ssa.NewConst("key"), it.KeyTyp)
			typ.AddField(ssa.NewConst("field"), it.FieldType)
		}
		next.SetType(typ)
	}
	if it, ok := next.Iter.GetType().(*ssa.ChanType); ok {
		typ.AddField(ssa.NewConst("key"), it.Elem)
		typ.AddField(ssa.NewConst("field"), ssa.BasicTypes[ssa.Null])
		next.SetType(typ)
		next.GetUsers().RunOnField(func(f *ssa.Field) {
			if f.Key.String() == "field" && len(f.GetAllVariables()) != 0 {
				// checkType(f, it.Elem)
				for _, variable := range f.GetAllVariables() {
					variable.NewError(ssa.Error, TITAG, InvalidChanType(it.Elem.String()))
				}
			}
		})
	}
}

func (t *TypeInference) TypeInferencePhi(phi *ssa.Phi) {
	// check
	// TODO: handler Acyclic graph
	if t.checkValuesNotFinish(phi.Edge) {
		return
	}

	// set type
	typs := collectTypeFromValues(
		phi.Edge,
		// // skip unreachable block
		func(index int, value ssa.Value) bool {
			block := phi.GetBlock().Preds[index]
			return block.Reachable() == -1
		},
	)

	// only first set type, phi will change
	phi.SetType(typs[0])
}

func (t *TypeInference) TypeInferenceBinOp(bin *ssa.BinOp) {
	XTyps := bin.X.GetType()
	YTyps := bin.Y.GetType()

	handlerBinOpType := func(x, y ssa.Type) ssa.Type {
		if x == nil {
			return y
		}
		if x.GetTypeKind() == y.GetTypeKind() {
			return x
		}

		if x.GetTypeKind() == ssa.Any {
			return y
		}
		if y.GetTypeKind() == ssa.Any {
			return x
		}

		// if y.GetTypeKind() == ssa.Null {
		if bin.Op >= ssa.OpGt && bin.Op <= ssa.OpNotEq {
			return ssa.BasicTypes[ssa.Boolean]
		}
		// }
		return nil
	}
	retTyp := handlerBinOpType(XTyps, YTyps)
	if retTyp == nil {
		// bin.NewError(ssa.Error, TITAG, "this expression type error: x[%s] %s y[%s]", XTyps, ssa.BinaryOpcodeName[bin.Op], YTyps)
		return
	}

	// typ := handler
	if bin.Op >= ssa.OpGt && bin.Op <= ssa.OpNotEq {
		bin.SetType(ssa.BasicTypes[ssa.Boolean])
		return
	} else {
		bin.SetType(retTyp)
		return
	}
}

func (t *TypeInference) TypeInferenceMake(i *ssa.Make) {
}

func (t *TypeInference) TypeInferenceField(f *ssa.Field) {
	if typ := f.Obj.GetType(); typ != nil {
		if methodTyp := ssa.GetMethod(typ, f.Key.String()); methodTyp != nil && f.GetType() != methodTyp {
			obj := f.Obj
			{
				names := lo.Keys(obj.GetAllVariables())
				if len(names) != 0 {
					// log.Errorf("method %s has no variable", ssa.LineDisasm(obj))
					// } else {
					obj.SetName(names[0])
				}
			}
			method := ssa.NewFunctionWithType(methodTyp.Name, methodTyp)
			f.GetUsers().RunOnCall(func(c *ssa.Call) {
				c.Args = utils.InsertSliceItem(c.Args, obj, 0)
				obj.AddUser(c)
			})
			ssa.ReplaceAllValue(f, method)
			f.GetProgram().SetInstructionWithName(method.GetName(), method)
			t.DeleteInst = append(t.DeleteInst, f)
			return
		}
		if utils.IsNil(typ) {
			typ = ssa.BasicTypes[ssa.Null]
		}
		switch typ.GetTypeKind() {
		case ssa.ObjectTypeKind, ssa.SliceTypeKind, ssa.MapTypeKind, ssa.StructTypeKind:
			interfaceTyp := f.Obj.GetType().(*ssa.ObjectType)
			fTyp := interfaceTyp.GetField(f.Key)
			if !utils.IsNil(fTyp) {
				f.SetType(fTyp)
				return
			}
		case ssa.String:
			f.SetType(ssa.BasicTypes[ssa.Number])
			return
		case ssa.Any:
			// pass
			f.SetType(ssa.BasicTypes[ssa.Any])
			return
		default:
		}
		// if object is call, just skip, because call will return map or slice, we don't know what in this map or slice
		c, ok := ssa.ToCall(f.Obj)
		if ok {
			if c.Unpack {
				return
			}
			funTyp, ok := ssa.ToFunctionType(c.Method.GetType())
			if ok {
				// if ssa.IsObjectType(funTyp.ReturnType) {
				if kind := funTyp.ReturnType.GetTypeKind(); kind == ssa.SliceTypeKind || kind == ssa.MapTypeKind {
					return
				}
			} else {
				return
			}
		}
		// if c, ok := ssa.ToCall(f.Obj); ok
		// } else {
		text := ""
		if ci, ok := ssa.ToConst(f.Key); ok {
			text = ci.String()
			want := ssa.TryGetSimilarityKey(ssa.GetAllKey(typ), text)
			if want != "" {
				f.NewError(ssa.Error, TITAG, ssa.ExternFieldError("Type", typ.String(), text, want))
				return
			}
		}
		if text == "" {
			list := strings.Split(*f.GetRange().SourceCode, ".")
			text = list[len(list)-1]
		}
		f.Key.NewError(ssa.Error, TITAG, InvalidField(typ.String(), text))
		// }
	}
	// use update
	// vs := lo.FilterMap(f.GetValues(), func(v ssa.Value, i int) (ssa.Value, bool) {
	// 	// switch v := v.(type) {
	// 	// // case *ssa.Update:
	// 	// // 	return v.Value, true
	// 	// default:
	// 	// 	return nil, false
	// 	// }
	// })

	// // check value finish
	// // TODO: handler Acyclic Graph
	// if t.checkValuesNotFinish(vs) {
	// 	return
	// }

	// ts := collectTypeFromValues(
	// 	// f.Update,
	// 	vs,
	// 	func(i int, v ssa.Value) bool { return false },
	// )
	// if len(ts) == 0 {
	// 	f.SetType(ssa.BasicTypes[ssa.Null])
	// } else if len(ts) == 1 {
	// 	f.SetType(ts[0])
	// } else {
	// 	f.SetType(ssa.BasicTypes[ssa.Any])
	// }
}

func (t *TypeInference) TypeInferenceCall(c *ssa.Call) {

	// get function type
	funcTyp, ok := ssa.ToFunctionType(c.Method.GetType())
	if !ok {
		return
	}

	sideEffect := funcTyp.SideEffects
	if funcTyp.IsMethod && funcTyp.IsModifySelf {
		sideEffect = append(sideEffect, c.Args[0].GetName())
	}

	// handle FreeValue
	if len(funcTyp.FreeValue) != 0 || len(sideEffect) != 0 {
		c.HandleFreeValue(funcTyp.FreeValue, sideEffect)
	}

	// handle ellipsis, unpack argument
	if c.IsEllipsis {
		// getField := func(object ssa.User, key ssa.Value) *ssa.Field {
		// 	var f *ssa.Field
		// 	if f = ssa.GetField(object, key); f == nil {
		// 		f = ssa.NewFieldOnly(key, object, c.Block)
		// 		ssa.EmitBefore(c, f)
		// 	}
		// 	return f
		// }
		// obj := c.Args[len(c.Args)-1].(ssa.User)
		// num := len(ssa.GetFields(obj))
		// if t, ok := obj.GetType().(*ssa.ObjectType); ok {
		// 	if t.Kind == ssa.Slice {
		// 		num = len(t.Key)
		// 	}
		// }

		// // fields := ssa.GetFields(obj)
		// c.Args[len(c.Args)-1] = getField(obj, ssa.NewConst(0))
		// for i := 1; i < num; i++ {
		// 	c.Args = append(c.Args, getField(obj, ssa.NewConst(i)))
		// }
	}

	// inference call instruction type
	if c.IsDropError {
		if t, ok := funcTyp.ReturnType.(*ssa.ObjectType); ok {
			if t.Combination && t.FieldTypes[len(t.FieldTypes)-1].GetTypeKind() == ssa.ErrorType {
				// if len(t.FieldTypes) == 1 {
				// 	c.SetType(ssa.BasicTypes[ssa.Null])
				// } else if len(t.FieldTypes) == 2 {
				if len(t.FieldTypes) == 2 {
					c.SetType(t.FieldTypes[0])
				} else {
					ret := ssa.NewStructType()
					ret.FieldTypes = t.FieldTypes[:len(t.FieldTypes)-1]
					ret.Keys = t.Keys[:len(t.Keys)-1]
					ret.KeyTyp = t.KeyTyp
					ret.Combination = true
					c.SetType(ret)
				}
				return
			}
		} else if t, ok := funcTyp.ReturnType.(*ssa.BasicType); ok && t.Kind == ssa.ErrorType {
			// pass
			c.SetType(ssa.BasicTypes[ssa.Null])
			for _, variable := range c.GetAllVariables() {
				variable.NewError(ssa.Error, TITAG, ValueIsNull())
			}
			return
		}
		c.NewError(ssa.Warn, TITAG, FunctionContReturnError())
	} else {
		c.SetType(funcTyp.ReturnType)
	}
}
