package sftps

import (
	"encoding/json"
	"log"
)

const (
	DEBUG = true
)

type FtpParameter struct {
	Host 						string
	Port 						int
	ListenPort 			int
	User 						string
	Pass 						string
	Passive 				bool
	Secure 					bool
	AlwaysTrust			bool
	SecureMode			string
	Cert 						string
	Key 						string
}
type SftpParameter struct {
	Host 						string
	Port 						int
	User 						string
	Pass 						string
	UseKey 					bool
	PrivateKey 			string
	UsePassphrase 	bool
	Passphrase			string
}
type Command struct {
	Cmd 						string
	Src 						string
	Dest 						string
}


type Sftps struct {
	protocol 				string
	ftpParameter 		*FtpParameter
	sftpParameter		*SftpParameter
	command 				*Command
}


func NewSftps(proto string, ftpParam *FtpParameter, sftpParam *SftpParameter, cmd *Command) (inst *Sftps) {
	log.Printf("%s", ftpParam.User)
	inst = &Sftps{proto, ftpParam, sftpParam, cmd}
	return
}


func (recv *Sftps) Exec() {
	log.Printf(recv.protocol)
	if recv.protocol == "FTP" || recv.protocol == "FTPS" {
		recv.ftpCommand()
	} else if recv.protocol == "SFTP" {
		recv.sftpCommand()
	}
}

func (recv *Sftps) ftpCommand() {
	if recv.command.Cmd == "ConnectTest" {
		recv.FtpConnectTest()
	} else if recv.command.Cmd == "GetFileList" {
		recv.ftpGetFileList()
	} else {
		ftp := NewFTP(recv.ftpParameter, DEBUG)
		ftp.Connect()
		ftp.Auth()
		ftp.BaseCommands()

		funcs := map[string]interface{} {
			"DownlloadFile":		ftp.DownloadFile,
			"UploadFile":      ftp.UploadFile,
			"RemoveFile":      ftp.RemoveFile,
			"CreateDirectory": ftp.CreateDirectory,
			"RemoveDirectory": ftp.RemoveDirectory,
			"Rename":          ftp.Rename,
		}
		ftp.Call(funcs, recv.command.Cmd, recv.command)
	}
}

func (recv *Sftps) FtpConnectTest() {
	ftp := NewFTP(recv.ftpParameter, DEBUG)
	ftp.Connect()
	ftp.Auth()
	Last("ConnectTest", "OK", ftp.CloseAll)
}

func (recv *Sftps) ftpGetFileList() {
	ftp := NewFTP(recv.ftpParameter, DEBUG)
	ftp.Connect()
	ftp.Auth()
	ftp.BaseCommands()
	list := ftp.GetList(recv.command.Dest)
	entities, err := StringToEntities(list)
	Err("GetFileList", err, ftp.CloseAll)
	Last("GetFileList", entities, nil)
}

func (recv *Sftps) sftpCommand() {
	if recv.command.Cmd == "ConnectTest" {
		recv.SftpConnectTest()
	} else if recv.command.Cmd == "GetFileList" {
		recv.sftpGetFileList()
	} else {
		sftp := NewSFTP(recv.sftpParameter, DEBUG)
		sftp.ConnectAndAuth()

		funcs := map[string]interface{} {
			"CreateDirectory": sftp.CreateDirectory,
			"Rename":          sftp.Rename,
			"UploadFile":      sftp.UploadFile,
			"DownloadFile":    sftp.DownloadFile,
		}
		sftp.Call(funcs, recv.command.Cmd, recv.command)
	}
}

func (recv *Sftps) SftpConnectTest() {
	sftp := NewSFTP(recv.sftpParameter, DEBUG)
	sftp.ConnectAndAuth()
	Last("ConnectTest", "OK", sftp.CloseAll)
}

func (recv *Sftps) sftpGetFileList() {
	sftp := NewSFTP(recv.sftpParameter, DEBUG)
	sftp.ConnectAndAuth()
	list := sftp.GetList(recv.command.Dest)
	entities, err := StringToEntities(list)
	Err("GetFileList", err, sftp.CloseAll)
	bytes, err := json.Marshal(entities)
	Err("GetFileList", err, sftp.CloseAll)
	Last("GetFileList", string(bytes), nil)
}
