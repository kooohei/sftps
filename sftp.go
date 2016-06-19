package sftps

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
	"reflect"
)

type SecureFTP struct {
	params     *SftpParameter
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	isDebug    bool
}

func NewSFTP(parameters *SftpParameter, debug bool) (sftp *SecureFTP) {
	sftp = new(SecureFTP)
	sftp.params = parameters
	sftp.isDebug = debug
	return
}

func (this *SecureFTP) Call(m map[string]interface{}, name string, params ...interface{}) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	f.Call(in)
	return
}

func (this *SecureFTP) CloseAll() {
	if this.sftpClient != nil {
		this.sftpClient.Close()
	}
	if this.sshClient != nil {
		this.sshClient.Close()
	}
}
func (this *SecureFTP) ConnectAndAuth() {
	var err error
	var signer ssh.Signer

	config := &ssh.ClientConfig{User: this.params.User}

	if this.params.UseKey {
		pemBytes, err := ioutil.ReadFile(this.params.PrivateKey)
		Err("ConnectAuth", err, this.CloseAll)
		if this.params.UsePassphrase {
			block, _ := pem.Decode(pemBytes)
			decPEM, err := x509.DecryptPEMBlock(block, []byte(this.params.Passphrase))
			Err("ConnectAuth", err, this.CloseAll)
			pkString := base64.StdEncoding.EncodeToString(decPEM)
			key := fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----\n", block.Type, pkString, block.Type)
			signer, err = ssh.ParsePrivateKey([]byte(key))
			Err("ConnectAuth", err, this.CloseAll)
		} else {
			signer, err = ssh.ParsePrivateKey(pemBytes)
			Err("ConnectAuth", err, this.CloseAll)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}
	if this.params.Pass != "" {
		config.Auth = append(config.Auth, ssh.Password(this.params.Pass))
	}

	config.SetDefaults()
	hosts, err := net.LookupIP(this.params.Host)
	Err("ConnectAuth", err, this.CloseAll)
	addr := fmt.Sprintf("%s:%d", hosts[0], this.params.Port)

	this.sshClient, err = ssh.Dial("tcp", addr, config)
	Err("ConnectAuth", err, this.CloseAll)
	this.sftpClient, err = sftp.NewClient(this.sshClient)
	Err("ConnectAuth", err, this.CloseAll)

	return
}
func (this *SecureFTP) GetList(path string) (list string) {
	session, err := this.sshClient.NewSession()
	defer session.Close()
	Err("GetList", err, this.CloseAll)
	cmd := fmt.Sprintf("ls -al %s", path)
	bytes, err := session.Output(cmd)
	Err("GetList", err, this.CloseAll)
	list = string(bytes)
	return
}

/*
func (this *SecureFTP) GetList(path string) (res string) {
	//list, err := this.sftpClient.ReadDir(path)
	//Err("GetList", err, this.CloseAll)
	//spew.Dump(list)
	//_ = list
	//spew.Dump("")
	//spew.Dump(list)
	this.getList(path)
	return ""
}
*/
func (this *SecureFTP) CreateDirectory(command *Command) {
	err := this.sftpClient.Mkdir(command.Dest)
	Err("CreateDirectory", err, this.CloseAll)
	Last("CreateDirectory", "OK", this.CloseAll)
}

func (this *SecureFTP) Remove(command *Command) {
	session, err := this.sshClient.NewSession()
	defer session.Close()
	Err("Remove", err, this.CloseAll)
	bytes, err := session.Output(fmt.Sprintf("rm -rf %s", command.Dest))
	Err("Remove", err, this.CloseAll)
	Last("Remove", string(bytes), this.CloseAll)
}

func (this *SecureFTP) CreateFile(command *Command) {
	_, err := this.sftpClient.Create(command.Dest)
	Err("CreateFile", err, this.CloseAll)
	Last("CreateFile", "OK", this.CloseAll)
}

func (this *SecureFTP) Rename(command *Command) {
	err := this.sftpClient.Rename(command.Src, command.Dest)
	Err("Rename", err, this.CloseAll)
	Last("Rename", "OK", this.CloseAll)
}

func (this *SecureFTP) UploadFile(command *Command) {

	reader, err := os.Open(command.Src)
	Err("UploadFile", err, this.CloseAll)
	defer reader.Close()

	src, err := ioutil.ReadAll(reader)
	Err("UploadFile", err, this.CloseAll)

	dest, err := this.sftpClient.Create(command.Dest)
	Err("UploadFile", err, this.CloseAll)
	defer dest.Close()

	_, err = dest.Write(src)
	Err("UploadFile", err, this.CloseAll)
	Last("UploadFile", "OK", this.CloseAll)
}

func (this *SecureFTP) DownloadFile(command *Command) {
	reader, err := this.sftpClient.Open(command.Dest)
	Err("DownloadFile", err, this.CloseAll)

	dest, err := ioutil.ReadAll(reader)
	Err("DownloadFile", err, this.CloseAll)

	src, err := os.Create(command.Src)
	Err("DownloadFile", err, this.CloseAll)

	_, err = src.Write(dest)
	Err("DownloadFile", err, this.CloseAll)
	Last("DownloadFile", "OK", this.CloseAll)
}
