
// chatDlg.h : 头文件
//

#pragma once

#include "RDBuf.h"
#include "afxcmn.h"
#include "afxwin.h"
// CchatDlg 对话框

class CMsgContent{
public:
	CMsgContent(){
		m_pBuf=NULL;
	}
	void SetContent(const char* pOrgMsg){
		if(m_pBuf){
			delete [] m_pBuf;
		}
		m_pBuf=new char[strlen(pOrgMsg)+1];
		strcpy(m_pBuf,pOrgMsg);
	}
	~CMsgContent(){ 
		if (m_pBuf) 
			delete[] m_pBuf;
	}
	int m_nType;
	long m_nMsgID;
	long m_nLen;
	int m_nFrom;
	int m_nTo;
	char m_szTime[50];
	char* m_pBuf;
};

class CChatter{
public:
	CChatter(){
		m_szName[0]=0;
		m_szDescr[0]=0;
		m_szFace[0]=0;
		m_szPhone[0]=0;
		m_nFace=0;
		m_bOnline=false;
	}
	~CChatter() { 
		int nLen=m_arMsgLog.GetCount();
		for(int i=0;i<nLen;i++){
			delete m_arMsgLog[i];
		}
	}
	char m_szName[128];
	unsigned long m_nUID;
	char m_szDescr[256];
	char m_szFace[512];
	char m_szPhone[50];
	int m_nFace;
	int m_nIndex;
	CArray<CMsgContent*> m_arMsgLog;
	CString m_strCurInput;
	bool m_bOnline;
	void ReloadMsgs(){
		char szFile[4096];
		FILE *fp=fopen(itoa(m_nUID,szFile,10),"r");
		if(fp){
			while(!feof(fp)){
				char buf[4096];
				fgets(buf,4095,fp);
				CMsgContent* pMsg=new CMsgContent;
				pMsg->SetContent(buf);
				m_arMsgLog.Add(pMsg);
			}
			fclose(fp);
		}else
			return ;
	}
};

class CchatDlg : public CDialogEx
{
// 构造
public:
	CchatDlg(CRDBuf* m_pConn,int nUID,CWnd* pParent = NULL);	// 标准构造函数
	CRDBuf *m_pConn;
	int m_nSelfID;
	int m_nCurChatID;
	bool m_bDisconn;
	unsigned long m_nMsgID;
	CChatter *m_pSelf;
	CMap<int,int,CString,CString> m_mapCurChatLog;
	CMap<UINT,UINT,CChatter*,CChatter*> m_mapUsers;
// 对话框数据
	enum { IDD = IDD_CHAT_DIALOG };

	protected:
	virtual void DoDataExchange(CDataExchange* pDX);	// DDX/DDV 支持

// 实现
protected:
	HICON m_hIcon;
	CWinThread *m_pRD, *m_pWR;
	CImageList m_imgList;
	// 生成的消息映射函数
	virtual BOOL OnInitDialog();
	void DoReconn();
	afx_msg void OnPaint();
	afx_msg HCURSOR OnQueryDragIcon();
	DECLARE_MESSAGE_MAP()
public:
	afx_msg void OnBnClickedCancel();
	afx_msg void OnBnClickedLogout();
	virtual ~CchatDlg(void);
	afx_msg void OnTimer(UINT_PTR nIDEvent);
protected:
	afx_msg LRESULT OnSendback(WPARAM wParam, LPARAM lParam);
	afx_msg LRESULT OnMsgcome(WPARAM wParam, LPARAM lParam);
	afx_msg LRESULT OnHbcome(WPARAM wParam, LPARAM lParam);
	CChatter* ParseUserInfo(char *desc);
public:
	afx_msg int OnCreate(LPCREATESTRUCT lpCreateStruct);
protected:
	afx_msg LRESULT OnUpdateoluser(WPARAM wParam, LPARAM lParam);
public:
	BOOL GetUserDetail(void);
	void ReloadChatLog();
	CListCtrl m_cUserList;
	CArray<CChatter*> m_arUsers;
	void ClearUsers();
public:
	afx_msg LRESULT OnUpdateuinfo(WPARAM wParam, LPARAM lParam);
	CChatter* FindUser(int nID);
	afx_msg void OnNMDblclkUsrlst(NMHDR *pNMHDR, LRESULT *pResult);
	CEdit m_cCurInput;
	CRichEditCtrl m_cCurChatLog;
	afx_msg void OnNMClickUsrlst(NMHDR *pNMHDR, LRESULT *pResult);
	afx_msg void OnKeyDown(UINT nChar, UINT nRepCnt, UINT nFlags);
	afx_msg void OnBnClickedSend();
	CButton m_cSend;
protected:
	afx_msg LRESULT OnReadError(WPARAM wParam, LPARAM lParam);
public:
	afx_msg void OnIdok();
	afx_msg void OnIdcancel();
	virtual BOOL PreTranslateMessage(MSG* pMsg);
	CStatic m_cCurChat;
};
