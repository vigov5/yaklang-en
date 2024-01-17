package javaclassparser

/*
*
string info itself does not store strings, only the constant pool index, which points to a CONSTANT_UTF8_INFO.

	CONSTANT_STRING_INFO {
		u1 tag;
		u2 string_index;
	}
*/
type ConstantStringInfo struct {
	Type               string
	StringIndex        uint16
	StringIndexVerbose string
}

func (self *ConstantStringInfo) readInfo(cp *ClassParser) {
	self.StringIndex = cp.reader.readUint16()
}
