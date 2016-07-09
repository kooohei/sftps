# SFTPS
sftps is an Package of the Golang for the FTP, FTPS and SFTP.


### Example ###

#### FTP use case ####
1. Import the package "github.com/kooohei/sftps"
```golang
import (
  "github.com/kooohei/sftps"
)
```

2. Create the parameters.
```golang
param := sftps.NewFtpParameters("[host]", [port], "[username]", "[password]", [bool for the Connection Keepalive])
```
