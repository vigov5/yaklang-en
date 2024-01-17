package yakvm

import (
	"fmt"
)

func (v *Frame) assign(left, right *Value) {
	if !(left.IsValueList() && right.IsValueList()) {
		panic("BUG: assign: left and right must be value list")
	}

	leftValueList := left.ValueList()
	rightValueList := right.ValueList()

	if len(leftValueList) <= 0 || len(rightValueList) <= 0 {
		return
	}

	// is when the type is needed. There is one value on the left and n values on the right. The right ones can be directly given to the left
	if len(leftValueList) == 1 {
		left := leftValueList[0]
		if len(rightValueList) == 1 {
			// , one on the left and one on the right. The simplest assignment is
			left.Assign(v, rightValueList[0])
		} else {
			// Left one right N
			left.Assign(v, NewValues(rightValueList))
			//tbl, err := v.RootSymbolTable.FindSymbolTableBySymbolId(leftValueList[0].SymbolId)
			//if err != nil {
			//	panic(err)
			//}
			//leftValueList[0].AssignBySymbol(tbl, right.ValueListToInterface())
		}
		return
	}

	// The left side has n values, and the right side It is 1 value, and the right side needs to be split. The most basic situation of Array and Slice
	if len(rightValueList) == 1 {
		right := rightValueList[0]
		if !right.IsIterable() {
			panic("multi-assign failed: right value is not iterable")
		}
		rightValueLen := right.Len()
		if rightValueLen != len(leftValueList) {
			panic(fmt.Sprintf("multi-assign failed: left value length[%d] != right value length[%d]", len(leftValueList), rightValueLen))
		}
		for index, val := range leftValueList {
			val.Assign(v, NewValue("__assign_middle__", right.CallSliceIndex(index), ""))
			//data := right.CallSliceIndex(index)
			//tbl, err := v.RootSymbolTable.FindSymbolTableBySymbolId(leftValueList[index].SymbolId)
			//if err != nil {
			//	panic(err)
			//}
			//val.AssignBySymbol(tbl, data)
		}
		return
	}

	// There are n on the left Value, there are m values on the right, all greater than one, then they must be equal, otherwise
	if len(rightValueList) != len(leftValueList) {
		panic("multi-assign failed: left value length[" + fmt.Sprint(len(leftValueList)) + "] != right value length[" + fmt.Sprint(len(rightValueList)) + "]")
	}

	for i := 0; i < len(rightValueList); i++ {
		leftValueList[i].Assign(v, rightValueList[i])
		//leftValue := leftValueList[i]
		//tbl, err := v.RootSymbolTable.FindSymbolTableBySymbolId(leftValue.SymbolId)
		//if err != nil {
		//	panic(err)
		//}
		//leftValue.AssignBySymbol(tbl, rightValueList[i].Value)
	}
}

func (v *Frame) luaLocalAssign(leftPart, rightPart *Value) {
	if !(leftPart.IsValueList() && rightPart.IsValueList()) {
		panic("BUG: assign: left and right must be value list")
	}

	leftValueList := leftPart.ValueList()
	rightValueList := rightPart.ValueList()

	if len(leftValueList) <= 0 || len(rightValueList) <= 0 {
		return
	}

	// . In this case, there is only one value on the right. There is an overwriting situation
	if len(rightValueList) == 1 {
		right := rightValueList[0]
		if right.Value == nil {
			for _, left := range leftValueList {
				left.Assign(v, right)
			}
			return
		}
		// In Lua, the table will be assigned directly when assigning a value, which belongs to the non-iterable corresponding table The implementation uses map, which is also non-iterable and just handles this situation.
		if !right.IsIterable() {
			for index, left := range leftValueList {
				if index == 0 {
					left.Assign(v, right)
					continue
				}
				left.Assign(v, undefined)
			}
			return
		}
		// can iterate and consider the number of values.
		rightValueLen := right.Len()
		leftValueLen := len(leftValueList)
		if rightValueLen != leftValueLen { // is not equal. At this time, you have to ignore the value or fill in nil depending on the situation.
			if rightValueLen > leftValueLen {
				if _, ok := right.Value.([]interface{}); ok {
					right.Value = right.Value.([]interface{})[:len(leftValueList)]
				} else {
					right.Value = right.Value.([]*Value)[:len(leftValueList)]
				}
			} else {
				for i := 0; i < leftValueLen-rightValueLen; i++ {
					if _, ok := right.Value.([]interface{}); ok {
						right.Value = append(right.Value.([]interface{}), nil)
					} else {
						right.Value = append(right.Value.([]*Value), undefined)
					}
				}
			}
		}
		for index, left := range leftValueList {
			if _, ok := right.CallSliceIndex(0).(*Value); ok {
				left.Assign(v, NewValue("__assign_middle__", right.CallSliceIndex(index).(*Value).Value, ""))
			} else {
				left.Assign(v, NewValue("__assign_middle__", right.CallSliceIndex(index), ""))
			}
		}
		return
	}

	// . There are multiple values on the left and right. At this time, there will be an operation similar to binding assignment.
	// . At this time, the left and right correspond to each other one-to-one. After the left and right are separated, it is similar to the situation of len(leftValueList) == 1.

	leftLen, rightLen := len(leftValueList), len(rightValueList)

	if leftLen != rightLen {
		if rightLen > leftLen {
			rightValueList = rightValueList[:len(leftValueList)]
		} else {
			for i := 0; i < leftLen-rightLen; i++ {
				rightValueList = append(rightValueList, undefined)
			}
		}
	}

	for index := 0; index < len(leftValueList); index++ {
		left := leftValueList[index]
		right := rightValueList[index]
		if right.Value == nil {
			left.Assign(v, right)
			continue
		}
		// There is a small The pitfall reflect.TypeOf(nil).Kind() will be abnormal, so first deal with the case where nil exists on the right side
		if right.IsIterable() {
			if _, ok := right.CallSliceIndex(0).(*Value); ok {
				left.Assign(v, NewValue("__assign_middle__", right.CallSliceIndex(0).(*Value).Value, ""))
			} else {
				left.Assign(v, NewValue("__assign_middle__", right.CallSliceIndex(0), ""))
			}
			continue
		} else { // The right side is non-iterable
			// , one on the left and one on the right. The simplest assignment is
			left.Assign(v, right)
			continue
		}
	}
}

