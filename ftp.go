package sftps

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	//"github.com/davecgh/go-spew/spew"
)

type FtpParameters struct {
	Host        string
	Port        int
	ListenPort  int
	User        string
	Pass        string
	Passive     bool
	KeepAlive   bool
	Secure      bool
	AlwaysTrust bool
	SecureMode  int
	RootCA      string
	Cert        string
	Key         string
}

type Ftp struct {
	rawConn  net.Conn
	tlsConn  *tls.Conn
	ctrlConn *textproto.Conn
	params   *FtpParameters
	State    int
}

func NewFtp(p *FtpParameters) (ftp *Ftp) {

	ftp = new(Ftp)
	ftp.params = p
	ftp.State = OFFLINE
	return
}

func (this *Ftp) connect() (res *Response, err error) {
	var ipaddr []net.IP
	var code int
	var msg string

	if ipaddr, err = net.LookupIP(this.params.Host); err != nil {
		return
	}

	addr := fmt.Sprintf("%s:%d", ipaddr[0], this.params.Port)

	dialer := new(net.Dialer)
	if dialer.Timeout, err = time.ParseDuration(TIMEOUT); err != nil {
		return
	}
	if dialer.KeepAlive, err = time.ParseDuration(KEEPALIVE); err != nil {
		return
	}
	this.rawConn, err = dialer.Dial("tcp", addr)
	if err != nil {

		return
	}
	this.ctrlConn = textproto.NewConn(this.rawConn)
	if code, msg, err = this.ctrlConn.ReadResponse(220); err != nil {

		return
	}

	res = &Response{
		command: "",
		code:    code,
		msg:     msg,
	}

	if this.params.Secure && this.params.SecureMode == IMPLICIT {
		if err = this.secureUpgrade(); err != nil {
			return
		}
	}

	this.State = ONLINE
	return
}

func (this *Ftp) secureUpgrade() (err error) {
	var conf *tls.Config
	if conf, err = this.getTLSConfig(); err != nil {
		return
	}
	this.tlsConn = tls.Client(this.rawConn, conf)
	this.ctrlConn = textproto.NewConn(this.tlsConn)
	return
}

func (this *Ftp) getTLSConfig() (conf *tls.Config, err error) {
	var certPair tls.Certificate
	var certPool *x509.CertPool
	var rcaPem []byte

	conf = new(tls.Config)
	conf.ClientAuth = tls.VerifyClientCertIfGiven
	conf.CipherSuites = []uint16{
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}

	if this.params.Cert != "" && this.params.Key != "" {
		if certPair, err = tls.LoadX509KeyPair(this.params.Cert, this.params.Key); err != nil {
			return
		}

		certPool = x509.NewCertPool()

		if this.params.RootCA != "" {
			if rcaPem, err = ioutil.ReadFile("./cert/rcaPem.pem"); err != nil {
				return
			}

			if this.params.AlwaysTrust {
				if !certPool.AppendCertsFromPEM(rcaPem) {
					panic("Failed to parse the Root Certificate")
				}
			}
			conf.RootCAs = certPool
		}

		conf.Certificates = make([]tls.Certificate, 1)
		conf.Certificates[0] = certPair
		conf.ClientCAs = certPool
	}
	conf.InsecureSkipVerify = this.params.AlwaysTrust
	return
}

func (this *Ftp) auth() (res []*Response, err error) {

	var r *Response

	res = []*Response{}

	if this.params.Secure && this.params.SecureMode == EXPLICIT {
		if r, err = this.Command("AUTH TLS", 234); err != nil {
			return
		}
		res = append(res, r)

		if err = this.secureUpgrade(); err != nil {
			return
		}
	}

	if r, err = this.Command(fmt.Sprintf("USER %s", this.params.User), 331); err != nil {
		return
	}
	res = append(res, r)

	if r, err = this.Command(fmt.Sprintf("PASS %s", this.params.Pass), 230); err != nil {
		return
	}
	res = append(res, r)

	return
}


