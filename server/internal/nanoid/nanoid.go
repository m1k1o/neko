package nanoid

import (
	"math/rand"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
)

var nano *NanoID

func init() {
	nano = &NanoID{
		alphabet: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		size:     16,
	}
}

func New(alphabet string, size int) *NanoID {
	return &NanoID{
		alphabet: alphabet,
		size:     size,
	}
}

type NanoID struct {
	alphabet string
	size     int
}

func (n *NanoID) NewID() (string, error) {
	return gonanoid.Generate(n.alphabet, n.size)
}

func (n *NanoID) NewIDSize(size int) (string, error) {
	return gonanoid.Generate(n.alphabet, size)
}

func (n *NanoID) NewIDRang(max int, min int) (string, error) {
	rand.Seed(time.Now().Unix())
	return gonanoid.Generate(n.alphabet, rand.Intn(max-min)+min)
}

func (n *NanoID) GenerateID(alphabet string, size int) (string, error) {
	return gonanoid.Generate(alphabet, size)
}

func (n *NanoID) GenerateIDRange(alphabet string, max int, min int) (string, error) {
	rand.Seed(time.Now().Unix())
	return gonanoid.Generate(alphabet, rand.Intn(max-min)+min)
}

func NewID() (string, error) {
	return nano.NewID()
}

func NewIDSize(size int) (string, error) {
	return nano.NewIDSize(size)
}

func NewIDRang(max int, min int) (string, error) {
	return nano.NewIDRang(max, min)
}

func GenerateID(alphabet string, size int) (string, error) {
	return nano.GenerateID(alphabet, size)
}

func GenerateIDRange(alphabet string, max int, min int) (string, error) {
	return nano.GenerateIDRange(alphabet, max, min)
}
