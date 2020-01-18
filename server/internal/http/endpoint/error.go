package endpoint

import "fmt"

type HandlerError struct {
  Status  int
  Message string
  Err     error
}

func (e *HandlerError) Error() string {
  if e.Err != nil {
    return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
  }

  return e.Message
}
