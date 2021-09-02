package pionlog

type nulllog struct{}

func (l nulllog) Trace(msg string)                          {}
func (l nulllog) Tracef(format string, args ...interface{}) {}
func (l nulllog) Debug(msg string)                          {}
func (l nulllog) Debugf(format string, args ...interface{}) {}
func (l nulllog) Info(msg string)                           {}
func (l nulllog) Infof(format string, args ...interface{})  {}
func (l nulllog) Warn(msg string)                           {}
func (l nulllog) Warnf(format string, args ...interface{})  {}
func (l nulllog) Error(msg string)                          {}
func (l nulllog) Errorf(format string, args ...interface{}) {}
