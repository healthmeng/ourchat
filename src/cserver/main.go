﻿package main

import (
	"bufio"
	"container/list"
	"io"
	"dbop"
	"os"
	"fmt"
	"log"
	"errors"
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
	MsgSet   map[int64]dbop.MsgInfo
	Sendlock *sync.RWMutex
	LastMsgID	int64
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

func (ouser *OLUser) ParseMsg(buf []byte, rd *bufio.Reader) (*dbop.MsgInfo, error) {
	// 	ToUID Type Length\n
	//	Content
	head := string(buf)
	var msg dbop.MsgInfo
	var filelen int64
	var curmsgid int64
	fmt.Sscanf(head, "%d%d%d%d", &curmsgid,&msg.ToUID, &msg.Type, &filelen)
	if curmsgid<= ouser.LastMsgID{
		return nil,errors.New("Message out of date")
	}
	ouser.LastMsgID=curmsgid
	msg.FromUID = ouser.UID
	msg.Arrived = 0
	tm := time.Now().Local()
	msg.SvrStamp = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second())
	switch msg.Type {
	case dbop.TypeTxt: // txt
		content, _, err := rd.ReadLine()
		if err != nil {
			log.Println("Get message content error")
			return nil, err
		}
		if int64(len(content)) != filelen {
			log.Println("Warning: read bytes != msg length")
		}
		msg.Content = string(content)
	case dbop.TypePic: // image
		imgtype,_,err:=rd.ReadLine()
		if err!=nil{
			log.Println("Get image type error:",err)
			return nil,err
		}
		savepath,err:=DownloadTmp(ouser.UID,rd,filelen,string(imgtype))
		if err!=nil{
			log.Println("Download file error:",err)
			return nil,err
		}
		msg.Content=savepath
	}
	return &msg, nil
}

func DownloadTmp(uid int64, rd *bufio.Reader, fsize int64, imgtype string)(string,error){
	for{
		tmpfile:=fmt.Sprintf("%s/ourchat-%d/%d.%s",os.TempDir(),uid,time.Now().UnixNano(),imgtype)
		if _,err:=os.Stat(tmpfile);err==nil{
			time.Sleep(time.Nanosecond*1000)
			continue
		}
		fd,_:=os.Create(tmpfile)
		ncp,err:=io.CopyN(fd,rd,fsize)
		if err!=nil || ncp!=fsize{
			fd.Close()
			os.Remove(tmpfile)
			return "",errors.New("Transfer error")
		}
		fd.Close()
		return tmpfile,nil
	}
}

func (ouser *OLUser) ReadProc() {
	rd := bufio.NewReader(ouser.NetConn)
	for {
		cmdbuf, _, err := rd.ReadLine()
		if err != nil {
			return
		}
		cmd := string(cmdbuf)
		switch cmd {
		case "Confirm": // confirm\n ID\n
			msgbuf, _, err := rd.ReadLine()
			if err != nil {
				log.Println("Get confirm id error:", err)
				return
			}

			msgid, err := strconv.ParseInt(string(msgbuf), 10, 64)
			if err != nil {
				log.Println("Incorrect confirm msgid")
				continue
			}
			err = ouser.ConfirmMsg(msgid)
			if err == nil {
		        msg, ok := ouser.MsgSet[msgid]
				if !ok {
					log.Println("Kidding me? Confirmed message id not found in msgset")
				}else{
					os.Remove(msg.Content)
				}
				ouser.Sendlock.Lock()
				delete(ouser.MsgSet, msgid)
				for it := ouser.SendQ.Front(); it != nil; it = it.Next() {
					id, _ := it.Value.(int64)
					if id == msgid {
						ouser.SendQ.Remove(it)
						break
					}
				}
				ouser.Sendlock.Unlock()
			}
			ouser.RdMsg <- 1
		case "Heartbeat":
			ouser.Newjob <- "Heartbeat"
			ouser.RdMsg <- 2

		case "GetUserInfo":
			ouser.Newjob <- "UsrUpdate"
		/*	usrlst,err:=dbop.ListUsers()
			if err!=nil{
				conn.Write([]byte("ERROR:"+err.Error()))
			}else{
				conn.Write([]byte(fmt.Sprintf("UserList\n%d\n",len(usrlst))))
				for _,usr:=range(usrlst){
					conn.Write([]byte(fmt.Sprintf("id:%d;name:%s;descr:%s;face:%s;phone:%s\n",usr.UID,usr.Username,usr.Descr,usr.Face,usr.Phone)))
				}
			}
		*/
			ouser.RdMsg<-5
		case "SendMsg":
			buf, _, err := rd.ReadLine()
			if err != nil {
				log.Println("Get client msgid error:", err)
				return
			}
			msginfo, err := ouser.ParseMsg(buf, rd) // *MsgInfo

			if err != nil {
			// may repeat messages
		//		log.Println("Error client message format")
				//		rd.Reset(ouser.NetConn) can't reset, or may lose heatbeat
				break
			} else {
				//	log.Println("msg:",msginfo.Content)
			}
			usr, err := dbop.LookforUID(msginfo.ToUID)
			if err != nil {
				log.Println("Lookfor uid error")
				//		rd.Reset(ouser.NetConn)
				break
			}
			if err := ouser.RegisterMsg(msginfo); err == nil {
				// todo:
				// get touid, if online, post to its write queue
				maplock.RLock()
				toclient, ok := online_user[usr.Username]
				maplock.RUnlock()
				ouser.RdMsg <- 3
				if ok {
					toclient.Newjob <- "SendMsg"
				}
			}
		case "Offline":
			// imform all online users reload user info
	//		ouser.DoOffline()
			ouser.RdMsg<-4
			return
		default:
			log.Println("Unknown message type:", cmd)
			//	rd.Reset(ouser.NetConn)
		}
	}
}

