package Code

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"strconv"
)

type SecretKey struct {
	n_ *big.Int
	d_ *big.Int
}

type PublicKey struct {
	n_ *big.Int
	e_ *big.Int
}

func (t *SecretKey) N() *big.Int {
	return t.n_
}

func (t *SecretKey) D() *big.Int {
	return t.d_
}

func (t *PublicKey) N() *big.Int {
	return t.n_
}

func (t *PublicKey) E() *big.Int {
	return t.e_
}

func (t *PublicKey) String() string {
	return "{n: " + t.n_.String() + ", e: " + t.e_.String() + "}"
}

func makePublicKey(n *big.Int, e *big.Int) *PublicKey {
	pk := new(PublicKey)
	pk.e_ = e
	pk.n_ = n
	return pk
}

func makePrivateKey(n *big.Int, d *big.Int) *SecretKey {
	sk := new(SecretKey)
	sk.n_ = n
	sk.d_ = d
	return sk
}

/* Returns a key of the length specified in the argument.
 */
func KeyGen(k int) (SecretKey, PublicKey) {
	k = k / 2
	p := big.NewInt(0)
	q := big.NewInt(0)
	n := big.NewInt(0)
	r := big.NewInt(0)

	determiningPrimes := true
	for determiningPrimes {
		//Primes p,q with the length of maxBitLengthOfPrimes
		p, _ = rand.Prime(rand.Reader, k)
		q, _ = rand.Prime(rand.Reader, k)

		// Ensure that p != q
		for p.Cmp(q) == 0 {
			q, _ = rand.Prime(rand.Reader, k)
		}
		n.Mul(p, q)
		// Checks if gcd((p-1)(q-1),3) = 1
		p.Sub(p, big.NewInt(1))
		q.Sub(q, big.NewInt(1))
		r = big.NewInt(0).Mul(p, q)
		if big.NewInt(0).GCD(big.NewInt(0), big.NewInt(0), r, big.NewInt(3)).Cmp(big.NewInt(1)) == 0 {
			determiningPrimes = false
		}
	}

	pk := makePublicKey(n, big.NewInt(3))
	sk := makePrivateKey(n, big.NewInt(0).ModInverse(big.NewInt(3), r))

	return *sk, *pk
}

func Sign(m string, sk SecretKey) string {
	hash := sha256.Sum256([]byte(m))
	z := big.NewInt(0)
	z.SetBytes(hash[:])
	return z.Exp(z, sk.D(), sk.N()).String()
}

func HashSHA(m string) string {
	hash := sha256.Sum256([]byte(m))
	z := big.NewInt(0)
	return z.SetBytes(hash[:]).String()
}

func SignTransaction(t Transaction, sk SecretKey) Transaction {
	m := buildStringToSign(t)
	s := Sign(m, sk)
	t.Signature = s
	return t
}

func buildStringToSign(t Transaction) string {
	var buf bytes.Buffer
	buf.WriteString(t.From.String())
	buf.WriteString(t.To.String())
	buf.WriteString(strconv.Itoa(t.Amount))
	buf.WriteString(t.ID)
	return buf.String()
}

func verify(m string, s string, pk PublicKey) bool {
	sInt := big.NewInt(0)
	sInt.SetString(s, 10)
	hash := sha256.Sum256([]byte(m))
	hashedMessage := big.NewInt(0)
	hashedMessage.SetBytes(hash[:])
	toVerify := sInt.Exp(sInt, pk.E(), pk.N())

	if hashedMessage.Cmp(toVerify) == 0 {
		return true
	}
	return false
}

func VerifyTransaction(t Transaction, pk PublicKey) bool {
	return verify(buildStringToSign(t), t.Signature, pk)
}