func (this *Ftp) Command(cmd string, code int) (res *Response, err error) {
	var c int
	var m string

	if _, err = this.ctrlConn.Cmd(cmd); err != nil {
		return
	}

	if c, m, err = this.ctrlConn.ReadResponse(code); err != nil {
		return
	}

	res = &Response{
		command: cmd,
		code:    c,
		msg:     m,
	}
	return
}

func (this *Ftp) options() (res []*Response, err error) {
	var r *Response
	if r, err = this.Command("SYST", 215); err != nil {
		return
	}
	res = []*Response{}
	res = append(res, r)

	if r, err = this.Command("FEAT", 211); err != nil {
		return
	}
	res = append(res, r)

	if r, err = this.Command("OPTS UTF8 ON", 200); err != nil {
		return
	}
	res = append(res, r)

	if this.params.Secure {
		if r, err = this.Command("PROT P", 200); err != nil {
			return
		}
		res = append(res, r)
	}
	if r, err = this.Command("TYPE I", 200); err != nil {
		return
	}
	res = append(res, r)
	return
}


func (this *Ftp) getLocalIP() (ip string, err error) {
	var addrs []net.Addr
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.To4().String()
			}
		}
	}
	if ip == "" {
		err = errors.New("Could not get the Local Address.")
		return
	}
	return
}

func (this *Ftp) ds2h(dec string) (hex string, err error) {
	var p int = 0
	if p, err = strconv.Atoi(dec); err != nil {
		return
	}
	hex = fmt.Sprintf("%x", p)
	return
}

func (this *Ftp) h2i(hex string) (res int, err error) {
	var r int64
	r, err = strconv.ParseInt(hex, 16, 64)
	res = int(r)
	return
}

func (this *Ftp) getSplitPorts() (port1 int, port2 int, err error) {
	hex := fmt.Sprintf("%x", this.params.Port)

	switch len(hex) {
	case 1, 2:
		port1 = 0
		if port2, err = this.h2i(hex); err != nil {
			return
		}
	case 3:
		p1 := hex[:1]
		p2 := hex[1:3]
		if port1, err = this.h2i(p1); err != nil {
			return
		}
		if port2, err = this.h2i(p2); err != nil {
			return
		}
	case 4:
		p1 := hex[:2]
		p2 := hex[2:4]
		if port1, err = this.h2i(p1); err != nil {
			return
		}
		if port2, err = this.h2i(p2); err != nil {
			return
		}
	default:
		err = errors.New("The Port Number could not converted to the Parameter Format for the Listen Function.")
	}
	return
}

func (this *Ftp) Port() (res *Response, dataConn net.Conn, err error) {
	var localIP string = ""
	if localIP, err = this.getLocalIP(); err != nil {
		return
	}
	ip := strings.Replace(localIP, ".", ",", -1)
	var p1, p2 int
	if p1, p2, err = this.getSplitPorts(); err != nil {
		return
	}
	cmd := fmt.Sprintf("PORT %s,%d,%d", ip, p1, p2)

	if res, err = this.Command(cmd, 200); err != nil {
		return
	}

	listener, e := net.Listen("tcp", fmt.Sprintf("%s:%d", localIP, this.params.ListenPort))
	defer listener.Close()

	if e != nil {
		return
	}
	if dataConn, err = listener.Accept(); err != nil {
		return
	}
	return
}

func (this *Ftp) Pasv() (res *Response, dataConn net.Conn, err error) {
	if res, err = this.Command("PASV", 227); err != nil {
		return
	}
	var ip []net.IP
	if ip, err = net.LookupIP(this.params.Host); err != nil {
		return
	}
	reg := regexp.MustCompile("([0-9]+?),([0-9]+?),([0-9]+?),([0-9]+?),([0-9]+?),([0-9]+)")
	matches := reg.FindAllStringSubmatch(res.msg, -1)
	tmp := matches[0]

	var hex1 string = ""
	var hex2 string = ""
	if hex1, err = this.ds2h(tmp[5]); err != nil {
		return
	}
	if hex2, err = this.ds2h(tmp[6]); err != nil {
		return
	}
	var port int
	if port, err = this.h2i(fmt.Sprintf("%s%s", hex1, hex2)); err != nil {
		return
	}
	param := fmt.Sprintf("%s:%d", ip[0], port)
	dataConn, err = net.Dial("tcp", param)
	return
}


