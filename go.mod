//pour faire ce fichier : go mod init example.com/dns-txt
//                        go get github.com/miekg/dns


module example.com/dns-txt

go 1.22.4

require (
	github.com/miekg/dns v1.1.63 
	golang.org/x/mod v0.18.0 
	golang.org/x/net v0.31.0 
	golang.org/x/sync v0.7.0 
	golang.org/x/sys v0.27.0 
	golang.org/x/tools v0.22.0 
)