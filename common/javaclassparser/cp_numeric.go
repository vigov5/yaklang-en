package javaclassparser

import "math"

/*
*
integer in the constant pool
in the constant pool Four-byte storage integer constant

	CONSTANT_INTEGER_INFO {
		u1 tag;
		u4 bytes;
	}
*/
type ConstantIntegerInfo struct {
	Type string
	//. In fact, boolean, byte, short, and char smaller than int can also be placed in
	Value int32
}

func (self *ConstantIntegerInfo) readInfo(cp *ClassParser) {
	bytes := cp.reader.readUint32()
	self.Value = int32(bytes)
}

/*
*
Float
in the constant pool is four-byte

	CONSTANT_FLOAT_INFO {
		u1 tag;
		u4 bytes;
	}
*/
type ConstantFloatInfo struct {
	Type  string
	Value float32
}

func (self *ConstantFloatInfo) readInfo(cp *ClassParser) {
	bytes := cp.reader.readUint32()
	self.Value = math.Float32frombits(bytes)
}

/*
*
long
. Some special eight bytes are divided into high 8 characters. Section and lower 8 bytes

	CONSTANT_LONG_INFO {
		u1 tag;
		u4 high_bytes;
		u4 low_bytes;
	}
*/
type ConstantLongInfo struct {
	Type  string
	Value int64
}

func (self *ConstantLongInfo) readInfo(cp *ClassParser) {
	bytes := cp.reader.readUint64()
	self.Value = int64(bytes)
}

/*
*
. Double
in the constant pool is the same special eight-byte

	CONSTANT_DOUBLE_INFO {
		u1 tag;
		u4 high_bytes;
		u4 low_bytes;
	}
*/
type ConstantDoubleInfo struct {
	Type  string
	Value float64
}

func (self *ConstantDoubleInfo) readInfo(cp *ClassParser) {
	bytes := cp.reader.readUint64()
	self.Value = math.Float64frombits(bytes)
}
