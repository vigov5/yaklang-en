package yakvm

import (
	"bytes"
	"fmt"
)

type OpcodeFlag int

const (
	OpNop OpcodeFlag = iota

	// OpTypeCast takes out two values from the stack, the first value is the type, and the second value is the specific value that needs to be converted, perform type conversion, and push the result into the stack
	OpTypeCast // type convert
	// Unary
	OpNot  // !
	OpNeg  // -
	OpPlus // +
	OpChan // <-

	OpPlusPlus   // ++
	OpMinusMinus // --
	/*
		^ & * <-
	*/

	// Binary
	OpShl    // <<
	OpShr    // >>
	OpAnd    // &
	OpAndNot // &^
	OpOr     // |
	OpXor    // ^

	OpAdd  // +
	OpSub  // -
	OpMul  // *
	OpDiv  // /
	OpMod  // %
	OpGt   // >
	OpLt   // <
	OpGtEq // >=
	OpLtEq // <=

	OpEq       // ==
	OpNotEq    // != <>
	OpPlusEq   // +=
	OpMinusEq  // -=
	OpMulEq    // *=
	OpDivEq    // /=
	OpModEq    // %=
	OpAndEq    // &=
	OpOrEq     // |=
	OpXorEq    // ^=
	OpShlEq    // <<=
	OpShrEq    // >>=
	OpAndNotEq // &^=

	OpIn       // in
	OpSendChan // <-

	OpType // type
	OpMake // make

	OpPush     // Push an Op1 onto the stack
	OpPushfuzz // Put an Op1, perform the Fuzz String operation, and push it onto the stack

	/*
		for range and for in needs to iterate the result of the rvalue expression, and needs to cooperate with Op to implement
	*/
	OpRangeNext // takes an element from the stack, then iterates, and pushes it onto the stack
	OpInNext    // takes an element from the stack, then iterates, and pushes it onto the stack. It is slightly different from range. For example, you can unpack a slice and the first value when iterating a slice is value instead of index.
	OpEnterFR   // Enter for range, create an iterator from the peek value in the stack, and push the iterator into the IteratorStack stack.
	OpExitFR    // exits for range and determines whether the IteratorStack has been End, if it is not over, jump to the beginning of the for range (Unary), otherwise it will pop IteratorStack and continue to execute the subsequent code

	OpList // lvalue reference, is an exclusive push generally used for assignment

	// OpAssign This operation has two parameters
	// will generally be ValueList, so TypeVerbose is list. , the type assertion [] will be set.*Value
	// After popArgN, 0 is an rvalue and 1 is an lvalue
	OpAssign

	// OpNewSlice is used to create Slice from the stack, take out unary data, infer the type, and then combine it into a slice
	// and take two values from the stack. arg1 is the left symbol and arg2 is the specific value.
	// performs fast assignment, directly assigns the value to the symbol left, and continues to push arg2 onto the stack.
	OpFastAssign

	// push A value corresponding to a symbol, this value has no operands op1 op2 only operates unary
	OpPushRef
	// OpFastAssign fast assignment, exists in special assignments **Value
	// This only operates. Unary, Unary passes the specific symbol
	OpPushLeftRef

	// . Except for the jump instruction, other instructions should not directly operate the index!
	// JMP to unconditionally jump to which instruction, and unary records the number of instructions
	OpJMP
	// takes a value from the stack. If the value is true, jumps to the instruction of unary.
	OpJMPT
	// takes a value from the stack. The value is false and jumps to the unary instruction.
	OpJMPF
	// with variable parameters and checks the latest value from the stack. If the value is true, jump to the unary position. Otherwise, pop the stack data TbbbT. The left and right values of
	OpJMPTOP
	// From the stack Check the latest value. If the value is false, jump to the unary position. Otherwise, pop the stack data
	OpJMPFOP

	// OpBreak This statement is set for break. It is generally used to record the jump location. It is basically equivalent to JMP.
	// The difference is that the position of Break cannot be predicted in advance.
	// . Therefore, the last step at the end of the for loop is needed to find the value that is not set in the current for loop. When passing the break statement
	// Before being set, Unary should be less than or equal to zero
	// Because Break also operates the pointer, so it should not happen after OpBreak ends. Operation pointer
	//
	// . The difference between/and then pop out the content that should be called in the stack
	// . Therefore, these two executions When you need to add the stack to push
	OpBreak
	OpContinue

	// OpCall / OpVariadicCall takes the number of pops from unary.
	// and JMP is that Break
	OpCall
	OpVariadicCall /*This is a call for variable parameters.*/

	// OpPushId push a reference name. This name may not be available. Symbols in the symbol table, but are used
	// that needs to be executed when the virtual machine exits, so this cannot use unary, use op1 type Identifier as the operand, if not found, it is nil or undefined
	OpPushId

	// OpPop is generally used to maintain stack balance. For example, push an expression statement, which generally does not pop. You need to use OpPop to pop.
	OpPop

	// OpNewMap This command is used to create a map and remove unary from the stack * needs to be replaced. This operation has one parameter. Take how many elements from the stack to form a list. Use unary: int to mark the 2 data of TbbbT, and then combine them. The left side is the key and the right side is the value.
	OpNewMap

	// OpNewMapWithType This instruction is used to create a map, take unary * 2 data, and then combines the two, with the key on the left and the value on the right, and then takes out the Type and combines it into Map
	OpNewMapWithType

	// Continue will be destroyed. scope stack balance
	OpNewSlice

	// OpNewSliceWithType Create Slice from the stack, take out unary operands, then take out the Type, combine it into Slice
	OpNewSliceWithType

	// OpSliceCall index Silice
	OpIterableCall

	// OpReturn takes a data from the stack and copies it to the return value cache data. Generally speaking, you can use lastStackValue to get the number.
	OpReturn

	// OpDefer executes op1, generally the value of op1 must be codes, which is []*Opcode
	// from the stack as the value
	OpDefer
	// OpAssert Takes one or two parameters from the stack, and then asserts the type. If it is false, panic the second parameter. The parameter
	OpAssert

	// OpMemberCall Gets member variables or methods of map or structure
	OpMemberCall

	// OpAsyncCall Execute goroutine
	// unary is
	OpAsyncCall

	// OpScope will create a new domain and stop the domain through OpScopeEnd
	// domain is a tree structure, which saves the reference of the parent domain, because we need to see the content of the parent domain.
	OpScope
	OpScopeEnd

	// include will directly pop the file path from the stack and then execute
	OpInclude

	// OpPanic actively panics and then returns the error. Implement
	OpPanic
	OpRecover

	// for Defers Recover. The OpEllipsis function calls unpacking
	OpEllipsis
	// OpBitwiseNot Bitwise inversion
	OpBitwiseNot

	OpCatchError
	OpStopCatchError
	OpExit
)

