package middleware

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type ctxKey struct {
  name string
}

func (k *ctxKey) String() string {
  return "neko/ctx/" + k.name
}
