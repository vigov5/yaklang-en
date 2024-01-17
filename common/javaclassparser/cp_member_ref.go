package javaclassparser

/*
*
ConstantFieldrefInfo, ConstantMethodrefInfo, ConstantInterfaceMethodrefInfo
These three structures inherit the concept of ConstantMemberrefInfo
Go language does not have“Inherits”, but are implemented through structure nesting
*/
type ConstantMemberrefInfo struct {
	ClassIndex              uint16
	ClassIndexVerbose       string
	NameAndTypeIndex        uint16
	NameAndTypeIndexVerbose string
}

/*
*
Field symbol reference

	CONSTANT_FIELDREF_INFO {
		u1 tag;
		u2 class_index;
		u2 name_and_type_index;
	}
*/
type ConstantFieldrefInfo struct {
	Type string
	ConstantMemberrefInfo
}

/*
*
Ordinary (non-interface) method symbol reference

	CONSTANT_METHODREF_INFO {
		u1 tag;
		u2 class_index;
		u2 name_and_type_index;
	}
*/
type ConstantMethodrefInfo struct {
	Type string
	ConstantMemberrefInfo
}

/*
*
Interface method symbol reference

	CONSTANT_INTERFACEMETHODREF_INFO {
		u1 tag;
		u2 class_index;
		u2 name_and_type_index;
	}
*/
type ConstantInterfaceMethodrefInfo struct {
	Type string
	ConstantMemberrefInfo
}

func (self *ConstantMemberrefInfo) readInfo(cp *ClassParser) {
	self.ClassIndex = cp.reader.readUint16()
	self.NameAndTypeIndex = cp.reader.readUint16()
}
