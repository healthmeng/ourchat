package main

import (
	"bufio"
	"container/list"
	"dbop"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	//"encoding/json"
)

type OLUser struct {
	*dbop.UserInfo
	CtrlIn   chan int
	CtrlOut  chan int
	RdMsg    chan int
	Newjob   chan string
	SendQ    *list.List // type MessageID，load from db
	MsgSet   map[int]dbop.MsgInfo
	Sendlock *sync.RWMutex
	NetConn  net.Conn
}

var online_user map[string]*OLUser
var maplock sync.RWMutex

func (ouser *OLUser) LoadMessageQ() error {
	// called only in writeproc
	msgs, err := ouser.GetUnsentMsg()
	if err != nil {
		return err
	}
	ouser.Sendlock.Lock()
	for _, msg := range msgs {
		if _, ok := ouser.MsgSet[msg.MsgID]; ok {
			continue
		} else {
			ouser.MsgSet[msg.MsgID] = msg
			ouser.SendQ.PushBack(msg.MsgID)
		}
	}
	ouser.Sendlock.Unlock()
	return nil
}

func (ouser *OLUser) ReadProc() {
	rd := bufio.NewReader(ouser.NetConn)
	for {
		cmd, _, err := rd.ReadLine()
		if err != nil {
			return
		}
		switch string(cmd) {
		case "Confirm": // confirm\n ID\n
			msgbuf, _, err := rd.ReadLine()
			if err != nil {
				log.Println("Get confirm id error:", err)
				return
			}

			msgid, err := strconv.Atoi(string(msgbuf))
			if err != nil {
				log.Println("Incorrect confirm msgid")
				continue
			}
			err = ouser.ConfirmMsg(msgid)
			if err == nil {
				ouser.Sendlock.Lock()
				delete(ouser.MsgSet, msgid)
				for it := ouser.SendQ.Front(); it != nil; it = it.Next() {
					id, _ := it.Value.(int)
					if id == msgid {
						ouser.SendQ.Remove(it)
						break
					}
				}
				ouser.Sendlock.Unlock()
			}
		case "SendMsg":
			buf,_,err:=rd.ReadLine()
			if err!=nil{
				log.Println("Get client msgid error:",err)
				return
			}
			msginfo,err:=ParseMsg(buf) // *MsgInfo
			if err!=nil{
				log.Println("Error client message format")
				rd.Reset(ouser.NetConn)
			}
			if err:=ouser.RegisterMsg(msginfo);err!=nil{
				// todo:
				// get touid, if online, post to its write queue
			}
/*
			idbuf,_,err:=rd.ReadLine()
			if err!=nil{
				log.Println("Get client msgid error:",err)
				return
			}
			msgid,err:=strconv.Atoi(string(idbuf))
			if err!=nil{
				log.Println("Get msg error: not valid msgid")
				rd.Reset(ouser.NetConn)
			}
			typebuf,_,err:=rd.ReadLine()
			if err!=nil{
				log.Println("Get client typeid error:",err)
				return
			}
			typeid,err:=strconv.Atoi(string(typebuf))
			switch typeid{
			case 1:
				content,_,err:=rd.ReadLine()
				if err!=nil{
					log.Println("Get client message content error:",err)
					return
				}
			default:
				log.Println("Unknown message type:",typeid)
				rd.Reset(ouser.NetConn)
			}*/
		}
	}
}

func (ouser *OLUser) DoSendMsg() {
	ouser.LoadMessageQ()
	ouser.Sendlock.RLock()
	for ele := ouser.SendQ.Front(); ele != nil; ele = ele.Next() {
		// do write to netconn
		msgid, _ := ele.Value.(int)
		msg, ok := ouser.MsgSet[msgid]
		if !ok {
			log.Println("Warning: message id not found in msgset")
			continue
		}
		switch msg.Type {
		case 1: // SendMsg\n MsgID(WindowID)\n MsgType\n Content\0"
			ouser.NetConn.Write(append([]byte("SendMsg\n"+fmt.Sprintf("%ld\n", msg.MsgID)+msg.Content), 0))
			//	case 2:
			//	case 3:
		}
	}
	ouser.Sendlock.RUnlock()
}

func (ouser *OLUser) WriteProc() {
	// todo  send online users to client
	users := "Users\n"
	maplock.RLock()
	for name, _ := range online_user {
		if name == ouser.Username {
			continue
		}
		users += name + "\n"
	}
	maplock.RUnlock()
	ouser.NetConn.Write([]byte(users))
	///////////////////////////
	for {
		select {
		case job := <-ouser.Newjob:
			com := strings.Split(job, "\n")
			switch com[0] {
			case "Offline":
				return
			case "Send":
				// find in db
				ouser.DoSendMsg()
			case "Reply": // OK\n+WindowNum\n
				ouser.NetConn.Write([]byte("ReplyID\n" + com[1] + "\n"))
			}
		case <-time.After(time.Minute):
			ouser.DoSendMsg()
		}
	}
}

