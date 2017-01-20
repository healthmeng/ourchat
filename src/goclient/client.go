package main

import (
"net"
"fmt"
"bufio"
"time"
"crypto/sha256"
//"encoding/json"
"os/exec"
"os"
// "compress/gzip"
)

var svraddr string = "127.0.0.1"
//var svraddr string = "123.206.55.31"
var svrport string = ":2048"
var connstr string=""

type UserInfo struct{
	UID int64
	Username string
	Password string // already sha256.Sum256
	Descr string
	Face string
	Phone string
	RegTime string
}


func doRegister(){
	info:=new (UserInfo)
	fmt.Println("username:")
	fmt.Scanf("%s",&info.Username)
	var orgpass , again string
	exec.Command("/bin/stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("/bin/stty", "-F", "/dev/tty", "-echo").Run()
	for{
		fmt.Println("Password:")
		fmt.Scanf("%s",&orgpass)
		fmt.Println("Password again:")
		fmt.Scanf("%s",&again)
		if again!=orgpass{
			fmt.Println("Not same,try again")
		}else{
			break
		}
	}
	exec.Command("/bin/stty", "-F", "/dev/tty", "echo").Run()
	sha:=sha256.Sum256([]byte(orgpass))
	info.Password=""
	for i:=0;i<sha256.Size;i++{
		info.Password+=fmt.Sprintf("%02x",sha[i])
	}

	conn,err:=net.Dial("tcp",connstr)
	if err!=nil{
		fmt.Println("Connect to server failed:",err)
		return
	}
	defer conn.Close()
	addtext:="AddUser\n"+info.Username+"\n"+info.Password+"\n"
/*	obj,err:=json.Marshal(info);
	if err!=nil{
		fmt.Println("json failed")
		return
	}
	addtext+=string(append(obj,'\n'))
*/
	conn.Write([]byte(addtext))
	brd:=bufio.NewReader(conn)
	if buf,_,err:=brd.ReadLine();err!=nil{
		fmt.Println("Read result error:",err)
		return
	}else{
		fmt.Println(string(buf))
	}
}

func doDel(){
	info:=new (UserInfo)
	fmt.Println("username:")
	fmt.Scanf("%s",&info.Username)
	var orgpass string
	exec.Command("/bin/stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("/bin/stty", "-F", "/dev/tty", "-echo").Run()
	fmt.Println("Password:")
	fmt.Scanf("%s",&orgpass)
	exec.Command("/bin/stty", "-F", "/dev/tty", "echo").Run()
	sha:=sha256.Sum256([]byte(orgpass))
	info.Password=""
	for i:=0;i<sha256.Size;i++{
		info.Password+=fmt.Sprintf("%02x",sha[i])
	}

	conn,err:=net.Dial("tcp",connstr)
	if err!=nil{
		fmt.Println("Connect to server failed:",err)
		return
	}
	defer conn.Close()
	addtext:="DelUser\n"+info.Username+"\n"+info.Password+"\n"
	conn.Write([]byte(addtext))
	brd:=bufio.NewReader(conn)
	if buf,_,err:=brd.ReadLine();err!=nil{
		fmt.Println("Read result error:",err)
		return
	}else{
		fmt.Println(string(buf))
	}

}

func doLogin(){
	info:=new (UserInfo)
	fmt.Println("username:")
	fmt.Scanf("%s",&info.Username)
	var orgpass string
	exec.Command("/bin/stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("/bin/stty", "-F", "/dev/tty", "-echo").Run()
	fmt.Println("Password:")
	fmt.Scanf("%s",&orgpass)
	exec.Command("/bin/stty", "-F", "/dev/tty", "echo").Run()
	sha:=sha256.Sum256([]byte(orgpass))
	info.Password=""
	for i:=0;i<sha256.Size;i++{
		info.Password+=fmt.Sprintf("%02x",sha[i])
	}

	conn,err:=net.Dial("tcp",connstr)
	if err!=nil{
		fmt.Println("Connect to server failed",err)
		return
	}
	defer conn.Close()
	addtext:="Login\n"+info.Username+"\n"+info.Password+"\n"
	conn.Write([]byte(addtext))
	brd:=bufio.NewReader(conn)
	ret,_,err:=brd.ReadLine()
	if err!=nil{
		fmt.Println("Get login result error:",err)
		return
	}
	sret:=string(ret)
	if sret!="OK"{
		fmt.Println("Login failed:"+sret)
		return
	}
	outf,_:=os.Open("/tmp/output")
	defer outf.Close()
	chw:=make(chan string,10)
	chmsg:=make(chan string,10)
	go OnlineRead(brd,chw,chmsg)
	go OnlineWrite(conn,chw)
	go ProcInput(chw)
	for{
		select{
			case <-time.After(time.Second*90):
				fmt.Println("Can't connect to server,try to logon again")
				return
			case retmsg:=<-chmsg:
				if retmsg!="Heartbeat"{
					fmt.Println(retmsg)
				//	outf.WriteString(retmsg)
				//	outf.Sync()
				}
		}
	}
}

func ProcInput(chw chan string){
	var  uid int
	var msg string
	for{
		fmt.Println("Send to(UID):")
		fmt.Scanf("%d",&uid)
		fmt.Println("Message:")
		fmt.Scanf("%s",&msg)
		output:=fmt.Sprintf("SendMsg\n%d %d %d\n%s\n",uid,1,len(msg),msg)
		chw<-output
	}
}

func OnlineWrite(conn net.Conn, chw chan string){
	tm:=time.NewTimer(time.Minute)
	for{
		select{
		case wr:=<-chw:
			if _,err:=conn.Write([]byte(wr));err!=nil{
				return
			}
		case <-tm.C:
			if _,err:=conn.Write([]byte("Heartbeat\n"));err!=nil{
				return
			}
			tm.Reset(time.Minute)
		}
	}

}

func OnlineRead(brd *bufio.Reader, chw, chmsg chan string){
	for{
		if buf,_,err:=brd.ReadLine();err==nil{
			switch string(buf){
			case "SendMsg":
				info,_,err:=brd.ReadLine()
				if err!=nil{
					return
				}
				var mid,mtype,mlen int
				var from  string;
				fmt.Sscanf(string(info),"%d%d%d%s",&mid,&mtype,&mlen,&from)
				/////
				detail,_,err:=brd.ReadLine()
				chmsg<-from+":"+string(detail)
				chw<-fmt.Sprintf("Confirm\n%d\n",mid)
			case "Heartbeat":
				chmsg<-"Heartbeat"
			case "Users":
				users,_,err:=brd.ReadLine()
				if err!=nil{
					return
				}

				chmsg<-("Users\n"+string(users))
			}
		}else{
			break
		}
	}
}

func initsvr(){
	if file,err:=os.Open("clt.ini");err==nil{
		if _,err:=fmt.Fscan(file,&connstr);err==nil{
			return
		}
	}
	connstr=svraddr+svrport
	fmt.Println(connstr)
}

func main(){
	if len(os.Args)==1{
		fmt.Println("client register(reg)\nclient del\nclient login")
		return
	}
	initsvr()
	if os.Args[1]=="reg" || os.Args[1]=="register"{
		doRegister()
	} else if os.Args[1]=="del" {
		doDel()
	} else if os.Args[1]=="login"{
		doLogin()
	}else{
		fmt.Println("client register\nclient del\nclient login")
	}
}
