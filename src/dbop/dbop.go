package dbop

import (
"database/sql"
"log"
"os"
"errors"
"fmt"
_"github.com/Go-SQL-Driver/MySQL"
"time"
)

var db *sql.DB

type UserInfo struct{
	UID int64
	Username string
	Password string // already sha256.Sum256
	Descr string
	Face string
	Phone string
	RegTime string
}

type MsgInfo struct{ // one table for each user
	MsgID int64
	Type int // 1 txt; 2 pic; 3 file
	Content string
	FromUID int // always same
	ToUID int64
	Arrived int
	SvrStamp string
}

func init(){
	var err error
	db,err=sql.Open("mysql","work:Work4All;@tcp(123.206.55.31:3306)/chat")
	if err!=nil{
		log.Println("Open database error:",err)
		os.Exit(1)
	}
}

func (info* UserInfo)GetUnsentMsg()([]MsgInfo,error){
	msgtb:=fmt.Sprintf("msg%d",info.UID)
	query:=fmt.Sprintf("select * from %s where arrived=0",msgtb)
	msgs:=make([]MsgInfo,0,20)
	res,err:=db.Query(query)
	if err!=nil{
		log.Println("Query messages error")
		return nil,err
	}
	for ;res.Next();{
		var msg MsgInfo
		err:=res.Scan(&msg.MsgID,&msg.Type,&msg.FromUID,&msg.ToUID,
					&msg.Arrived,&msg.SvrStamp)
		if err!=nil{
			log.Println("Parse db message error:",err)
			return nil,err
		}
		msgs=append(msgs,msg)
	}
	return msgs,nil
}

func (info* UserInfo)LoadInfo() error{
	dbinfo,_:=FindUser(info.Username)
	if dbinfo!=nil{
		*info=*dbinfo
		return nil
	} else{
		return errors.New("LoadInfo: user not found")
	}
}

func (info* UserInfo)SaveInfo() error{
	dbinfo,_:=FindUser(info.Username)
	if dbinfo!=nil{
	}else{
		return errors.New("SaveInfo: user not found")
	}
	query:=fmt.Sprintf("update users set pwsha256='%s,descr='%s',face='%s',phone='%s' where uid=%d",info.Password,info.Descr,info.Face,info.Phone,info.UID)
	if _,err:=db.Exec(query);err!=nil{
		log.Println("Update db error:",err)
		return err
	}
	return nil
}

func FindUser(username string) (* UserInfo,error){
	query:=fmt.Sprintf("select * from users where username='%s'",username)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("find user query error:",err)
		return nil,err
	}
	if res.Next(){
		info:=new(UserInfo)
		if err:=res.Scan(&info.UID,	&info.Username,
				&info.Password,&info.Descr,&info.Face,
				&info.Phone,&info.RegTime);err!=nil{
			log.Println("Query error:",err)
			return nil,err
		}
		return info,nil
	}
	return nil,nil
}

func AddUser(info *UserInfo) error{
//	return nil
// Add user info in db,add user msg table in db
	if find,_:=FindUser(info.Username);find!=nil{
		return errors.New("User already exists")
	}
	tm:=time.Now().Local()
	info.RegTime=fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second())
	query:=fmt.Sprintf("insert into users (username,pwsha256,descr,face,phone,regtime) values ('%s','%s','%s','%s','%s','%s')",info.Username,info.Password,info.Descr,info.Face,info.Phone,info.RegTime)
	if result,err:=db.Exec(query);err!=nil{
		log.Println("Insert db error:",err)
		return err
	}else{
		info.UID ,_= result.LastInsertId()
		query=fmt.Sprintf("create table `msg%d` (`msgid` int(11) not null AUTO_INCREMENT, `type` smallint(3) not null, `content` varchar(1024), `fromuid` int(11) not null, `touid` int(11) not null, `arrived` tinyint(1) not null, `svrstamp` datetime, PRIMARY KEY(`msgid`))",info.UID)
		if _,err:=db.Exec(query);err!=nil{
			log.Println("Create msg table error:",err)
			return err
		}
		return nil
	}
}

func DelUser(name string, passwd string)error{
	info,err:=FindUser(name)
	if err!=nil{
		log.Println("Del user error:",err)
		return err
	}
	if info==nil{
		return errors.New("User not found")
	}
	if passwd!=info.Password{
		return errors.New("Username/Password is incorrect")
	}
	query:=fmt.Sprintf("delete from users where username='%s'",info.Username)
	if _,err:=db.Exec(query);err!=nil{
		log.Println("Delete user failed:",err)
		return err
	}
	query=fmt.Sprintf("drop table if exists msg%d",info.UID)
	db.Exec(query)
	return nil
}