func (this *Ftp) readBytes(dataConn net.Conn) (res *Response, bytes []byte, err error) {
	defer dataConn.Close()
	if this.params.Secure {
		var conf *tls.Config
		if conf, err = this.getTLSConfig(); err != nil {
			return
		}
		dataTLS := tls.Client(dataConn, conf)
		defer dataTLS.Close()

		if bytes, err = ioutil.ReadAll(dataTLS); err != nil {
			return
		}
		dataTLS.Close() // Important the Buffer flush out.

	} else {
		if bytes, err = ioutil.ReadAll(dataConn); err != nil {
			return
		}
		dataConn.Close() // Important the Buffer flush out.
	}

	c, m, e := this.ctrlConn.ReadResponse(226)
	if e != nil {
		err = e
		return
	}
	res = &Response{
		command: "",
		code:    c,
		msg:     m,
	}
	return
}

func (this *Ftp) quit() (res *Response, err error) {
	defer this.ctrlConn.Close()
	defer this.tlsConn.Close()
	defer this.rawConn.Close()

	if res, err = this.Command("QUIT", 221); err != nil {
		return
	}
	return
}

func (this *Ftp) getDataConn() (res *Response, dataConn net.Conn, err error) {
	if this.params.Passive {
		if res, dataConn, err = this.Pasv(); err != nil {
			return
		}

	} else {
		if res, dataConn, err = this.Port(); err != nil {
			return
		}
	}

	return
}

func (this *Ftp) list(p string) (res []*Response, list string, err error) {
	if !this.params.KeepAlive {
		defer this.quit()
	}

	var dataConn net.Conn
	var bytes []byte
	var r *Response
	res = []*Response{}

	cmd := fmt.Sprintf("LIST -aL %s", p)

	if r, dataConn, err = this.getDataConn(); err != nil {
		return
	}
	defer dataConn.Close()
	res = append(res, r)


	if r, err = this.Command(cmd, 150); err != nil {
		return
	}
	res = append(res, r)

	if r, bytes, err = this.readBytes(dataConn); err != nil {
		return
	}
	res = append(res, r)

	list = string(bytes)

	return
}

func (this *Ftp) fileTransfer(direction int, uri string, itf interface{}) (res *Response, len int64, err error) {
	var dataConn net.Conn
	var file *os.File
	defer dataConn.Close()
	defer file.Close()

	if this.params.Passive {
		if c, ok := itf.(net.Conn); ok {
			dataConn = c
		} else {
			err = errors.New("Invalid parameter were bound, Value of the argument 'itf' must be the Type 'net.Conn' when the Passive Mode specified by the Parameter.")
			return
		}
	} else {
		if listener, ok := itf.(net.Listener); ok {
			defer listener.Close()
			if dataConn, err = listener.Accept(); err != nil {
				return
			}
		} else {
			err = errors.New("Invalid parameter were bound, Value of the argument 'itf' must be the Type 'net.Listener' whern the Active Mode speciffied by the Parameter")
			return
		}
	}

	var r io.ReadCloser = file
	var w io.WriteCloser = file
	var rw io.ReadWriteCloser = dataConn

	if this.params.Secure {
		var conf *tls.Config
		if conf, err = this.getTLSConfig(); err != nil {
			return
		}
		dataTLS := tls.Client(dataConn, conf)
		defer dataTLS.Close()
		rw = dataTLS
	}

	if direction == DOWNLOAD {
		r = rw
	} else if direction == UPLOAD {
		w = rw
	} else {
		err = errors.New("The Argument 'direction' must be the either 'DOWNLOAD' or 'UPLOAD'.")
		return
	}

	if len, err = io.Copy(w, r); err != nil {
		return
	}
	var code int
	var msg string
	if code, msg, err = this.ctrlConn.ReadResponse(226); err != nil {
		return
	}
	res = &Response{
		command: "",
		code:    code,
		msg:     msg,
	}
	return
}


