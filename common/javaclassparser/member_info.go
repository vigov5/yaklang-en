package javaclassparser

/*
*
field/Method
*/
type MemberInfo struct {
	Type               string
	AccessFlags        uint16
	AccessFlagsVerbose []string
	NameIndex          uint16
	NameIndexVerbose   string
	//descriptor
	DescriptorIndex        uint16
	DescriptorIndexVerbose string
	//attribute table
	Attributes []AttributeInfo
}
