export GOPATH=`pwd`
go get github.com/Go-SQL-Driver/MySQL 
go install server 
go build -gcflags "-N" goclient
