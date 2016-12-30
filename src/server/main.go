package main

import(
"fmt"
"log"
"net"
"bufio"
"time"
"dbop"
"encoding/json"
)

func procConn(conn net.Conn){
	conn.SetDeadline(time.Now().Add(time.Second*30))
	defer conn.Close()
	rd:=bufio.NewReader(conn)
	command,_,err:=rd.ReadLine()
	if err!=nil{
		log.Println("Get command error:",err)
		return
	}
	switch string(command){
		case "AddUser":
			buf,_,err:=rd.ReadLine()
			if err!=nil{
				log.Println("Read user register info error:",err)
				return
			}
			user:=new(dbop.UserInfo)
			if err:=json.Unmarshal(buf,user);err!=nil{
				log.Println("Resolve user info error:",err)
				return
			}
			if err:=dbop.AddUser(user);err!=nil{
				log.Println("Add user error:",err)
				conn.Write([]byte(err.Error()+"\n"))
			}else{
				conn.Write([]byte("OK"+"\n"))
			}

		case "Login":
		case "DelUser":
	}
}

func main(){
	fmt.Println("Start")
	lisn,err:=net.Listen("tcp",":2048")
	if err!=nil{
		fmt.Println("Server listen error:",err)
		return
	}
	defer lisn.Close()
	for{
		conn,err:=lisn.Accept()
		if err!=nil{
			fmt.Println("Server accept error:",err)
			//return
		}
		go procConn(conn)
	}
}
