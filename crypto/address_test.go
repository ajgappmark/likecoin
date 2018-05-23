package crypto

import (
	"strings"
	"testing"

	"github.com/likecoin-pro/likecoin/crypto/base58"
	"github.com/stretchr/testify/assert"
)

func TestAddress_String(t *testing.T) {
	for i := 1; i < 1e3; i++ {
		addr := newAddress(randBytes(AddressLength))

		sAddr := addr.String()

		assert.True(t, strings.HasPrefix(sAddr, "Like"))
		assert.Equal(t, 43, len(sAddr))
	}
}

func TestAddress_isValidBase58(t *testing.T) {
	strAddr := newAddress(randBytes(AddressLength)).String()

	_, err := base58.Decode(strAddr)

	assert.NoError(t, err)
}

func TestAddress_IsNil(t *testing.T) {
	var addr Address

	isNil := addr.IsNil()

	assert.True(t, isNil)
}

func TestAddress_TaggedString(t *testing.T) {
	addr := newAddress(randBytes(AddressLength))

	sAddr := addr.TaggedString(666)

	assert.True(t, strings.HasPrefix(sAddr, "Like"))
	assert.True(t, len(sAddr) > 43)
}

func TestParseAddress(t *testing.T) {
	randData := randBytes(AddressLength)
	strAddr := newAddress(randData).String()

	addr, tag, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.Equal(t, randData, addr[:])
	assert.EqualValues(t, 0, tag)
}

func TestParseAddress_withTag(t *testing.T) {
	strAddr := newAddress(randBytes(AddressLength)).TaggedString(0x19720000abba0000)

	_, tag, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.EqualValues(t, 0x19720000abba0000, tag)
}

func TestParseAddress_withTag2(t *testing.T) {
	for i := uint64(1); i < 1e3; i++ {
		strAddr := newAddress(randBytes(AddressLength)).TaggedString(i)

		_, tag, err := ParseAddress(strAddr)

		assert.NoError(t, err)
		assert.Equal(t, i, tag)
	}
}

func TestParseAddress_withSimpleTagSuffix(t *testing.T) {
	strAddr := "Like5DuaVTk8KgpRh98xDvHvnpaAWxSoYh6uLRvyar50666"

	addr, tag, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.Equal(t, "Like5DuaVTk8KgpRh98xDvHvnpaAWxSoYh6uLRvyar5", addr.String())
	assert.EqualValues(t, 0x666, tag)
}

func TestParseAddress_fail(t *testing.T) {
	strAddr := newAddress(randBytes(AddressLength)).String()
	strAddr += "a"

	_, _, err := ParseAddress(strAddr)

	assert.Error(t, err)
}

func TestAddress_Hex(t *testing.T) {
	addr := MustParseAddress("Like5DuaVTk8KgpRh98xDvHvnpaAWxSoYh6uLRvyar5")

	s := addr.Hex()

	assert.Equal(t, "0x9a8a9d2b5766b5c3962f4dd301c01765bdc37a6387f24250", s)
}

func TestParseAddress_hex(t *testing.T) {
	addr, tag, err := ParseAddress("0x9a8a9d2b5766b5c3962f4dd301c01765bdc37a6387f24250")

	assert.NoError(t, err)
	assert.Equal(t, "Like5DuaVTk8KgpRh98xDvHvnpaAWxSoYh6uLRvyar5", addr.String())
	assert.EqualValues(t, 0, tag)
}

func TestParseAddress_hex_tag(t *testing.T) {
	addr, tag, err := ParseAddress("0x9a8a9d2b5766b5c3962f4dd301c01765bdc37a6387f242500123456789abcdef")

	assert.NoError(t, err)
	assert.Equal(t, "Like5DuaVTk8KgpRh98xDvHvnpaAWxSoYh6uLRvyar5", addr.String())
	assert.EqualValues(t, 0x0123456789abcdef, tag)
}

func TestAddress_Encode_nilAddress(t *testing.T) {
	var addr Address

	data := addr.Encode()

	assert.Equal(t, 0, len(data))
}

func TestAddress_Decode(t *testing.T) {
	a := newAddress(randBytes(AddressLength))
	data := a.Encode()

	var b Address
	err := b.Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestAddress_Decode_nilAddress(t *testing.T) {
	var addr Address
	err := addr.Decode(nil)

	assert.NoError(t, err)
	assert.True(t, addr.IsNil())
}
