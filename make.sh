export GOPATH=`pwd`
go get github.com/Go-SQL-Driver/MySQL 
go install cserver 
go build -gcflags "-N" goclient