func (ouser *OLUser) DoSendMsg() {
	ouser.LoadMessageQ()
	ouser.Sendlock.RLock()
	for ele := ouser.SendQ.Front(); ele != nil; ele = ele.Next() {
		// do write to netconn
		msgid, _ := ele.Value.(int64)
		msg, ok := ouser.MsgSet[msgid]
		if !ok {
			log.Println("Warning: message id not found in msgset")
			continue
		}
		switch msg.Type {
		case dbop.TypeTxt: // SendMsg\n MsgID(WindowID) MsgType MsgLen time\n Content\n"
			ouser.NetConn.Write([]byte("SendMsg\n" + fmt.Sprintf("%d %d %d %d %s\n", msg.MsgID, msg.Type, len(msg.Content)+1, msg.FromUID, msg.SvrStamp) + msg.Content+"\n"))
		case dbop.TypePic:
			finfo,err:=os.Stat(msg.Content)
			if err!=nil{
				log.Println("Image tmp file not found")
				continue
			}
			names:=strings.Split(msg.Content,".")
			num:=len(names)
			var imgtype string
			if num>1{
				imgtype=names[num-1]
			}else{
				imgtype="cimg"
			}
			fd,_:=os.Open(msg.Content)
			fsize:=finfo.Size()
			ouser.NetConn.Write([]byte("SendMsg\n"+fmt.Sprintf("%d %d %d %d %s\n%s\n",msg.MsgID,msg.Type,fsize,msg.FromUID,msg.SvrStamp,imgtype)))
			io.CopyN(ouser.NetConn,fd,fsize)
			//	case 3:
		}
	}
	ouser.Sendlock.RUnlock()
}

func (ouser *OLUser) UpdateUser() {
	users := "Users\n"
	maplock.RLock()
	for name, cuser := range online_user {
		if name == ouser.Username {
			continue
		}
		users += fmt.Sprintf("%d:%s|", cuser.UID, name)
	}
	maplock.RUnlock()
	ouser.NetConn.Write([]byte(users + "\n"))
}

func (ouser *OLUser) SendUserList(){
	usrlst, err := dbop.ListUsers()
	if err != nil {
		ouser.NetConn.Write([]byte("ERROR:" + err.Error()))
	} else {
		ouser.NetConn.Write([]byte(fmt.Sprintf("UserList\n%d\n", len(usrlst))))
		for _, usr := range usrlst {
			ouser.NetConn.Write([]byte(fmt.Sprintf("%d;%s;%s;%s;%s\n", usr.UID, usr.Username, usr.Descr, usr.Face, usr.Phone)))
		}
	}
}

