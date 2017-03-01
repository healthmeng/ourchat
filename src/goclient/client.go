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
"strconv"
// "compress/gzip"
"runtime"
"strings"
"convgbk"
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

func setEcho(enable bool){
	if strings.ToLower(runtime.GOOS)!="windows"{
		if !enable{
			exec.Command("/bin/stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
			exec.Command("/bin/stty", "-F", "/dev/tty", "-echo").Run()
		}else{
			exec.Command("/bin/stty", "-F", "/dev/tty", "echo").Run()
		}
	}
}

func doRegister(){
	info:=new (UserInfo)
	fmt.Println("username:")
	fmt.Scanf("%s",&info.Username)
	var orgpass , again string
	setEcho(false)
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
	setEcho(true)
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
	setEcho(false)
	fmt.Println("Password:")
	fmt.Scanf("%s",&orgpass)
	setEcho(true)
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
	setEcho(false)
	fmt.Println("Password:")
	fmt.Scanf("%s",&orgpass)
	setEcho(true)
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
	ret,_,_=brd.ReadLine()
	fmt.Println("Your uid:",string(ret))

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
				if retmsg=="Closed"{
					fmt.Println("Remote connection closed.")
					return
				}
				if retmsg!="Heartbeat"{
					fmt.Println("receive:",time.Now(),":",retmsg)
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
		msg,_=convgbk.UTF2GB(msg)
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
			fmt.Println("write error!")
				return
			}
		case <-tm.C:
			fmt.Println("send hearbeat:",time.Now())
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
				var mid,mtype,mlen,from int
				var  tmstamp string;
				fmt.Sscanf(string(info),"%d%d%d%d%s",&mid,&mtype,&mlen,&from,&tmstamp)
				/////
				detail,_,err:=brd.ReadLine()
				if err!=nil{
					return
				}
				strdetail,_:=convgbk.GB2UTF(string(detail))
				chmsg<-fmt.Sprintf("%d %s :%s",from,tmstamp,strdetail)
				chw<-fmt.Sprintf("Confirm\n%d\n",mid)
			case "Heartbeat":
				chmsg<-"Heartbeat"
			case "Users":
				users,_,err:=brd.ReadLine()
				if err!=nil{
					return
				}
				chmsg<-("Users\n"+string(users))
			case "UserList":
				if count,_,err:=brd.ReadLine();err==nil{
					cnt,er:=strconv.Atoi(string(count))
					if er==nil{
						println("Total ",cnt,"users.")
					}else{
						fmt.Println(er)
						return
					}
					for i:=0;i<cnt;i++{
						line,_,err:=brd.ReadLine()
						if err!=nil{
							fmt.Println("Read error:",err)
							break
						}
						fmt.Println(string(line))
					}
				}
			}
		}else{ // read error 
			chmsg<-"Closed"
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
