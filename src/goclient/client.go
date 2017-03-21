package main

import (
"net"
"io"
"fmt"
"bufio"
"time"
"crypto/sha256"
"os/exec"
"os"
"strconv"
"runtime"
"strings"
"convgbk"
"dbop"
)

var svraddr string = "127.0.0.1"
//var svraddr string = "123.206.55.31"
var svrport string = ":2048"
var connstr string=""
var gmsgid int64=0

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
				}else if strings.HasPrefix(retmsg,"Text:"){
					fmt.Println("receive:",strings.TrimPrefix(retmsg,"Text:"))
				}else if strings.HasPrefix(retmsg,"Users\n"){
					fmt.Println(retmsg)
				}
		}
	}
}

func ProcInput(chw chan string){
	var user string
	var msg string
	var output string
	for{
		user=""
		fmt.Println("Send to(username):")
		fmt.Scanf("%s",&user)
		if user=="quit"{
			chw<-"Offline\n"
			return
		}
		info,_:=dbop.FindUser(user)
		if info==nil{
			fmt.Println("Wrong user")
			continue
		}
		fmt.Println("Message:")
		fmt.Scanf("%s",&msg)
		if strings.HasPrefix(msg,"img:"){
			fname:=strings.TrimPrefix(msg,"img:")
			finfo,err:=os.Stat(fname)
			if err!=nil{
				fmt.Println("File not found")
				continue
			}
			output=fmt.Sprintf("Img:%d:%d:%d:%d:%s",gmsgid,info.UID,2,finfo.Size(),fname)
		}else{
			msg,_=convgbk.UTF2GB(msg)
			output=fmt.Sprintf("SendMsg\n%d %d %d %d\n%s\n",gmsgid,info.UID,1,len(msg),msg)
		}
		gmsgid++
		chw<-output
	}
}

func showimg(fname string){
	exec.Command("okular",fname).Run()
	os.Remove(fname)
}

func OnlineWrite(conn net.Conn, chw chan string){
	tm:=time.NewTimer(time.Second*30)
	for{
		select{
		case wr:=<-chw:
			if strings.HasPrefix(wr,"Img:"){
				fields:=strings.Split(wr,":")
				if len(fields)!=6 {
						return
				}
				fname:=fields[5]
				names:=strings.Split(fname,".")
				num:=len(names)
				var imgtype string
				if num>1{
					imgtype=names[num-1]
				}else{
					imgtype="cimg"
				}
				conn.Write([]byte(fmt.Sprintf("SendMsg\n%s %s %s %s\n%s\n",strings.TrimSuffix(fields[1],":"),strings.TrimSuffix(fields[2],":"),
					strings.TrimSuffix(fields[3],":"),strings.TrimSuffix(fields[4],":"),imgtype)))
				fd,_:=os.Open(fname)
				defer fd.Close()
				fsize,_:=strconv.ParseInt(strings.TrimSuffix(fields[4],":"),10,64)
				io.CopyN(conn,fd,fsize)
			}else{
				if _,err:=conn.Write([]byte(wr));err!=nil{
					fmt.Println("write error!")
					return
				}
			}
		case <-tm.C:
			if _,err:=conn.Write([]byte("Heartbeat\n"));err!=nil{
				return
			}
			tm.Reset(time.Second*30)
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
				var mid,mtype,from int
				var mlen int64
				var  tmstamp,dt,tm string;
				fmt.Sscanf(string(info),"%d%d%d%d%s%s",&mid,&mtype,&mlen,&from,&dt,&tm)
				/////
				tmstamp=dt+" "+tm
				switch mtype{
				case dbop.TypeTxt:
					detail,_,err:=brd.ReadLine()
					if err!=nil{
						return
					}
					strdetail,_:=convgbk.GB2UTF(string(detail))
					chmsg<-fmt.Sprintf("Text:%d %s :%s",from,tmstamp,strdetail)
				case dbop.TypePic:
					imgtype,_,err:=brd.ReadLine()
					if err!=nil{
						println("Get image type error")
						return
					}
					fname:=fmt.Sprintf("%s/img-%d.%s",os.TempDir(),time.Now().UnixNano(),string(imgtype))
					fd,_:=os.Create(fname)
					io.CopyN(fd,brd,mlen)
					fd.Close()
					chmsg<-fmt.Sprintf("Text:%d %s :Image",from,tmstamp)
					go showimg(fname)
				}
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
