package main

import(
"fmt"
"log"
"net"
"bufio"
"time"
"dbop"
"container/list"
"time"
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

func (msg* Message)SendOK(){
}


type OLUser struct{
	*dbop.UserInfo
	Ctrloff chan int
	Statmsg chan int
	Newjob	chan string 
	SendQ *list.List // type Message，load from db
	Sendlock *sync.Mutex
	NetConn net.Conn
}



var online_user map[string] *OLUser
var maplock sync.RWMutex

func (ouser* OLUser)LoadMessageQ(){
	// called only in writeproc
}

func (ouser* OLUser)ReadProc(){
}

func (ouser* OLUser)WriteProc(){
	for{
		select{
			case job:=<-ouser.Newjob:
			case time.After(timer.Second*
		}
	}
}

func doClose(conn net.Conn, pClose *bool){
	if *pClose{
		conn.Close()
	}
}

func DoOnline(uinfo* dbop.UserInfo, conn net.Conn){
	oluser:= &OLUser{dbopUser,
					make(chan, int), make (chan int), make (chan string),
					list.New(), new (sync.Mutex),conn}
/*
		1. Create ouser obj in map
		2. Monitor ctrl chan
		3. Try to Send messages to client and enter communication mode with client(read/write)
*/
	maplock.Lock()
	online_user[userinfo.Username]=oluser
	maplock.Unlock()
	go  oluser.ReadProc
	go  oluser.WriteProc
	for{
		select{
		case ctl:= <-oluser.Ctrloff:
			// do offline
		case msg:=<-oluser.Statmsg:
/* 	1. get send response
	2. get heart beat
	3. get client's new message
	4. get offline inform
	*/
			select msg{
				case 1:

				case 2: // heartbeat
				default:
			}
		case <-time.After(time.Second*60):
			// no message in 60 seconds, timeout
			// do offline
		}
	}
}

func procConn(conn net.Conn){
	conn.SetDeadline(time.Now().Add(time.Second*30))
//	pClose:=new(bool)
//	*pClose=true
//	defer doClose(conn,pClose)
	conn.Close()
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
			usr,_,err:=rd.ReadLine()
			if err!=nil{
				log.Println("Login read username error:",err)
				return
			}
			psw,_,err:=rd.ReadLine()
			if err!=nil{
				log.Println("Login read password error:",err)
				return
			}
			maplock.RLock()
			online,ok:=online_user[user.Username]
			maplock.RUnlock()
			if ok{
				online.Ctrloff<-1 // start offline
				<-online.Ctrloff // finished 
			//	delete(online_user,user.Username) //should be done in connection routine
			}

			uinfo,_:=dbop.FindUser(string(usr))
			if uinfo==nil{
				conn.Write([]byte("ERROR: No such user\n"))
				return
			}
			if uinfo.Password!=psw{
				conn.Write([]byte("ERROR: Bad user/passwd\n"))
				return
			}
			conn.Write([]byte("OK\n"))
		//	*pClose=false
			DoOnline(uinfo,conn)


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


			maplock.RLock()
			online,ok:=online_user[user.Username]
			maplock.RUnlock()
			if ok{
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
