package sftps

import (
	"errors"
)

type FtpResponse struct {
	command string
	code    int
	msg     string
}

type Sftps struct {
	state    int
	protocol int
	recv     interface{}
	keepalive bool
	isDebug  bool
}

func New(proto int, param interface{}) (sftps *Sftps, err error) {
	sftps = new(Sftps)

	if proto == FTP || proto == FTPS {
		var parameter *ftpParameters
		if p, ok := param.(*ftpParameters); ok {
			parameter = p
			sftps.recv = newFtp(parameter)
			sftps.keepalive = parameter.keepAlive
		} else {
			err = errors.New("the 'param' could not cast to the *ftpParameters type.")
			return
		}
	} else if proto == SFTP {
		var parameter *sftpParameters
		if p, ok := param.(*sftpParameters); ok {
			parameter = p
			sftps.recv = newSftp(parameter)
			sftps.keepalive = parameter.keepAlive
		} else {
			err = errors.New("the 'param' could not cast to the *sftpParameters type.")
		}
	} else {
		err = errors.New("Invalid parameter were bound. the Protocol must be FTP, FTPS or SFTP")
		return
	}
	sftps.protocol = proto
	sftps.state = OFFLINE
	return
}

func (this *Sftps) Connect() (res []*FtpResponse, err error) {

	if this.protocol == FTP || this.protocol == FTPS {
		var r *FtpResponse
		if r, err = this.recv.(*Ftp).connect(); err != nil {
			return
		}
		res = []*FtpResponse{}
		res = append(res, r)

		var rs []*FtpResponse
		if rs, err = this.recv.(*Ftp).auth(); err != nil {
			return
		}
		res = append(res, rs...)

		if rs, err = this.recv.(*Ftp).options(); err != nil {
			return
		}
		res = append(res, rs...)
	} else
	if this.protocol == SFTP {
		if err = this.recv.(*SecureFtp).connect(); err != nil {
			return
		}
	}

	this.state = ONLINE
	return
}

func (this *Sftps) Quit() (res *FtpResponse, err error) {
	if this.protocol == FTP || this.protocol == FTPS {
		if res, err = this.recv.(*Ftp).quit(); err != nil {
			return
		}
	} else
	if this.protocol == SFTP {
		if err = this.recv.(*SecureFtp).quit(); err != nil {
			return
		}
	}
	this.state = OFFLINE
	return
}

func (this *Sftps) StringToEntities(raw string) (ents []*Entity, err error) {
	ents, err = stringToEntities(raw)
	return
}

func (this *Sftps) List(baseDir string) (res []*FtpResponse, list string, err error) {

	if this.state == OFFLINE {
		err = errors.New("Connection is not established.")
		return
	}
	if this.protocol == FTP || this.protocol == FTPS {
		var ftp *Ftp
		if recv, ok := this.recv.(*Ftp); ok {
			ftp = recv
		}
		if res, list, err = ftp.list(baseDir); err != nil {
			return
		} else {
			var r *FtpResponse
			if !this.keepalive {
				if r, err = ftp.quit(); err != nil {
					return
				}
			}
			res = append(res, r)
		}
	} else
	if this.protocol == SFTP {
		var sftp *SecureFtp
		if recv, ok := this.recv.(*SecureFtp); ok {
			sftp = recv
		}
		if list, err = sftp.list(baseDir); err != nil {
			return
		}

		if !this.keepalive {
			if err = sftp.quit(); err != nil {
				return
			}
		}
	}
	return
}


func (this *Sftps) Mkdir(p string) (res []*FtpResponse, err error) {
	if this.state == OFFLINE {
		err = errors.New("Connection is not established.")
		return
	}
	if this.protocol == FTP || this.protocol == FTPS {
		var ftp *Ftp
		var r *FtpResponse
		res = new([]*FtpResponse)
		if recv, ok := this.recv.(*Ftp); ok {
			ftp = recv
		}
		if r, err = ftp.mkdir(p); err != nil {
			return
		}
		res = append(res, r)

		if !this.keepalive {
			if r, err = ftp.quit(); err != nil {
				return
			}
			res = append(res, r)
		}
	} else
	if this.protocol == SFTP {
		var sftp *SecureFtp
		if recv, ok := this.recv.(*SecureFtp); ok {
			sftp = recv
		}
		if err = sftp.mkdir(p); err != nil {
			return
		}

		if !this.keepalive {
			if err = sftp.quit(); err != nil {
				return
			}
		}
	}
	return
}

