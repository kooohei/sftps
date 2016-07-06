package sftps

import (
	"errors"
//"github.com/davecgh/go-spew/spew"
)

type Response struct {
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
			sftps.recv = NewFtp(parameter)
		} else {
			err = errors.New("'param' could not cast to the *ftpParameter")
			return
		}
	} else if proto == SFTP {

	} else {
		err = errors.New("Invalid parameter were bound. the Protocol must be FTP, FTPS or SFTP")
		return
	}
	sftps.protocol = proto
	sftps.state = OFFLINE
	return
}

func (this *Sftps) Connect() (res []*Response, err error) {

	if this.protocol == FTP || this.protocol == FTPS {
		var r *Response
		if r, err = this.recv.(*Ftp).connect(); err != nil {
			return
		}
		res = []*Response{}
		res = append(res, r)

		var rs []*Response
		if rs, err = this.recv.(*Ftp).auth(); err != nil {
			return
		}
		res = append(res, rs...)

		if rs, err = this.recv.(*Ftp).options(); err != nil {
			return
		}
		res = append(res, rs...)
	}

	if this.protocol == SFTP {
		//err := this.Receiver.(*Sftp).ConnectAndAuth()
	}

	this.state = ONLINE
	return
}

func (this *Sftps) Quit() {
	if this.protocol == FTP || this.protocol == FTPS {
		this.recv.(*Ftp).quit()
	}
}

func (this *Sftps) StringToEntities(raw string) (ents []*Entity, err error) {
	ents, err = stringToEntities(raw)
	return
}

func (this *Sftps) List(baseDir string) (res []*Response, list string, err error) {
	if this.state == OFFLINE {
		err = errors.New("Connection is not established.")
		return
	}
	if this.protocol == FTP || this.protocol == FTPS {
		var ftp *Ftp
		if r, ok := this.recv.(*Ftp); ok {
			ftp = r
		} else {
			err = errors.New("Internal error. Error occurred in the Receiver.")
			return
		}
		if res, list, err = ftp.list(baseDir); err != nil {
			if !this.keepalive {
				ftp.quit()
			}
			return
		}
	}
	return
}


func (this *Sftps) Mkdir(p string) (res *Response, err error) {
	if this.state == OFFLINE {
		err = errors.New("Connection is not established.")
		return
	}
	if this.protocol == FTP || this.protocol == FTPS {
		var ftp *Ftp
		if r, ok := this.recv.(*Ftp); ok {
			ftp = r
		} else {
			err = errors.New("Internal error. Error occurred in the Receiver.")
			return
		}

		if res, err = ftp.mkdir(p); err != nil {
			if !this.keepalive {
				ftp.quit()
			}
			return
		}
	}
	return
}

func (this *Sftps) Rmdir(p string) (res *Response, err error) {
	if this.state == OFFLINE {
		err = errors.New("Connection is not established.")
		return
	}
	if this.protocol == FTP || this.protocol == FTPS {
		var ftp *Ftp
		if r, ok := this.recv.(*Ftp); ok {
			ftp = r
		} else {
			err = errors.New("Internal error. Error occurred in the Receiver.")
			return
		}

		if res, err = ftp.rmdir(p); err != nil {
			if !this.keepalive {
				ftp.quit()
			}
			return
		}
	}
	return
}
