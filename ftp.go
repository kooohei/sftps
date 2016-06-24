package sftps

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/textproto"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// FTP
// Structure of the this Module.
type FTP struct {
	rawConn  net.Conn
	ctrlConn *textproto.Conn
	tlsConn  *tls.Conn
	params   *FtpParameter
	isDebug  bool
}

/*
	NewFTP is the Factory Metrhod, Create instance for the FTP Connection.
*/
func NewFTP(parameter *FtpParameter, debug bool) (ftp *FTP) {
	ftp = new(FTP)
	ftp.params = parameter
	ftp.isDebug = debug
	return
}

func (this *FTP) Call(m map[string]interface{}, name string, params ...interface{}) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	f.Call(in)
	return
}

/*
	Auth()
	Authentication process
*/
func (this *FTP) Auth() {
	if this.params.Secure && this.params.SecureMode == "explicit" {
		this.Command("AUTH TLS", 234)
		this.SecureUpgrade()
	}
	this.Command(fmt.Sprintf("USER %s", this.params.User), 331)
	if this.params.Pass != "" {
		this.Command(fmt.Sprintf("PASS %s", this.params.Pass), 230)
	}
}

/*
	Connect
	Connected to the server
*/
func (this *FTP) Connect() {
	ipaddr, err := net.LookupIP(this.params.Host)
	Err("Connect", err, nil)
	addr := fmt.Sprintf("%s:%d", ipaddr, this.params.Port)

	this.rawConn, err = net.Dial("tcp", addr)
	Err("Connect", err, nil)
	this.ctrlConn = textproto.NewConn(this.rawConn)
	code, msg, err := this.ctrlConn.ReadResponse(220)
	Err("Connect", err, this.CloseAll)

	if this.params.Secure && this.params.SecureMode == "implicit" {
		this.SecureUpgrade()
	}

	if this.isDebug {
		log.Printf("[RESPONSE] %d %s", code, msg)
	}
	return
}

func (this *FTP) CloseAll() {
	if this.ctrlConn != nil {
		this.ctrlConn.Close()
	}
	if this.tlsConn != nil {
		this.tlsConn.Close()
	}
	if this.rawConn != nil {
		this.rawConn.Close()
	}
}

/*
	Upgrade the Secure Control connection with TLS
*/
func (this *FTP) SecureUpgrade() {
	conf := this.GetTLSConfig()
	this.tlsConn = tls.Client(this.rawConn, conf)
	this.ctrlConn = textproto.NewConn(this.tlsConn)
}

/*
	GetTLSConfig
	Get Configuration structure for the TLS Connection.
*/
func (this *FTP) GetTLSConfig() (conf *tls.Config) {

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
	conf.InsecureSkipVerify = this.params.AlwaysTrust
	if this.params.Cert != "" && this.params.Key != "" {
		rootPEM, err := ioutil.ReadFile("./cert/bundle.crt")
		Err("", err, this.CloseAll)
		certPair, err := tls.LoadX509KeyPair(this.params.Cert, this.params.Key)
		Err("", err, this.CloseAll)
		certPool := x509.NewCertPool()
		
		if this.params.AlwaysTrust {
			if !certPool.AppendCertsFromPEM(rootPEM) {
				panic("Failed to parse root certificate")
			}
		}
		conf.Certificates = make([]tls.Certificate, 1)
		conf.Certificates[0] = certPair
		conf.RootCAs = certPool
		conf.ClientCAs = certPool
	}

	return
}

/*
	Command
	Request FTP Command to the server then it will get response with code.
*/
func (this *FTP) Command(cmd string, rc int) {
	if this.isDebug {
		log.Printf("[REQUEST COMMAND] %s\n", cmd)
	}

	ary := strings.Split(cmd, " ")
	Command := ary[0]

	_, err := this.ctrlConn.Cmd(cmd)
	Err(Command, err, this.CloseAll)
	code, msg, err := this.ctrlConn.ReadResponse(rc)
	Err(Command, err, this.CloseAll)
	if this.isDebug {
		log.Printf("[RESPONSE] %d %s\n", code, msg)
	}
	return
}

