export GOPATH=`pwd`
go get github.com/Go-SQL-Driver/MySQL 
go build -gcflags "-N" cserver 
#go install cserver 
go build -gcflags "-N" goclient