func (v *Frame) luaGlobalAssign(leftPart, rightPart *Value) {
	if !(leftPart.IsValueList() && rightPart.IsValueList()) {
		panic("BUG: assign: left and right must be value list")
	}

	leftValueList := leftPart.ValueList()
	rightValueList := rightPart.ValueList()

	if len(leftValueList) <= 0 || len(rightValueList) <= 0 {
		return
	}

	//will fail. Multiple assignments are a bit complicated. There are several situations. The right side of
	//There is only one value on the left, which is a special case of multi-assignment.
	//if len(leftValueList) == 1 {
	//	left := leftValueList[0]
	//	right := rightValueList[0]
	//	if right.Value == nil {
	//		left.GlobalAssign(v, right)
	//		return
	//	}
	//	// There is a small The pitfall reflect.TypeOf(nil).Kind() will be abnormal, so first deal with the case where nil exists on the right side
	//	if right.IsIterable() {
	//		if _, ok := right.CallSliceIndex(0).(*Value); ok { // Here is the case where the return value of the detection function is a function
	//			left.GlobalAssign(v, NewValue("__assign_middle__", right.CallSliceIndex(0).(*Value).Value, ""))
	//		} else {
	//			left.GlobalAssign(v, NewValue("__assign_middle__", right.CallSliceIndex(0), ""))
	//		}
	//		return
	//	} else { // The right side is non-iterable
	//		// , one on the left and one on the right. The simplest assignment is
	//		left.GlobalAssign(v, right)
	//		return
	//	}
	//}

	// . In this case, there is only one value on the right. There is an overwriting situation
	if len(rightValueList) == 1 {
		right := rightValueList[0]
		if right.Value == nil {
			for _, left := range leftValueList {
				left.GlobalAssign(v, right)
			}
			return
		}
		// In Lua, the table will be assigned directly when assigning a value, which belongs to the non-iterable corresponding table The implementation uses map, which is also non-iterable and just handles this situation.
		if !right.IsIterable() {
			for index, left := range leftValueList {
				if index == 0 {
					left.GlobalAssign(v, right)
					continue
				}
				left.GlobalAssign(v, undefined)
			}
			return
		}
		// can iterate and consider the number of values.
		rightValueLen := right.Len()
		leftValueLen := len(leftValueList)
		if rightValueLen != leftValueLen { // is not equal. At this time, you have to ignore the value or fill in nil depending on the situation.
			if rightValueLen > leftValueLen {
				if _, ok := right.Value.([]interface{}); ok {
					right.Value = right.Value.([]interface{})[:len(leftValueList)]
				} else {
					right.Value = right.Value.([]*Value)[:len(leftValueList)]
				}
			} else {
				for i := 0; i < leftValueLen-rightValueLen; i++ {
					if _, ok := right.Value.([]interface{}); ok {
						right.Value = append(right.Value.([]interface{}), nil)
					} else {
						right.Value = append(right.Value.([]*Value), undefined)
					}
				}
			}
		}
		for index, left := range leftValueList {
			if _, ok := right.CallSliceIndex(0).(*Value); ok {
				left.GlobalAssign(v, NewValue("__assign_middle__", right.CallSliceIndex(index).(*Value).Value, ""))
			} else {
				left.GlobalAssign(v, NewValue("__assign_middle__", right.CallSliceIndex(index), ""))
			}
		}
		return
	}

	// . There are multiple values on the left and right. At this time, there will be an operation similar to binding assignment.
	// . At this time, the left and right correspond to each other one-to-one. After the left and right are separated, it is similar to the situation of len(leftValueList) == 1.

	leftLen, rightLen := len(leftValueList), len(rightValueList)

	if leftLen != rightLen {
		if rightLen > leftLen {
			rightValueList = rightValueList[:len(leftValueList)]
		} else {
			for i := 0; i < leftLen-rightLen; i++ {
				rightValueList = append(rightValueList, undefined)
			}
		}
	}

	for index := 0; index < len(leftValueList); index++ {
		left := leftValueList[index]
		right := rightValueList[index]
		if right.Value == nil {
			left.GlobalAssign(v, right)
			continue
		}
		// There is a small The pitfall reflect.TypeOf(nil).Kind() will be abnormal, so first deal with the case where nil exists on the right side
		if right.IsIterable() {
			if _, ok := right.CallSliceIndex(0).(*Value); ok {
				left.GlobalAssign(v, NewValue("__assign_middle__", right.CallSliceIndex(0).(*Value).Value, ""))
			} else {
				left.GlobalAssign(v, NewValue("__assign_middle__", right.CallSliceIndex(0), ""))
			}
			continue
		} else { // The right side is non-iterable
			// , one on the left and one on the right. The simplest assignment is
			left.GlobalAssign(v, right)
			continue
		}
	}
}
