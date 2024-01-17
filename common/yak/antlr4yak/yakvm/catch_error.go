package yakvm

func (v *Frame) catchErrorRun(catchCodeIndex, id int) (err interface{}) {
	// backs up the scope and restores the scope after exiting the try block. , to prevent incorrect scope after abnormal exit.
	scopeBackpack := v.scope

	// Exit Try case
	// 1. (OpExitCatchError)OpBreak
	// 2. (OpExitCatchError)OpContinue
	// 3. (OpExitCatchError)OpReturn
	// 4. panic
	// 5. Exit
	// 1/2 does not require exiting the scope, because break and continue will handle the scope by themselves.
	// situation 3 requires manual restoration of the scope to outside try-catch-finally.
	// Case 4 requires manual restoration of the scope to try-catch-finally
	// after normal execution of try-block Case 5 requires manual restoration of the role Scope to try-catch-finally
	defer func() {
		// Divide 1/2 situation requires restoration to try-catch-finally scope.
		if !(v.codes[v.codePointer+1].Opcode == OpBreak || v.codes[v.codePointer+1].Opcode == OpContinue) {
			v.scope = scopeBackpack
			// Case 3 requires exiting the try-catch-finally scope
			if v.codes[v.codePointer+1].Opcode == OpReturn {
				v.ExitScope()
			}
		}

		//After an error occurs, assign a value to err and jump to the catch block.
		if err != nil {
			v.setCodeIndex(catchCodeIndex)
			if id > 0 {
				NewValueRef(id).Assign(v, NewAutoValue(err))
			}
		}
	}()
	v.codePointer++
	v.continueExec()
	err = v.recover().GetData()
	return
}
