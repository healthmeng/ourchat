package main

import(
"fmt"
"log"
"net"
"bufio"
"time"
"dbop"
"container/list"
"sync"
//"encoding/json"
)

type Message struct{
	MsgID int
	Msgtype int
	Msgtxt string
	Msgfrom int64
	Msgto	int64
	Msgtime string
}

type OLUser struct{
	*dbop.UserInfo
	Ctrloff chan int
	Ctrlmsg chan int
	SendQ *list.List // type Message
	Sendlock sync.Mutex
	PostQ *list.List
	PostLock sync.Mutex
}

var online_user map[string] *OLUser

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
		case "AddUser":// register
			user:=new(dbop.UserInfo)
			if buf,_,err:=rd.ReadLine();err!=nil{
				log.Println("Read user register user error:",err)
				return
			}else{
				user.Username=string(buf)
			}
			if buf,_,err:=rd.ReadLine();err!=nil{
				log.Println("Read user register password error:",err)
				return
			}else{
				user.Password=string(buf)
			}

			if err:=dbop.AddUser(user);err!=nil{
				log.Println("Add user error:",err)
				conn.Write([]byte(err.Error()+"\n"))
			}else{
				conn.Write([]byte("OK"+"\n"))
			}

		case "Login": // keep heartbeat, if read or write failed or timeout, treat as logout; pic and wav，use tcp(COPYN)
	/* 
<---
		login\n
		username\n
		passwdsha256\n
--->
		OK\n | ERROR:err\n(close)
		go routine (read,write,chan)
	*/
		/*	buf,_,err:=rd.ReadLine()
			if err!=nil{
				log.Println("Read login info error:",err)
			}*/
		case "DelUser":
/*			buf,_,err:=rd.ReadLine()
			if err!=nil{
				log.Println("Read user register info error:",err)
				return
			}
			user:=new(dbop.UserInfo)
			if err:=json.Unmarshal(buf,user);err!=nil{
				log.Println("Resolve user info error:",err)
				return
			}
*/
			user:=new(dbop.UserInfo)
			if buf,_,err:=rd.ReadLine();err!=nil{
				log.Println("Read user register user error:",err)
				return
			}else{
				user.Username=string(buf)
			}
			if buf,_,err:=rd.ReadLine();err!=nil{
				log.Println("Read user register password error:",err)
				return
			}else{
				user.Password=string(buf)
			}


			if online,ok:=online_user[user.Username];ok{
				online.Ctrloff<-1 // start offline
				<-online.Ctrloff // finished 
			//	delete(online_user,user.Username) //should be done in connection routine
			}
			if err:=dbop.DelUser(user.Username,user.Password);err!=nil{
				log.Println("Del user error:",err)
				conn.Write([]byte(err.Error()+"\n"))
			}else{
				conn.Write([]byte("OK"+"\n"))
			}

	}
}

func main(){
	fmt.Println("Start")
	online_user=make( map[string] *OLUser)
	lisn,err:=net.Listen("tcp",":2048")
	if err!=nil{
		fmt.Println("Server listen tcp error:",err)
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
