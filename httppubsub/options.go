package httppubsub

// ServerOption is an option for the Server.
type ServerOption func(*Server)

// Logger is a minimal logging interface for httppubsub.
type Logger func(fmt string, v ...interface{})

// WithLogger adds a logger to the pubsub server.
func WithLogger(logger Logger) ServerOption {
	return func(s *Server) {
		s.log = logger
	}
}

// nopLogger returns a logger that does nothing.
func nopLogger() Logger {
	return func(_ string, _ ...interface{}) {}
}

// WithMaxRequestSize sets a limit on the max broadcast request size.
func WithMaxRequestSize(sizeInBytes int64) ServerOption {
	return func(s *Server) {
		s.maxRequestSize = sizeInBytes
	}
}