func (this *Sftps) Rmdir(p string) (res []*FtpResponse, err error) {
	if this.state == OFFLINE {
		err = errors.New("Connection is not established.")
		return
	}
	if this.protocol == FTP || this.protocol == FTPS {
		res = new([]*FtpResponse)
		var r *FtpResponse
		var ftp *Ftp
		if recv, ok := this.recv.(*Ftp); ok {
			ftp = recv
		}
		if r, err = ftp.rmdir(p); err != nil {
			return
		}
		res = append(res, r)

		if !this.keepalive {
			if r, err = ftp.quit(); err != nil {
				return
			}
			res = append(res, r)
		}
	} else
	if this.protocol == SFTP {
		var sftp *SecureFtp
		if recv, ok := this.recv.(*SecureFtp); ok {
			sftp = recv
		}
		if err = sftp.remove(p); err != nil {
			return
		}
		if !this.keepalive {
			if err = sftp.quit(); err != nil {
				return
			}
		}
	}
	return
}

func (this *Sftps) Rename(old string, new string) (res []*FtpResponse, err error) {
	if this.state == OFFLINE {
		err = errors.New("Connection is not established")
		return
	}

	if this.protocol == FTP || this.protocol == FTPS {
		var ftp *Ftp
		if recv, ok := this.recv.(*Ftp); ok {
			ftp = recv
		}
		if res, err = ftp.rename(old, new); err != nil {
			return
		}
		if !this.keepalive {
			var r *FtpResponse
			if r, err = ftp.quit(); err != nil {
				return
			}
			res = append(res, r)
		}
	} else
	if this.protocol == SFTP {
		var sftp *SecureFtp
		if recv, ok := this.recv.(*SecureFtp); ok {
			sftp = recv
		}
		if err = sftp.rename(old, new); err != nil {
			return
		}
		if !this.keepalive {
			if err = sftp.quit(); err != nil {
				return
			}
		}
	}
	return
}

/**
	parameter's explain. local is the local path for the file, whether remote.
 */
func (this *Sftps) Upload(local string, remote string) (res []*FtpResponse, len int64, err error) {
	if this.state == OFFLINE {
		err = errors.New("Connection is not established")
		return
	}

	if this.protocol == FTP || this.protocol == FTPS {
		var ftp *Ftp
		var r *FtpResponse

		if recv, ok := this.recv.(*Ftp); ok {
			ftp = recv
		}
		if res, len, err = ftp.upload(local, remote); err != nil {
			return
		}
		if !this.keepalive {
			if r, err = ftp.quit(); err != nil {
				return
			}
			res = append(res, r)
		}
	} else
	if this.protocol == SFTP {
		var sftp *SecureFtp
		if recv, ok := this.recv.(*SecureFtp); ok {
			sftp = recv
		}
		if len, err = sftp.upload(local, remote); err != nil {
			return
		}
		if !this.keepalive {
			if err = sftp.quit(); err != nil {
				return
			}
		}
	}
	return
}

func (this *Sftps) Download(local string, remote string) (res []*FtpResponse, len int64, err error) {
	if this.state == OFFLINE {
		err = errors.New("Connection is not established")
		return
	}

	if this.protocol == FTP || this.protocol == FTPS {
		var ftp *Ftp
		var r *FtpResponse

		if recv, ok := this.recv.(*Ftp); ok {
			ftp = recv
		}
		if res, len, err = ftp.download(local, remote); err != nil {
			return
		}
		if !this.keepalive {
			if r, err = ftp.quit(); err != nil {
				return
			}
			res = append(res, r)
		}
	} else
	if this.protocol == SFTP {
		var sftp *SecureFtp
		if recv, ok := this.recv.(*SecureFtp); ok {
			sftp = recv
		}
		if len, err = sftp.download(local, remote); err != nil {
			return
		}
		if !this.keepalive {
			if err = sftp.quit(); err != nil {
				return
			}
		}
	}
	return
}
