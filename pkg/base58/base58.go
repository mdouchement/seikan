package base58

import (
	"fmt"
	"math/big"
)

// base58 alphabet used by bitcoin.
const base58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Encode encodes the given p in base58.
func Encode(p []byte) string {
	bn0 := big.NewInt(0)
	bn58 := big.NewInt(58)

	zero := base58[0]
	idx := len(p)*138/100 + 1
	buf := make([]byte, idx)
	bn := new(big.Int).SetBytes(p)
	var mo *big.Int

	for bn.Cmp(bn0) != 0 {
		bn, mo = bn.DivMod(bn, bn58, new(big.Int))
		idx--
		buf[idx] = base58[mo.Int64()]
	}

	for i := range p {
		if p[i] != 0 {
			break
		}
		idx--
		buf[idx] = zero
	}
	return string(buf[idx:])
}

// Decode decodes the given base58 s into bytes.
func Decode(s string) []byte {
	decode := make([]int8, 128)
	for i := range decode {
		decode[i] = -1
	}
	for i := 0; i < 58; i++ {
		decode[base58[i]] = int8(i)
	}

	//

	bn58 := big.NewInt(58)
	zero := base58[0]

	var zcount int
	for i := 0; i < len(s) && s[i] == zero; i++ {
		zcount++
	}
	leading := make([]byte, zcount)

	var padChar rune = -1
	src := []byte(s)
	j := 0
	for ; j < len(s) && src[j] == byte(padChar); j++ {
	}

	n := new(big.Int)
	for i := range src[j:] {
		c := decode[src[i]]
		if c == -1 {
			panic(fmt.Errorf("illegal base58 data at input index: %d", i))
		}
		n.Mul(n, bn58)
		n.Add(n, big.NewInt(int64(c)))
	}
	return append(leading, n.Bytes()...)
}