func (f OpcodeFlag) IsJmp() bool {
	return f == OpJMP || f == OpJMPT || f == OpJMPF || f == OpJMPTOP || f == OpJMPFOP || f == OpRangeNext || f == OpInNext || f == OpBreak || f == OpContinue || f == OpEnterFR || f == OpExitFR
}

func OpcodeToName(op OpcodeFlag) string {
	i, ok := OpcodeVerboseName[op]
	if ok {
		return i
	}
	return fmt.Sprintf("unknown[%v]", op)
}

var OpcodeVerboseName = map[OpcodeFlag]string{
	OpBitwiseNot: `not`,
	OpAnd:        `and`,
	OpAndNot:     `and-not`,
	OpOr:         `or`,
	OpXor:        `xor`,
	OpShl:        `shl`,
	OpShr:        `shr`,
	OpTypeCast:   `type-cast`,
	OpPlusPlus:   `self-add-one`,
	OpMinusMinus: `self-minus-one`,
	OpNot:        `bang`,
	OpNeg:        `neg`,
	OpPlus:       `plus`,
	OpChan:       `chan-recv`,
	OpAdd:        `add`,
	OpSub:        `sub`,
	OpMod:        `mod`,
	OpMul:        `mul`,
	OpDiv:        `div`,
	OpGt:         `gt`,
	OpLt:         `lt`,
	OpLtEq:       `lt-eq`,
	OpGtEq:       `gt-eq`,
	OpNotEq:      `neq`,
	OpEq:         `eq`,
	OpPlusEq:     `self-plus-eq`,
	OpMinusEq:    `self-minus-eq`,
	OpMulEq:      `self-mul-eq`,
	OpDivEq:      `self-div-eq`,
	OpModEq:      `self-mod-eq`,
	OpAndEq:      `self-and-eq`,
	OpOrEq:       `self-or-eq`,
	OpXorEq:      `self-xor-eq`,
	OpShlEq:      `self-shl-eq`,
	OpShrEq:      `self-shr-eq`,
	OpAndNotEq:   `self-and-not-eq`,

	OpIn:       `in`,
	OpSendChan: `chan-send`,

	OpRangeNext: `range-next`,
	OpInNext:    `in-next`,
	OpEnterFR:   `enter-for-range`,
	OpExitFR:    `exit-for-range`,

	OpType:        `type`,
	OpMake:        `make`,
	OpPush:        `push`,
	OpList:        `list`,
	OpAssign:      `assign`,
	OpFastAssign:  `fast-assign`,
	OpPushRef:     `pushr`,
	OpPushLeftRef: `pushleftr`,
	OpJMP:         `jmp`,
	OpJMPT:        `jmpt`,
	OpJMPF:        `jmpf`,
	OpJMPTOP:      `jmpt-or-pop`,
	OpJMPFOP:      `jmpf-or-pop`,

	OpCall:             `call`,
	OpVariadicCall:     `callvar`,
	OpPushId:           `pushid`,
	OpPop:              `pop`,
	OpPushfuzz:         `pushf`,
	OpNewMap:           `newmap`,
	OpNewMapWithType:   `typedmap`,
	OpNewSlice:         `newslice`,
	OpNewSliceWithType: `typedslice`,
	OpIterableCall:     `iterablecall`,
	OpReturn:           `return`,
	OpAssert:           `assert`,
	OpDefer:            `defer`,
	OpMemberCall:       `membercall`,
	OpBreak:            `break`,
	OpContinue:         `continue`,
	OpAsyncCall:        "async-call",

	OpScope:    `new-scope`,
	OpScopeEnd: `end-scope`,

	OpInclude: `include`,

	OpRecover:  `recover`,
	OpPanic:    `panic`,
	OpEllipsis: `ellipsis`,

	OpCatchError:     `catch-error`,
	OpStopCatchError: `stop-catch-error`,
	OpExit:           `exit`,
}