/*
	BaseCommands
	Execute implicit commands. i,e that is between after authenticate and before user specify command.
*/
func (this *FTP) BaseCommands() {
	this.Command("SYST", 215)
	this.Command("FEAT", 211)
	this.Command("OPTS UTF8 ON", 200)
	if this.params.Secure {
		this.Command("PROT P", 200)
	}
	this.Command("TYPE I", 200)
}

/*
	Port
	"PORT 123,123,123,123,12,34", this command would send port number
*/
func (this *FTP) Port() (listener net.Listener) {
	localIP, err := GetLocalIP()

	ip := strings.Replace(localIP, ".", ",", -1)
	Err("Port", err, this.CloseAll)
	port1, port2 := GetListenSplitPort(this.params.ListenPort)
	_, err = this.ctrlConn.Cmd("PORT %s,%d,%d", ip, port1, port2)
	Err("Port", err, this.CloseAll)
	code, msg, err := this.ctrlConn.ReadResponse(200)
	Err("Port", err, this.CloseAll)

	if this.isDebug {
		log.Printf("[RESPONSE] %d %s\n", code, msg)
	}
	listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", localIP, this.params.ListenPort))
	Err("Port", err, this.CloseAll)
	return
}

/*

 */
func (this *FTP) ActiveReadBytes(listener net.Listener) (bytes []byte) {
	defer listener.Close()

	dataConn, err := listener.Accept()
	Err("", err, this.CloseAll)
	defer dataConn.Close()
	listener.Close()

	if this.params.Secure {
		dataTLS := tls.Client(dataConn, this.GetTLSConfig())
		defer dataTLS.Close()
		bytes, err = ioutil.ReadAll(dataTLS)
		Err("", err, this.CloseAll)
		dataTLS.Close()
	} else {
		bytes, err = ioutil.ReadAll(dataConn)
		Err("", err, this.CloseAll)
	}
	dataConn.Close()
	code, msg, err := this.ctrlConn.ReadResponse(226)
	Err("", err, this.CloseAll)
	if this.isDebug {
		log.Printf("[RESPONSE] %d %s", code, msg)
	}
	return
}

/*

 */
func (this *FTP) FileToActiveConn(filePath string, listener net.Listener) {
	defer listener.Close()
	dataConn, err := listener.Accept()
	Err("", err, this.CloseAll)
	defer dataConn.Close()
	listener.Close()

	file, err := os.Open(filePath)
	Err("", err, this.CloseAll)
	defer file.Close()

	var writeLen int64
	if this.params.Secure {
		dataTLS := tls.Client(dataConn, this.GetTLSConfig())
		defer dataTLS.Close()
		writeLen, err = io.Copy(dataTLS, file)
		Err("", err, this.CloseAll)
	} else {
		writeLen, err = io.Copy(dataConn, file)
		Err("", err, this.CloseAll)
	}
	dataConn.Close()
	code, msg, err := this.ctrlConn.ReadResponse(226)
	Err("", err, this.CloseAll)
	if this.isDebug {
		log.Printf("[RESPONSE] %d %s", code, msg)
		log.Printf("[DATA TRANSFER] COMPLETE %d bytes", writeLen)
	}
	return
}

/*

 */
func (this *FTP) ActiveConnToFile(filePath string, listener net.Listener) {
	defer listener.Close()
	dataConn, err := listener.Accept()
	Err("DownloadFile", err, this.CloseAll)
	defer dataConn.Close()
	listener.Close()

	file, err := os.Create(filePath)
	Err("DownloadFile", err, this.CloseAll)
	defer file.Close()

	var writeLen int64
	if this.params.Secure {
		dataTLS := tls.Client(dataConn, this.GetTLSConfig())
		defer dataTLS.Close()
		writeLen, err = io.Copy(file, dataTLS)
		Err("DownloadFile", err, this.CloseAll)
	} else {
		writeLen, err = io.Copy(file, dataConn)
		Err("DownloadFile", err, this.CloseAll)
	}
	dataConn.Close()
	code, msg, err := this.ctrlConn.ReadResponse(226)
	Err("DownloadFile", err, this.CloseAll)
	if this.isDebug {
		log.Printf("[RESPONSE] %d %s", code, msg)
		log.Printf("[DATA TRANSFER] COMPLETE %d bytes", writeLen)
	}
	return
}

