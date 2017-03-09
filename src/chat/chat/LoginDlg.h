#pragma once
#include "RDBuf.h"

// CLoginDlg 对话框


class CLoginDlg : public CDialogEx
{
	DECLARE_DYNAMIC(CLoginDlg)

public:
	CLoginDlg(CWnd* pParent = NULL);   // 标准构造函数
	virtual ~CLoginDlg();
	bool m_bLoginOK;
	bool m_bConn;
	SOCKET m_hConn;
	CRDBuf *m_pBuf;
// 对话框数据
	enum { IDD = IDD_LOGIN };

protected:
	virtual void DoDataExchange(CDataExchange* pDX);    // DDX/DDV 支持

	DECLARE_MESSAGE_MAP()
public:
	CString m_strPasswd;
	CString m_strUser;
	CString m_strSvrAddr;
	afx_msg void OnBnClickedOk();
	bool DoLogin(void);
	int m_nSelfID;
};
