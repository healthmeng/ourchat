package dbop

import (
"database/sql"
"log"
"errors"
"fmt"
_"github.com/Go-SQL-Driver/MySQL"
"time"
)

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
	Type int
	Content string
//	FromUID int // always same
	ToUID int64
	Arrived int
	SvrStamp string
}

var dbdrv string="mysql"
var dblogin string="work:abcd1234@tcp(127.0.0.1:3306)/chat"

func FindUser(username string) (* UserInfo,error){
	db,err:=sql.Open(dbdrv,dblogin)
	if err!=nil{
		log.Println("Open database failed")
		return nil,err
	}
	defer db.Close()
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
	db,err:=sql.Open(dbdrv,dblogin)
	if err!=nil{
		log.Println("Open database failed")
		return err
	}
	defer db.Close()
	tm:=time.Now().Local()
	info.RegTime=fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second())
	query:=fmt.Sprintf("insert into users (username,pwsha256,descr,face,phone,regtime) values ('%s','%s','%s','%s','%s','%s')",info.Username,info.Password,info.Descr,info.Face,info.Phone,info.RegTime)
	if result,err:=db.Exec(query);err!=nil{
		log.Println("Insert db error:",err)
		return err
	}else{
		info.UID ,_= result.LastInsertId()
		query=fmt.Sprintf("create table `msg%d` (`msgid` int(11) not null AUTO_INCREMENT, `type` smallint(3) not null, `content` varchar(1024), `touid` int(11) not null, `arrived` tinyint(1) not null, `svrstamp` datetime, PRIMARY KEY(`msgid`))",info.UID)
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
	db,err:=sql.Open(dbdrv,dblogin)
	if err!=nil{
		log.Println("Open database failed")
		return err
	}
	defer db.Close()
	query:="delete from users where username="+name
	if _,err:=db.Exec(query);err!=nil{
		log.Println("Delete user failed:",err)
		return err
	}
	query=fmt.Sprintf("drop table if exists msg%d",info.UID)
	db.Exec(query)
	return nil
}