/*
	Pasv
	Execute user decided command after the PASV command.
	the result from server via the data connection.
*/
func (this *FTP) Pasv() (dataConn net.Conn) {
	_, err := this.ctrlConn.Cmd("PASV")
	Err("Pasv", err, this.CloseAll)
	code, msg, err := this.ctrlConn.ReadResponse(227)
	Err("Pasv", err, this.CloseAll)

	if this.isDebug {
		log.Printf("[RESPONSE] %d %s\n", code, msg)
	}

	ip, err := net.LookupIP(this.params.Host)
	Err("Pasv", err, this.CloseAll)
	reg := regexp.MustCompile("([0-9]+?),([0-9]+?),([0-9]+?),([0-9]+?),([0-9]+?),([0-9]+)")
	res := reg.FindAllStringSubmatch(msg, -1)
	tmp := res[0]

	hex1 := DecString2HexString(tmp[5])
	hex2 := DecString2HexString(tmp[6])
	port := HexString2Int(fmt.Sprintf("%s%s", hex1, hex2))
	param := fmt.Sprintf("%s:%d", ip[0], port)
	dataConn, err = net.Dial("tcp", param)
	Err("Pasv", err, this.CloseAll)

	return dataConn
}

/*
	PasvReadBytes
	Read data of the []byte type from data connection with Passive mode.
*/
func (this *FTP) PasvReadBytes(dataConn net.Conn) (bytes []byte) {
	var err error
	if this.params.Secure {
		dataTLS := tls.Client(dataConn, this.GetTLSConfig())
		defer dataTLS.Close()
		bytes, err = ioutil.ReadAll(dataTLS)
		Err("", err, this.CloseAll)
		dataTLS.Close()
	} else {
		bytes, err = ioutil.ReadAll(dataConn)
		Err("", err, this.CloseAll)
	}

	dataConn.Close()
	code, msg, err := this.ctrlConn.ReadResponse(226)
	Err("", err, this.CloseAll)
	if this.isDebug {
		log.Printf("[RESPONSE] %d %s", code, msg)
	}
	return
}

/*

 */
func (this *FTP) FileToPasvConn(filePath string, dataConn net.Conn) {
	defer dataConn.Close()

	file, err := os.Open(filePath)
	Err("UploadFile", err, this.CloseAll)
	defer file.Close()

	var writeLen int64
	if this.params.Secure {
		dataTLS := tls.Client(dataConn, this.GetTLSConfig())
		defer dataTLS.Close()
		writeLen, err = io.Copy(dataTLS, file)
		Err("UploadFile", err, this.CloseAll)
	} else {
		writeLen, err = io.Copy(dataConn, file)
		Err("UploadFile", err, this.CloseAll)
	}

	dataConn.Close()
	code, msg, err := this.ctrlConn.ReadResponse(226)
	Err("UploadFile", err, this.CloseAll)

	if this.isDebug {
		log.Printf("[RESPONSE] %d %s", code, msg)
		log.Printf("[DATA TRANSFER] COMPLETE %d bytes", writeLen)
	}
	return
}

/*

 */
func (this *FTP) PasvConnToFile(filePath string, dataConn net.Conn) {
	defer dataConn.Close()

	file, err := os.Create(filePath)
	Err("", err, this.CloseAll)
	defer file.Close()

	var writeLen int64
	if this.params.Secure {
		dataTLS := tls.Client(dataConn, this.GetTLSConfig())
		defer dataTLS.Close()
		writeLen, err = io.Copy(file, dataTLS)
		Err("DownloadFile", err, this.CloseAll)
	} else {
		writeLen, err = io.Copy(file, dataConn)
		Err("DownloadFile", err, this.CloseAll)
	}

	dataConn.Close()
	code, msg, err := this.ctrlConn.ReadResponse(226)
	Err("DownloadFile", err, this.CloseAll)

	if this.isDebug {
		log.Printf("[RESPONSE] %d %s", code, msg)
		log.Printf("[DATA TRANSFER] COMPLETE %d bytes", writeLen)
	}
	return
}

/*
	GetList
	Get file list from passed directory path
*/
func (this *FTP) GetList(path string) string {
	var bytes []byte
	cmd := fmt.Sprintf("LIST -aL %s", path)
	if this.params.Passive {
		dataConn := this.Pasv()
		this.Command(cmd, 150)
		bytes = this.PasvReadBytes(dataConn)
		if this.isDebug {
			log.Printf("[RESPONSE DATA] %v", string(bytes))
		}
	} else {
		listener := this.Port()
		this.Command(cmd, 150)
		bytes = this.ActiveReadBytes(listener)
		if this.isDebug {
			log.Printf("[RESPONSE DATA] %v", string(bytes))
		}
	}
	return string(bytes)
}

