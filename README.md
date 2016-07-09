# SFTPS
sftps is an Package of the Golang for the FTP, FTPS and SFTP.


### Example ###

#### FTP, FTPS use case ####
1. Import the package "github.com/kooohei/sftps"
```golang
import (
  "github.com/kooohei/sftps"
)
```

2. Create the parameters.
```golang
param := sftps.NewFtpParameters("[host]", [port], "[username]", "[password]", [bool for the Connection Keepalive])
// param.ActiveMode(123456)
// param.Secure(true)
// param.Implicit(990)
```

3. Create the Receiver
```golang
if ftp, err := sftps.New(sftps.FTP, param); err != nil {
  panic(err)
}
// defer ftp.Quit() /* this line should be uncomment if specified true value to the keepalive */
```

4. Execute Commands

##### Get the File List #####
```golang
if res, list, err := ftp.List("."); err != nil {
  return
}
```
the "res" is the pointer of the FtpResponse, that contain a Executed Command String, FTP Response Code and Response Message.
