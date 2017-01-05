package main

import (
"net"
"fmt"
"bufio"
"crypto/sha256"
"encoding/json"
"os/exec"
"os"
// "compress/gzip"
)

var svraddr string = "123.206.55.31"
var svrport string = ":2048"


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

	conn,err:=net.Dial("tcp",svraddr+svrport)
	if err!=nil{
		fmt.Println("Connect to server failed")
		return
	}
	defer conn.Close()
	addtext:="AddUser\n"
	obj,err:=json.Marshal(info);
	if err!=nil{
		fmt.Println("json failed")
		return
	}
	addtext+=string(append(obj,'\n'))
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
}

func doLogin(){
}

func main(){
	if len(os.Args)==1{
		fmt.Println("client register\nclient del\nclient login")
		return
	}
	if os.Args[1]=="register"{
		doRegister()
	} else if os.Args[1]=="del" {
		doDel()
	} else if os.Args[1]=="login"{
		doLogin()
	}else{
		fmt.Println("client register\nclient del\nclient login")
	}
}