func (ouser *OLUser) WriteProc() {
	ouser.SendUserList()
	ouser.UpdateUser() // send all users info to client
	ouser.DoSendMsg()  // try to send offline messages to client
	for {
		select {
		case job := <-ouser.Newjob:
			com := strings.Split(job, "\n")
			switch com[0] {
			case "Offline":
	//			log.Println("user:",ouser.Username,"now offline")
				return
			case "Heartbeat":
				ouser.NetConn.Write([]byte("Heartbeat\n"))
			//case "UserInfo":
			case "SendMsg":
				// find in db
				ouser.DoSendMsg()
				//			case "Reply": // OK\n+WindowNum\n
				//				ouser.NetConn.Write([]byte("ReplyID\n" + com[1] + "\n"))
			case "Refresh":
				//////////
				ouser.UpdateUser()
			case "UsrUpdate":
				ouser.SendUserList()
			}
		case <-time.After(time.Minute):
			ouser.DoSendMsg()
		}
	}
}

func (ouser *OLUser) DoOffline() {
	maplock.Lock()
//	ouser.NetConn.Close()
	delete(online_user, ouser.Username) //should be done in connection routine
	maplock.Unlock()
	ouser.Newjob <- "Offline\n" // quit write proc, \n to be splitted
	///////////////

	maplock.RLock()
	for _, usr := range online_user {
		usr.Newjob <- "Refresh"
	}
	maplock.RUnlock()
	ouser.CtrlOut <- 1
}

func doClose(conn net.Conn, pClose *bool) {
	if *pClose {
		conn.Close()
	}
}

func DoOnline(uinfo *dbop.UserInfo, conn net.Conn) {
	oluser := &OLUser{uinfo, make(chan int, 1), make(chan int, 1),
		make(chan int, 10), make(chan string, 10),
		list.New(), make(map[int64]dbop.MsgInfo),
		new(sync.RWMutex), -1,conn}
	oluser.SendQ.Init()
	os.Mkdir(fmt.Sprintf("%s/ourchat-%d",os.TempDir(),oluser.UID),0644)
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
	////////////////
	// todo: inform all other online users to send new user online msg
	for name, cuser := range online_user {
		if name == oluser.Username {
			continue
		}
		cuser.Newjob <- "Refresh"
	}

	for {
		select {
		case <-oluser.CtrlIn:
			// do offline
			oluser.DoOffline()
			return
		case msg := <-oluser.RdMsg:
			/* 	//1. get send response
			   	//2. get heart beat
			   	//3. get client's new message
			   	//4. get offline inform
			   	//5. get user list 
			*/
			switch msg {
			//	case 1: // confirm ok
			//	case 2: // heartbeat
			//	case 3:
			//	break
			case 4:
				oluser.DoOffline()
				return
			}
		case <-time.After(time.Second * 100):
			oluser.DoOffline()
			return
			// no message in 120 seconds, timeout
			// do offline
		}
	}
}

func procConn(conn net.Conn) {
	//	conn.SetDeadline(time.Now().Add(time.Second * 30))
	//	pClose:=new(bool)
	//	*pClose=true
	//	defer doClose(conn,pClose)
	defer func(){
//		log.Println("Now close conn")
		 conn.Close()
	}()
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
					OK\nID\n | ERROR:err\n(close)
					go routine (read,write,chan)
		*/
		usr, _, err := rd.ReadLine()
		user := string(usr)
		if err != nil {
			log.Println("Login read username error:", err)
			return
		}

		psw, _, err := rd.ReadLine()
		if err != nil {
			log.Println("Login read password error:", err)
			return
		}

		uinfo, _ := dbop.FindUser(user)
		if uinfo == nil {
			conn.Write([]byte("ERROR: No such user\n"))
			return
		}
		if uinfo.Password != string(psw) {
			conn.Write([]byte("ERROR: Bad user/passwd\n"))
			return
		}

		conn.Write([]byte(fmt.Sprintf("OK\n%d\n", uinfo.UID)))
		maplock.RLock()
		online, ok := online_user[user]
		maplock.RUnlock()
		if ok {
			online.CtrlIn <- 1 // start offline
			<-online.CtrlOut   // finished
			//	delete(online_user,user.Username) //should be done in connection routine
		}


		//	*pClose=false
		DoOnline(uinfo, conn)

	case "DelUser":
		/*		buf,_,err:=rd.ReadLine()
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