func (ouser *OLUser) DoOffline() {
	maplock.Lock()
	delete(online_user, ouser.Username) //should be done in connection routine
	maplock.Unlock()
	ouser.Newjob <- "Offline\n"
	ouser.CtrlOut <- 1
}

func doClose(conn net.Conn, pClose *bool) {
	if *pClose {
		conn.Close()
	}
}

func DoOnline(uinfo *dbop.UserInfo, conn net.Conn) {
	oluser := &OLUser{uinfo, make(chan int, 1), make(chan int, 1), make(chan int, 1),
		make(chan string, 10), list.New(), make(map[int]dbop.MsgInfo),
		new(sync.RWMutex), conn}
	oluser.SendQ.Init()
	/*
		1. Create ouser obj in map
		2. Monitor ctrl chan
		3. Send online users to client
		4. Try to Send messages to client and enter communication mode with client(read/write)
	*/
	maplock.Lock()
	online_user[uinfo.Username] = oluser
	maplock.Unlock()
	go oluser.ReadProc()
	go oluser.WriteProc()
	for {
		select {
		case <-oluser.CtrlIn:
			// do offline
			oluser.DoOffline()
		case msg := <-oluser.RdMsg:
			/* 	1. get send response
			   	2. get heart beat
			   	3. get client's new message
			   	4. get offline inform
			*/
			switch msg {
			case 1:
				fallthrough
			case 2: // heartbeat
				fallthrough
			case 3:
				break
			default:
				oluser.DoOffline()
				return
			}
		case <-time.After(time.Second * 60):
			oluser.DoOffline()
			return
			// no message in 60 seconds, timeout
			// do offline
		}
	}
}

func procConn(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(time.Second * 30))
	//	pClose:=new(bool)
	//	*pClose=true
	//	defer doClose(conn,pClose)
	conn.Close()
	rd := bufio.NewReader(conn)
	command, _, err := rd.ReadLine()
	if err != nil {
		log.Println("Get command error:", err)
		return
	}
	switch string(command) {
	case "AddUser": // register
		user := new(dbop.UserInfo)
		if buf, _, err := rd.ReadLine(); err != nil {
			log.Println("Read user register user error:", err)
			return
		} else {
			user.Username = string(buf)
		}
		if buf, _, err := rd.ReadLine(); err != nil {
			log.Println("Read user register password error:", err)
			return
		} else {
			user.Password = string(buf)
		}

		if err := dbop.AddUser(user); err != nil {
			log.Println("Add user error:", err)
			conn.Write([]byte(err.Error() + "\n"))
		} else {
			conn.Write([]byte("OK" + "\n"))
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
		usr, _, err := rd.ReadLine()
		if err != nil {
			log.Println("Login read username error:", err)
			return
		}
		psw, _, err := rd.ReadLine()
		if err != nil {
			log.Println("Login read password error:", err)
			return
		}
		maplock.RLock()
		online, ok := online_user[string(usr)]
		maplock.RUnlock()
		if ok {
			online.CtrlIn <- 1 // start offline
			<-online.CtrlOut   // finished
			//	delete(online_user,user.Username) //should be done in connection routine
		}

		uinfo, _ := dbop.FindUser(string(usr))
		if uinfo == nil {
			conn.Write([]byte("ERROR: No such user\n"))
			return
		}
		if uinfo.Password != string(psw) {
			conn.Write([]byte("ERROR: Bad user/passwd\n"))
			return
		}
		conn.Write([]byte("OK\n"))
		//	*pClose=false
		DoOnline(uinfo, conn)

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
		user := new(dbop.UserInfo)
		if buf, _, err := rd.ReadLine(); err != nil {
			log.Println("Read user register user error:", err)
			return
		} else {
			user.Username = string(buf)
		}
		if buf, _, err := rd.ReadLine(); err != nil {
			log.Println("Read user register password error:", err)
			return
		} else {
			user.Password = string(buf)
		}

		maplock.RLock()
		online, ok := online_user[user.Username]
		maplock.RUnlock()
		if ok {
			online.CtrlIn <- 1 // start offline
			<-online.CtrlOut   // finished
			//	delete(online_user,user.Username) //should be done in connection routine
		}
		if err := dbop.DelUser(user.Username, user.Password); err != nil {
			log.Println("Del user error:", err)
			conn.Write([]byte(err.Error() + "\n"))
		} else {
			conn.Write([]byte("OK" + "\n"))
		}

	}
}

func main() {
	fmt.Println("Start")
	online_user = make(map[string]*OLUser)
	lisn, err := net.Listen("tcp", ":2048")
	if err != nil {
		fmt.Println("Server listen tcp error:", err)
		return
	}
	defer lisn.Close()
	for {
		conn, err := lisn.Accept()
		if err != nil {
			fmt.Println("Server accept error:", err)
			//return
		}
		go procConn(conn)
	}
}