/*
	UploadFile
*/
func (this *FTP) UploadFile(command *Command) {
	localPath := command.Src
	remotePath := command.Dest

	cmd := fmt.Sprintf("STOR %s", remotePath)
	if this.params.Passive {
		dataConn := this.Pasv()
		this.Command(cmd, 150)
		this.FileToPasvConn(localPath, dataConn)
	} else {
		listener := this.Port()
		this.Command(cmd, 150)
		this.FileToActiveConn(localPath, listener)
	}
	Last("UploadFile", "OK", this.CloseAll)
}

/*

 */
func (this *FTP) DownloadFile(command *Command) {
	localPath := command.Src
	remotePath := command.Dest

	cmd := fmt.Sprintf("RETR %s", remotePath)
	if this.params.Passive {
		dataConn := this.Pasv()
		this.Command(cmd, 150)
		this.PasvConnToFile(localPath, dataConn)
	} else {
		listener := this.Port()
		this.Command(cmd, 150)
		this.ActiveConnToFile(localPath, listener)
	}
	Last("DownloadFile", "OK", this.CloseAll)
}

/*

 */
func (this *FTP) Rename(command *Command) {
	oldName := command.Src
	newName := command.Dest

	fr := fmt.Sprintf("RNFR %s", oldName)
	to := fmt.Sprintf("RNTO %s", newName)

	this.Command(fr, 350)
	this.Command(to, 250)

	Last("Rename", "OK", this.CloseAll)
}

/*

 */
func (this *FTP) RemoveFile(command *Command) {
	filePath := command.Dest

	cmd := fmt.Sprintf("DELE %s", filePath)
	this.Command(cmd, 250)
	Last("RemoveFile", "OK", this.CloseAll)
}

/*
	CreateDirectory
	Create directory to the server
*/
func (this *FTP) CreateDirectory(command *Command) {
	path := command.Dest

	cmd := fmt.Sprintf("MKD %s", path)
	this.Command(cmd, 257)
	Last("CreateDirectory", "OK", this.CloseAll)
}

/*
	RemoveDirectory
	Delete specified directory from server.
*/
func (this *FTP) RemoveDirectory(command *Command) {
	path := command.Dest

	cmd := fmt.Sprintf("RMD %s", path)
	this.Command(cmd, 250)
	Last("RemoveDirectory", "OK", this.CloseAll)
}

/*

 */
func (this *FTP) Quit() {
	this.Command("Quit", 221)
	defer this.ctrlConn.Close()
	defer this.tlsConn.Close()
	defer this.rawConn.Close()
}

/*
	GetListenSplitPort
	Get numbers for PORT command
*/
func GetListenSplitPort(port int) (port1 int, port2 int) {
	hex := fmt.Sprintf("%x", port)
	switch len(hex) {
	case 1:
	case 2:
		port1 = 0
		port2 = HexString2Int(hex)
	case 3:
		p1 := hex[0:1]
		p2 := hex[1:3]
		port1 = HexString2Int(p1)
		port2 = HexString2Int(p2)
	case 4:
		p1 := hex[0:2]
		p2 := hex[2:4]
		port1 = HexString2Int(p1)
		port2 = HexString2Int(p2)
	}
	return
}

/*
	GetLocalIP
	Return IP address of the local machine
*/
func GetLocalIP() (ip string, err error) {
	var addrs []net.Addr
	addrs, err = net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.To4().String()
			}
		}
	}
	return
}

/*
	Decimal of the string type to decimal of the integer type.
*/
/*
func DecString2Int(str string) (res int) {
	r, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		panic(err)
	}
	res = int(r)
	return
}
*/

/*
	Decimal of the string type to Hexadecimal of the string type.
*/
func DecString2HexString(dec string) string {
	p, err := strconv.Atoi(dec)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", p)
}

/*
	Hexadecimal of the string type to decimal of the integer type.
*/
func HexString2Int(hex string) (res int) {
	r, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		panic(err)
	}
	res = int(r)
	return
}
