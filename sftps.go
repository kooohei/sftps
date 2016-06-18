package sftps

import "encoding/json"

const (
	DEBUG = false
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
	UsePassphrase 	string
	Passphrase			string
}
type Command struct {
	Cmd 						string
	Src 						string
	Dest 						string
}


type Sftps struct {
	protocol 				string
	ftpParameter 		FtpParameter
	sftpParameter		SftpParameter
	command 				Command
}


func NewSftps(proto string, ftpParam FtpParameter, sftpParam SftpParameter) (inst *Sftps) {
	inst = &Sftps{proto, ftpParam, sftpParam}
	return
}


func (recv *Sftps) exec() {
	if recv.protocol == "FTP" || recv.protocol == "FTPS" {
		recv.FtpCommand()
	} else if recv.protocol == "SFTP" {
		recv.SftpCommand()
	}
}

func (recv *Sftps) FtpCommand() {
	if recv.command.Cmd == "ConnectTest" {
		recv.FtpConnectTest()
	} else if recv.command.Cmd == "GetFileList" {
		recv.FtpGetFileList()
	} else {
		ftp := NewFTP(recv.ftpParameter, recv.command)
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

func (recv *Sftps) FtpGetFileList() {
	ftp := NewFTP(recv.ftpParameter, DEBUG)
	ftp.Connect()
	ftp.Auth()
	ftp.BaseCommands()
	list := ftp.GetList(recv.command.Dest)

	entities, err := StringToEntities(list)
	Err("GetFileList", err, ftp.CloseAll)
	Last("GetFileList", entities, nil)
}

func (recv *Sftps) SftpCommand() {
	if recv.command.Cmd == "ConnectTest" {
		recv.SftpConnectTest()
	} else if recv.command.Cmd == "GetFileList" {
		recv.SftpGetFileList()
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

func (recv *Sftps) SftpGetFileList() {
	sftp := NewSFTP(recv.sftpParameter, DEBUG)
	sftp.ConnectAndAuth()
	list := sftp.GetList(recv.command.Dest)
	entities, err := StringToEntities(list)
	Err("GetFileList", err, sftp.CloseAll)
	bytes, err := json.Marshal(entities)
	Err("GetFileList", err, sftp.CloseAll)
	Last("GetFileList", string(bytes), nil)
}
