# SFTPS
sftps is an Package of the Golang for the FTP, FTPS and SFTP.


### Example ###

1 Import the package "github.com/kooohei/sftps"
```golang
import (
  "github.com/kooohei/sftps"
)
```

2  Create the parameters.
```golang
/*
  FTP, FTPS
*/
param := sftps.NewFtpParameters("[host]", [port], "[username]", "[password]", [bool for the Connection Keepalive])
// param.ActiveMode(123456)
// param.Secure(true)
// param.Implicit(990)
```

```golang
/*
  SFTP
*/
param := sftps.NewSftpParameters("[host]", [port], "[username]", "[password]", [bool for the Connection Keepalive])
// param.Keys("[path to the private key]", [bool for the use passphrase to the Key], "[passphrase]")
```

3  Create the Receiver

```golang
/* FTP, FTPS */
var ftp *sftps.Ftp
var err error
if ftp, err = sftps.New(sftps.FTP, param); err != nil {
  panic(err)
}
// defer ftp.Quit() /* this line should be uncomment if specified true value to the keepalive */
```

```golang
/* SFTP */
var sftp *sftps.SecureFtp
var err error
if sftp, err = sftps.New(sftps.SFTP, param); err != nil {
  panic(err)
}
```

4 Connect to Server

```golang
/* FTP, FTPS */
var res []*sftps.FtpResponse
var err error

if res, err = ftp.Connect(); err != nil {
  return
}
```
The FtpResponse contains The Executed Command String, Ftp Response Code and Response Message,
Note that: The "res" is not necessarily nil if the "err" is not nil.

```golang
/* SFTP */
if err := sftp.Connect(); err != nil {
  return
}
```


5 Execute Commands

##### Get the File List #####
```golang
/* FTP, FTPS */
var res []*sftps.FtpResponse
var err error

if res, list, err = ftp.List("."); err != nil {
  return
}
```

```golang
/* SFTP */
if list, err = sftp.List("."); err != nil {
  return
}
```

###### Useful function StringToEntities ######
```golang
if ents, err := ftp.StringToEntities(list); err != nil {
// if ents, err := sftp.StringToEntities(list); err =! nil {
  return
}
/*
  The ents is the Slice of a *Entity
  that contains information of the file.
*/
```


##### Create Directory #####
```golang
/* FTP, FTPS */
if res, err =  ftp.Mkdir("testDir"); err != nil {
  return
}
```

```golang
/* SFTP */
if err =  sftp.Mkdir("testDir"); err != nil {
  return
}
```

##### Remove Directory #####
```golang
/* FTP, FTPS */
if res, err = ftp.Rmdir("testDir"); err != nil {
  return
}
```

```golang
/* SFTP */
if err = ftp.Rmdir("testDir"); err != nil {
  return
}
```

##### Rename file or directory ######
```golang
/* FTP, FTPS */
if res, err = ftp.Rmdir("testDir"); err != nil {
  return
}
```

```golang
/* SFTP */
if err = sftp.Rmdir("testDir"); err != nil {
  return
}
```



##### Upload File #####
```golang
/* FTP, FTPS */
if res, len, err = ftp.Upload("./upload.txt", "remote.txt") err != nil {
  return
}
```

```golang
/* SFTP */
if len, err = sftp.Upload("./upload.txt", "remote.txt") err != nil {
  return
}
```

##### Download File #####
```golang
/* FTP, FTPS */
if res, len, err = ftp.Download("./downloaded.txt", "remote.txt"); err != nil {
  return
}
````

```golang
/* SFTP */
if len, err = ftp.Download("./downloaded.txt", "remote.txt"); err != nil {
  return
}
````

other functions will be ready soon.
