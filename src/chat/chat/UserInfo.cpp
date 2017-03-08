#include "stdafx.h"
#include "UserInfo.h"


CUserInfo::CUserInfo(int nUID,CString strUser):m_nUID(nUID)
{
	if(strUser!="")
		UpdateUserInfo();
}

bool CUserInfo::UpdateUserInfo()
{
	return true;
}

CUserInfo::~CUserInfo(void)
{
}
