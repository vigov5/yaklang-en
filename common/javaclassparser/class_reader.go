package javaclassparser

import "encoding/binary"

/*
*
jvm defines u1, u2, and u4 to represent 1, 2, and 4-byte unsigned integers.
Multiple pieces of data of the same type are generally stored in class files in the form of tables, consisting of table headers and table items. The header is a u2 or u4 integer.
. Assume that the header is 10, followed by 10 entries. Data
*/
type ClassReader struct {
	//class data is stored in the smallest unit of byte, and 8-bit
	data []byte
}

func NewClassReader(data []byte) *ClassReader {
	return &ClassReader{data: data}
}

/*
*
is equivalent to javas byte 8-bit unsigned integer.
*/
func (this *ClassReader) readUint8() uint8 {
	val := this.data[0]
	this.data = this.data[1:]
	return val
}

/*
*
is equivalent to Javas short 16-bit unsigned integer
. Here, the class file is stored in the big endian method in the file system.
*/
func (this *ClassReader) readUint16() uint16 {
	//. The big endian method reads 16-bit data.
	val := binary.BigEndian.Uint16(this.data)
	this.data = this.data[2:]
	return val
}

/*
*
, which is equivalent to Javas int 32-bit unsigned integer
*/
func (this *ClassReader) readUint32() uint32 {
	val := binary.BigEndian.Uint32(this.data)
	this.data = this.data[4:]
	return val
}

/*
*
is equivalent to javas long 64-bit unsigned integer
*/
func (this *ClassReader) readUint64() uint64 {
	val := binary.BigEndian.Uint64(this.data)
	this.data = this.data[8:]
	return val
}

/*
*
reads a uint16 table. The size of the table is indicated by the uint16 data at the beginning.
*/
func (this *ClassReader) readUint16s() []uint16 {
	n := this.readUint16()
	s := make([]uint16, n)
	for i := range s {
		s[i] = this.readUint16()
	}
	return s
}

/*
*
reads the specified number of bytes of length
*/
func (this *ClassReader) readBytes(length uint32) []byte {
	bytes := this.data[:length]
	this.data = this.data[length:]
	return bytes
}
