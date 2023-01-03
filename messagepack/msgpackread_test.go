//messagepackRaw_test.go
package messagepack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
83 A1 61 CB 3F F3 33 33 33 33 33 33 A1 62 CA 3F 99 99 9A A1 63 3
84 A1 61 CB 3F F3 33 33 33 33 33 33 A1 62 CA 3F 99 99 9A A1 63 3 A1 64 FD

86 A1 61 CB 3F F3 33 33 33 33 33 33 A1 62 CA 3F 99 99 9A A1 63 3 A1 64 FD A1 65 D1 F4 48 A1 66 CD F A0
*/

func TestIntReads(t *testing.T) {
	i, e := ReadInt(bytes.NewBuffer([]byte{0x1}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(1), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0xff}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(-1), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0x2}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(2), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0xfe}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(-2), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0x7f}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(127), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0xd0, 0xd4}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(-44), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0xcc, 0x96}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(150), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0xd1, 0xfe, 0x44}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(-444), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0xd2, 0xff, 0xff, 0x52, 0x64}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(-44444), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0xce, 0x00, 0x06, 0x1a, 0x80}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(400000), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0xd3, 0xff, 0xff, 0xff, 0xfe, 0xf7, 0x17, 0x28, 0xe4}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(-4444444444), i)

	i, e = ReadInt(bytes.NewBuffer([]byte{0xcf, 0x00, 0x00, 0x00, 0x02, 0xdf, 0xdc, 0x1c, 0x35}))
	assert.Equal(t, nil, e)
	assert.Equal(t, int64(12345678901), i)

}
