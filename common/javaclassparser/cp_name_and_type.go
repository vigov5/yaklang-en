package javaclassparser

/*
*
is given The name and descriptor of the field or method

	CONSTANT_NAMEANDTYPE_INFO {
		u1 tag;
		u2 name_index;
		u2 descriptor_index
	}
*/
type ConstantNameAndTypeInfo struct {
	Type string
	//The field or method name points to a CONSTANT_UTF8_INFO
	NameIndex        uint16
	NameIndexVerbose string
	//The descriptor of the field or method points to a CONSTANT_UTF8_INFO
	DescriptorIndex        uint16
	DescriptorIndexVerbose string
}

func (self *ConstantNameAndTypeInfo) readInfo(cp *ClassParser) {
	self.NameIndex = cp.reader.readUint16()
	self.DescriptorIndex = cp.reader.readUint16()
}

/**
(1) Type descriptor
	①Basic type
	byte -> B
	short -> S
	char -> C
	int -> I
	long -> J *. Note that the descriptor of long is J not L
	float -> F
	double -> D
	②The descriptor of the reference type is L+the fully qualified name of the class + semicolon
	③The descriptor of the array type is [+array element type descriptor

(2) The field descriptor is the descriptor of the field type
(3) The method descriptor is (semicolon separated parameter type descriptor) + Return value type descriptor, void return value is represented by a single letter V

Ziyuan Note: The basic type of boolean should be Z
*/
