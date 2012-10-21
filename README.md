go-gelf - GELF library and writer for Go
========================================

GELF is greylog2's UDP logging format.  This library provides an API
that applications can use to log messages directly to a greylog2
server, along with an io.Writer that can be use to redirect the
standard library's log messages (or os.Stdout), to a greylog2 server.
