// LoginDlg.cpp : 实现文件
//

#include "stdafx.h"
#include "chat.h"
#include "LoginDlg.h"
#include "afxdialogex.h"
#include <openssl/sha.h>

// CLoginDlg 对话框
#pragma comment(lib, "libssl.lib")
#pragma comment(lib, "libcrypto.lib")

IMPLEMENT_DYNAMIC(CLoginDlg, CDialogEx)

CLoginDlg::CLoginDlg(CWnd* pParent /*=NULL*/)
	: CDialogEx(CLoginDlg::IDD, pParent)
	, m_strPasswd(_T(""))
	, m_strSvrAddr(_T("123.206.55.31"))
	, m_strUser(_T(""))
	, m_nSelfID(0)
{
	m_bLoginOK=false;
	m_bConn=false;
}

CLoginDlg::~CLoginDlg()
{
/*	if(m_bConn)
		closesocket(m_hConn);*/
}

void CLoginDlg::DoDataExchange(CDataExchange* pDX)
{
	CDialogEx::DoDataExchange(pDX);
	DDX_Text(pDX, IDC_PASSWD, m_strPasswd);
	DDX_Text(pDX, IDC_SVRADDR, m_strSvrAddr);
	DDX_Text(pDX, IDC_USER, m_strUser);
}


BEGIN_MESSAGE_MAP(CLoginDlg, CDialogEx)
	ON_BN_CLICKED(IDOK, &CLoginDlg::OnBnClickedOk)
END_MESSAGE_MAP()


// CLoginDlg 消息处理程序


void CLoginDlg::OnBnClickedOk()
{
	// TODO: 在此添加控件通知处理程序代码
	UpdateData();
	if (!DoLogin()){
	//	AfxMessageBox(_T("Login error"));
		if(m_bConn){
			closesocket(m_hConn);
			m_bConn=false;
		}
		return;
	}
	CDialogEx::OnOK();
}


bool CLoginDlg::DoLogin(void)
{
	if (!m_bConn)
	{
//		m_sConn.Create(0,SOCK_STREAM);
		m_hConn=socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
		m_bConn=true;
	}
	struct sockaddr_in ServerAddr;
	ServerAddr.sin_family = AF_INET;
	ServerAddr.sin_addr.s_addr = inet_addr(m_strSvrAddr.GetBuffer());  
	ServerAddr.sin_port = htons(2048);
	memset(ServerAddr.sin_zero, 0x00, 8);
	if(SOCKET_ERROR ==connect(m_hConn,(struct sockaddr*)&ServerAddr, sizeof(ServerAddr)))
	{
		AfxMessageBox("Can't connect to remove server");
		return false;
	}
	unsigned char res[33]="";
	char szPass[65]="";
	SHA256((const unsigned char*)m_strPasswd.GetBuffer(),m_strPasswd.GetLength(),res);
	m_strPasswd.ReleaseBuffer();
	for(int i=0;i<32;++i)
		sprintf(szPass+i*2,"%02x",res[i]);
	char szSdBuf[1024];
	sprintf(szSdBuf,"Login\n%s\n%s\n",m_strUser.GetBuffer(),szPass);
	m_strUser.ReleaseBuffer();
	send(m_hConn,szSdBuf,strlen(szSdBuf),0);
	char szRdBuf[4096]="";
	int nRecv=recv(m_hConn,szRdBuf,4095,0);
	if(nRecv<=0) {
		AfxMessageBox("Receive login data error");
		return false;
	}
	//szRdBuf[nRecv]='\0';
	if(strncmp(szRdBuf,"OK\n",3)!=0)
		AfxMessageBox(szRdBuf);
	else
	{
		m_bLoginOK=true;
		BUFFER *pBuf=new BUFFER;
		int nLeft=nRecv-3;
		if (nLeft>0)
			strncpy(pBuf->m_arBuf,szRdBuf+3,nLeft);
		pBuf->m_nCap=nLeft;
		m_pBuf=new CRDBuf(m_hConn,pBuf);
		char id[20];
		m_pBuf->ReadLine(id,100);
		m_nSelfID=atoi(id);
	}
	return m_bLoginOK;
}
