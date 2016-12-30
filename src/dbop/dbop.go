package dbop

type UserInfo struct{
	UID int
	Username string
	Password string
	Descr string
	Face string
	Phone string
}

func AddUser(info *UserInfo) error{
	return nil
}

