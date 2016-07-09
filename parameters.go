package sftps


type ftpParameters struct {
	host        string
	port        int
	listenPort  int
	user        string
	pass        string
	passive     bool
	keepAlive   bool
	secure      bool
	alwaysTrust bool
	secureMode  int
	rootCA      string
	cert        string
	key         string
}

type sftpParameters struct {
	host          string
	port          int
	user          string
	pass          string
	useKey        bool
	privateKey    string
	usePassphrase bool
	passphrase    string
	keepAlive     bool
}

func NewSftpParameters(host string, port int, user string, pass string, keepAlive bool) *sftpParameters {
	if host == "" || user == "" {
		panic("Invalid parameter were bound.")
	}
	param := &sftpParameters {
		host: host,
		port: port,
		user: user,
		pass: pass,
		useKey:   false,
		privateKey:  "",
		usePassphrase: false,
		passphrase:  "",
		keepAlive:  keepAlive,
	}
	return param
}

func (param *sftpParameters) Keys(privateKey string, usePassphrase bool, passphrase string) {
	param.useKey = true
	if usePassphrase {
		if passphrase == "" {
			panic("The passphrase must not be empty when specified true to usePassphrase.")
		}
		param.usePassphrase = true
		param.passphrase = passphrase
	}
}


func NewFtpParameters(host string, port int, user string, pass string, keepalive bool) *ftpParameters {
	if host == "" || user == "" || pass == "" {
		panic("Invalid parameter were bound.")
	}
	param := &ftpParameters{
		host:        host,
		port:        port,
		listenPort:  0,
		user:        user,
		pass:        pass,
		passive:     true,
		keepAlive:   keepalive,
		secure:      false,
		alwaysTrust: false,
		secureMode:  EXPLICIT,
		rootCA:      "",
		cert:        "",
		key:         "",
	}
	return param
}
func (param *ftpParameters) ActiveMode(actvPort int) {
	param.passive = false
	param.listenPort = actvPort
}
func (param *ftpParameters) Secure(skipVerify bool) {
	param.secure = true
	param.alwaysTrust = skipVerify
}
func (param *ftpParameters) Certs(rca string, cert string, key string) {
	param.secure = true
	param.rootCA = rca
	param.cert = cert
	param.key = key
}

func (param *ftpParameters) Implicit(port int) {
	param.secure = true
	param.secureMode = IMPLICIT
	if port == 0 {
		param.port = 990
	} else {
		param.port = port
	}
}
