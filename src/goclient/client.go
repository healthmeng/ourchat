package main

import (
"net"
"fmt"
"bufio"
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
		fmt.Println("Connect to server failed")
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
		fmt.Println("Connect to server failed")
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
		fmt.Println("Connect to server failed")
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
	chw:=make(chan string,10)
	go OnlineRead(brd,chw)
	go OnlineWrite(conn,chw)
	for{
	}
}

func OnlineRead(brd *bufio.Reader, chw chan string){
	for{
		if buf,_,err:=brd.ReadLine();err==nil{
			switch string(buf){
			case "SendMsg":
				info,_,err:=brd.ReadLine()
				var mid,mtype,mlen int
				var tm msg string;
				fmt.Sscanf(string(info),"%d%d%d%s",&mid,&mtype,mlen)
			
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
