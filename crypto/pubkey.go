package crypto

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/likecoin-pro/likecoin/crypto/base58"
	"golang.org/x/crypto/sha3"
)

type PublicKey struct {
	x *big.Int
	y *big.Int
}

var (
	errPublicKeyDecode = errors.New("crypto-error: incorrect length of public key")
)

// Verify verifies the signature in r, s of hash using the public key, PublicKey. Its
// return value records whether the signature is valid.
func (pub *PublicKey) Verify(data []byte, sign []byte) bool {
	if pub.Empty() {
		return false
	}
	if len(sign) != PublicKeySize {
		return false
	}
	r := new(big.Int).SetBytes(sign[:KeySize])
	s := new(big.Int).SetBytes(sign[KeySize:])

	if r.Sign() == 0 || r.Cmp(curveParams.N) >= 0 {
		return false
	}
	if s.Sign() == 0 || s.Cmp(curveParams.N) >= 0 {
		return false
	}

	e := hashInt(data)
	w := new(big.Int).ModInverse(s, curveParams.N)

	u1 := e.Mul(e, w)
	u2 := w.Mul(r, w)

	u1.Mod(u1, curveParams.N)
	u2.Mod(u2, curveParams.N)

	x1, y1 := curve.ScalarBaseMult(u1.Bytes())
	x2, y2 := curve.ScalarMult(pub.x, pub.y, u2.Bytes())
	x, y := curve.Add(x1, y1, x2, y2)
	if x.Sign() == 0 && y.Sign() == 0 {
		return false
	}
	x.Mod(x, curveParams.N)
	return x.Cmp(r) == 0
}

func (pub *PublicKey) Empty() bool {
	return pub == nil || pub.x == nil && pub.y == nil
}

func (pub *PublicKey) String() string {
	return base58.Encode(pub.Encode())
}

func (pub *PublicKey) Equal(p *PublicKey) bool {
	return pub != nil && p != nil && pub.x.Cmp(p.x) == 0 && pub.y.Cmp(p.y) == 0
}

func (pub *PublicKey) Address() Address {
	h2 := sha256.New()
	h2.Write(intToBytes(pub.x))
	h2.Write(intToBytes(pub.y))
	h3 := sha3.Sum512(h2.Sum(nil))
	return newAddress(h3[:AddressSize])
}

func (pub *PublicKey) Encode() []byte {
	return append(
		intToBytes(pub.x),    // 32 bytes X
		intToBytes(pub.y)..., // 32 bytes Y
	)
}

func (pub *PublicKey) Decode(data []byte) error {
	if len(data) != PublicKeySize {
		return errPublicKeyDecode
	}
	pub.x = new(big.Int).SetBytes(data[:KeySize])
	pub.y = new(big.Int).SetBytes(data[KeySize:])
	return nil
}

func (pub *PublicKey) MarshalJSON() ([]byte, error) {
	return []byte(`"` + pub.String() + `"`), nil
}

func (pub *PublicKey) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if p, err := ParsePublicKey(str); err != nil {
		return err
	} else {
		pub.x = p.x
		pub.y = p.y
		return nil
	}
}

func MustParsePublicKey(pubkey string) *PublicKey {
	if pub, err := ParsePublicKey(pubkey); err != nil {
		panic(err)
	} else {
		return pub
	}
}

func ParsePublicKey(str64 string) (pub *PublicKey, err error) {
	data, err := base58.DecodeFixed(str64, PublicKeySize)
	if err != nil {
		return
	}
	return DecodePublicKey(data)
}

func DecodePublicKey(data []byte) (pub *PublicKey, err error) {
	pub = &PublicKey{}
	err = pub.Decode(data)
	return
}