type Code struct {
	Opcode OpcodeFlag

	Unary int
	Op1   *Value
	Op2   *Value

	// records the position of Opcode.
	SourceCodeFilePath *string
	SourceCodePointer  *string
	StartLineNumber    int
	StartColumnNumber  int
	EndLineNumber      int
	EndColumnNumber    int
}

func (c *Code) IsJmp() bool {
	return c.Opcode.IsJmp()
}

func (c *Code) GetJmpIndex() int {
	flag := c.Opcode
	if !flag.IsJmp() {
		return -1
	}
	if flag == OpInNext || flag == OpRangeNext {
		return c.Op1.Int()
	}
	return c.Unary
}

func (c *Code) RangeVerbose() string {
	return fmt.Sprintf(
		"%v:%v->%v:%v",
		c.StartLineNumber, c.StartColumnNumber,
		c.EndLineNumber, c.EndColumnNumber,
	)
}

func (c *Code) String() string {
	var buf bytes.Buffer
	op, ok := OpcodeVerboseName[c.Opcode]
	if !ok {
		op = "unknown[" + fmt.Sprint(c.Opcode) + "]"
	}
	buf.WriteString(fmt.Sprintf("OP:%-20s", op) + " ")
	switch c.Opcode {
	case OpBitwiseNot, OpAnd, OpAndNot, OpOr, OpXor, OpShl, OpShr:
	case OpTypeCast:
	case OpPlusPlus, OpMinusMinus:
	case OpNot, OpNeg, OpPlus, OpChan:
	case OpScope:
		buf.WriteString(fmt.Sprint(c.Unary))
	case OpScopeEnd:
	case OpMake:
	case OpType:
		buf.WriteString(c.Op1.TypeVerbose)
	case OpPush:
		buf.WriteString(c.Op1.String())
		if c.Unary == 1 {
			buf.WriteString(" (copy)")
		}
	case OpPushId, OpPushfuzz:
		buf.WriteString(c.Op1.String())
		// Special push is used to process f as prefix prefix push string
	case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpIn, OpSendChan:
	case OpGt, OpLt, OpGtEq, OpLtEq, OpEq, OpNotEq, OpPlusEq, OpMinusEq, OpMulEq, OpDivEq, OpModEq, OpAndEq, OpOrEq, OpXorEq, OpShlEq, OpShrEq:
	case OpPop, OpReturn, OpRecover, OpPanic:
	case OpCall, OpVariadicCall, OpDefer, OpAsyncCall:
		buf.WriteString(fmt.Sprintf("vlen:%d", c.Unary))
	case OpRangeNext, OpInNext, OpFastAssign:
	case OpJMP, OpJMPT, OpJMPF, OpJMPTOP, OpJMPFOP, OpEnterFR, OpExitFR:
		buf.WriteString(fmt.Sprintf("-> %d", c.Unary))
	case OpAssign:
		switch c.Op1.String() {
		case "nasl_global_declare", "nasl_declare":
			buf.WriteString("-> with pop")
		}
	case OpContinue:
		buf.WriteString(fmt.Sprintf("-> %d (-%d scope)", c.Unary, c.Op1.Int()))
	case OpBreak:
		if c.Op2.Int() != 2 {
			buf.WriteString(fmt.Sprintf("-> %d (-%d scope) mode: %v", c.Unary, c.Op1.Int(), c.Op2.String()))
		} else {
			buf.Reset()
			buf.WriteString(fmt.Sprintf("OP:%-20s", "fallthrough") + " ")
			buf.WriteString(fmt.Sprintf("-> %d (-%d scope)", c.Unary, c.Op1.Int()))
		}
	case OpList, OpPushRef, OpNewMap, OpNewMapWithType, OpNewSlice, OpNewSliceWithType, OpPushLeftRef:
		buf.WriteString(fmt.Sprint(c.Unary))
	case OpCatchError:
		buf.WriteString(fmt.Sprintf("err -> %d", c.Op1.Int()+1))
	case OpStopCatchError, OpExit:
	default:
		if c.Unary > 0 {
			buf.WriteString("off:" + fmt.Sprint(c.Unary) + " ")
		}
		buf.WriteString("op1: " + c.Op1.String())
		buf.WriteString("\t\t")
		buf.WriteString("op2: " + c.Op2.String())
	}
	return buf.String()
}
func (c *Code) Dump() {
	println(fmt.Sprintf("%-13s %v", c.RangeVerbose(), c.String()))
}
