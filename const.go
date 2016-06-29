package sftps

const (
	NONE int = 0
)

const (
	ONLINE  int = 1
	OFFLINE int = 2
)
const (
	FTP  int = 1
	FTPS int = 2
	SFTP int = 3
)
const (
	DOWNLOAD int = 1
	UPLOAD   int = 2
)
const (
	IMPLICIT int = 1
	EXPLICIT int = 2
)

const (
	// When handshake to the server
	TIMEOUT string = "10s"
	// The Keep Alive Period for an active network connection.
	KEEPALIVE string = "30s"
)
