#pragma once
class CUserInfo
{
public:
	CUserInfo(int nUID,CString strUser="");
	~CUserInfo(void);
	bool UpdateUserInfo();
private:
	CStringList m_lstChatlog;
	CString m_strName;
	int m_nUID;
};

