package messagepack

import (
	"bytes"
	"testing"
)

func TestWrites(t *testing.T) {

	b := new(bytes.Buffer)
	WriteFixmap(b, 6)

	WriteString(b, "a") //1
	writeFloat64(b, 1.2)

	WriteString(b, "b") //2
	writeFloat32(b, 1.2)

	WriteString(b, "c") //3
	WriteInt(b, 3)

	WriteString(b, "d") //4
	WriteInt(b, -3)

	WriteString(b, "e") //5
	WriteInt(b, -3000)

	WriteString(b, "f") //6
	WriteInt(b, 4000)

	/*
		fmt.Printf("TESTWRITE\n")
		code := b.Bytes()
		for _, v := range code {
			fmt.Printf("%X ", v)
		}
		fmt.Printf("\n")*/
}
