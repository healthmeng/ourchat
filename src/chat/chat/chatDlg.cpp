
// chatDlg.cpp : 实现文件
//

#include "stdafx.h"
#include "chat.h"
#include "chatDlg.h"
#include "afxdialogex.h"
#include <stdio.h>
#include <time.h>

#ifdef _DEBUG
#define new DEBUG_NEW
#endif

#define BMP_FACE0 9000
#define BMP_FACEBW0	9100

// CchatDlg 对话框

UINT readproc(LPVOID param)
{
	CchatDlg* pDlg=(CchatDlg*) param;
	while(1)
	{
		char buf[4096];
		int nRet=pDlg->m_pConn->ReadLine(buf,4095);
		if(nRet<0){
			TRACE("Read failed\n");
			break;
		}else{
			if(strcmp(buf,"Heartbeat")==0){
				pDlg->PostMessage(WM_HBCOME,NULL,NULL);
			}
			else if(strcmp(buf,"Users")==0){
				nRet=pDlg->m_pConn->ReadLine(buf,4095);
				if(nRet<0) break;
				char *p=new char[strlen(buf)+1]; 
				strcpy(p,buf);
				pDlg->PostMessage(WM_UPDATEOLUSER,(WPARAM)p,NULL);
			}
			else if(strcmp(buf,"UserList")==0){
				nRet=pDlg->m_pConn->ReadLine(buf,4095);
				if(nRet<0) break;
				int nCnt;
				sscanf(buf,"%d",&nCnt);
				char** ppUsr=(char**)malloc(sizeof(char*)*nCnt);
				for(int i=0;i<nCnt;i++){
					nRet=pDlg->m_pConn->ReadLine(buf,4095);
					if(nRet<0) break;
					ppUsr[i]=(char*)malloc(strlen(buf)+1);
					strcpy(ppUsr[i],buf);
					//TRACE("%s\n",buf);
				}
				pDlg->PostMessage(WM_UPDATEUINFO,(WPARAM)nCnt,(LPARAM)ppUsr);
			}
			else if(strcmp(buf,"SendMsg")==0){
				nRet=pDlg->m_pConn->ReadLine(buf,4095);
				if(nRet<0)
					break;
				CMsgContent* pMsg=new CMsgContent;
				char szDate[20],szTime[20];
				sscanf(buf,"%d%d%d%d%s%s",&pMsg->m_nMsgID,&pMsg->m_nType,&pMsg->m_nLen,&pMsg->m_nFrom,szDate,szTime);
				sprintf_s(pMsg->m_szTime,50,"%s %s",szDate,szTime);
				pMsg->m_pBuf=new char[pMsg->m_nLen+1];
				if(pMsg->m_nType==1){
					nRet=pDlg->m_pConn->ReadLine(pMsg->m_pBuf,pMsg->m_nLen);
					if(nRet<0){
						delete pMsg;
						break;
					}
					pMsg->m_nTo=pDlg->m_nSelfID;
				}
				pDlg->PostMessage(WM_MSGCOME,(WPARAM)pMsg,NULL);
			}
		}
	}
	if(!pDlg->m_bDisconn) 
	// if logout already, thread should been Waited in ~pDlg,
		// so pDlg should always be valid
		pDlg->PostMessage(WM_LOGOUT,NULL,NULL);
	return 0;
}

CchatDlg::CchatDlg(CRDBuf* pConn, int nUID,CWnd* pParent /*=NULL*/)
	: CDialogEx(CchatDlg::IDD, pParent),m_pConn(pConn),m_nSelfID(nUID)
{
	AfxInitRichEdit2();
	m_hIcon = AfxGetApp()->LoadIcon(IDR_MAINFRAME);
	m_nCurChatID=-1;
	m_pSelf=NULL;
	m_mapCurChatLog.InitHashTable(256,1);
	m_mapUsers.InitHashTable(256);
	m_nMsgID=0;
	m_bDisconn=false;
}

void CchatDlg::DoDataExchange(CDataExchange* pDX)
{
	CDialogEx::DoDataExchange(pDX);
	DDX_Control(pDX, IDC_USRLST, m_cUserList);
	DDX_Control(pDX, IDC_EDIT1, m_cCurInput);
	DDX_Control(pDX, IDC_RICHEDIT22, m_cCurChatLog);
	DDX_Control(pDX, IDC_SEND, m_cSend);
	DDX_Control(pDX, IDC_CURCHAT, m_cCurChat);
}

BEGIN_MESSAGE_MAP(CchatDlg, CDialogEx)
	ON_WM_PAINT()
	ON_WM_QUERYDRAGICON()
	ON_BN_CLICKED(ID_LOGOUT, &CchatDlg::OnBnClickedLogout)
	ON_WM_TIMER()
	ON_MESSAGE(WM_SENDBACK, &CchatDlg::OnSendback)
	ON_MESSAGE(WM_MSGCOME, &CchatDlg::OnMsgcome)
	ON_MESSAGE(WM_HBCOME, &CchatDlg::OnHbcome)
	ON_WM_CREATE()
	ON_MESSAGE(WM_UPDATEOLUSER, &CchatDlg::OnUpdateoluser)
	ON_MESSAGE(WM_UPDATEUINFO, &CchatDlg::OnUpdateuinfo)
	ON_NOTIFY(NM_DBLCLK, IDC_USRLST, &CchatDlg::OnNMDblclkUsrlst)
	ON_NOTIFY(NM_CLICK, IDC_USRLST, &CchatDlg::OnNMClickUsrlst)
	ON_WM_KEYDOWN()
	ON_BN_CLICKED(IDC_SEND, &CchatDlg::OnBnClickedSend)
	ON_MESSAGE(WM_LOGOUT, &CchatDlg::OnReadError)
	ON_COMMAND(IDOK, &CchatDlg::OnIdok)
	ON_COMMAND(IDCANCEL, &CchatDlg::OnIdcancel)
END_MESSAGE_MAP()


// CchatDlg 消息处理程序

BOOL CchatDlg::OnInitDialog()
{
	CDialogEx::OnInitDialog();

	// 设置此对话框的图标。当应用程序主窗口不是对话框时，框架将自动
	//  执行此操作
	SetIcon(m_hIcon, TRUE);			// 设置大图标
	SetIcon(m_hIcon, FALSE);		// 设置小图标

	// TODO: 在此添加额外的初始化代码
	LONG lStyle;
	lStyle=GetWindowLong(m_cUserList.m_hWnd,GWL_STYLE);
	lStyle &= ~LVS_TYPEMASK;
	lStyle |=LVS_LIST;
	SetWindowLong(m_cUserList.m_hWnd,GWL_STYLE, lStyle);
	m_imgList.Create(32,32,ILC_MASK|ILC_COLOR24,6,6);
	//m_imgList.Add(theApp.LoadIcon(IDR_MAINFRAME));
	//m_imgList.Add(theApp.LoadIcon(IDI_ICON1));
	for(int i=0;i<6;i++){
		HBITMAP hbmp=(HBITMAP)LoadImage(AfxGetInstanceHandle(),
			MAKEINTRESOURCE(IDB_FACEBW0+i) ,IMAGE_BITMAP,32,32,
			LR_CREATEDIBSECTION);//| LR_LOADFROMFILE);
		CBitmap bmp;
		bmp.Attach(hbmp);
		m_imgList.Add(&bmp,RGB(255,255,255));
	}
	for(int i=0;i<6;i++){
		HBITMAP hbmp=(HBITMAP)LoadImage(AfxGetInstanceHandle(),
			MAKEINTRESOURCE(IDB_FACE0+i) ,IMAGE_BITMAP,32,32,
			LR_CREATEDIBSECTION);//| LR_LOADFROMFILE);
		CBitmap bmp;
		bmp.Attach(hbmp);
		m_imgList.Add(&bmp,RGB(255,255,255));
	}
	m_cUserList.SetImageList(&m_imgList,LVSIL_SMALL );
	
	m_cUserList.SetExtendedStyle(LVS_EX_FULLROWSELECT);
	m_cSend.EnableWindow(FALSE);
	m_cCurChat.SetWindowText(_T("当前对话：无"));
//	m_pWR=AfxBeginThread(writeproc,this);
	return TRUE;  // 除非将焦点设置到控件，否则返回 TRUE
}

// 如果向对话框添加最小化按钮，则需要下面的代码
//  来绘制该图标。对于使用文档/视图模型的 MFC 应用程序，
//  这将由框架自动完成。

void CchatDlg::OnPaint()
{
	if (IsIconic())
	{
		CPaintDC dc(this); // 用于绘制的设备上下文

		SendMessage(WM_ICONERASEBKGND, reinterpret_cast<WPARAM>(dc.GetSafeHdc()), 0);

		// 使图标在工作区矩形中居中
		int cxIcon = GetSystemMetrics(SM_CXICON);
		int cyIcon = GetSystemMetrics(SM_CYICON);
		CRect rect;
		GetClientRect(&rect);
		int x = (rect.Width() - cxIcon + 1) / 2;
		int y = (rect.Height() - cyIcon + 1) / 2;

		// 绘制图标
		dc.DrawIcon(x, y, m_hIcon);
	}
	else
	{
		CDialogEx::OnPaint();
	}
}

//当用户拖动最小化窗口时系统调用此函数取得光标
//显示。
HCURSOR CchatDlg::OnQueryDragIcon()
{
	return static_cast<HCURSOR>(m_hIcon);
}

void CchatDlg::OnBnClickedCancel()
{
	// TODO: 在此添加控件通知处理程序代码

}

void CchatDlg::OnBnClickedLogout()
{
	// TODO: 在此添加控件通知处理程序代码
	// do logout
	char strOffline[]="Offline\n";
	send(m_pConn->m_hConn,strOffline,strlen(strOffline),0);
	m_bDisconn=true;
	OnOK();
}

CchatDlg::~CchatDlg(void)
{
	closesocket(m_pConn->m_hConn);
	 ::WaitForSingleObject(m_pRD->m_hThread, INFINITE);
	// ::WaitForSingleObject(m_pWR->m_hThread, INFINITE);
	 ClearUsers();
	 delete m_pConn;

}

void CchatDlg::DoReconn(){
	char strOffline[]="Offline\n";
	send(m_pConn->m_hConn,strOffline,strlen(strOffline),0);
	m_bDisconn=true;
	OnCancel();
}

void CchatDlg::OnTimer(UINT_PTR nIDEvent)
{
	// TODO: 在此添加消息处理程序代码和/或调用默认值

	if(nIDEvent==1) // heartbeat process
	{
		time_t tm=time(NULL);
		if(tm-m_pConn->m_nLastRead >90)
			DoReconn();
		char hb[]="Heartbeat\n";
		send(m_pConn->m_hConn,hb,strlen(hb),0);
	}else if(nIDEvent>=100){
		CChatter* pUser=m_arUsers[nIDEvent-100];
		if(pUser){
			if(m_cUserList.GetItemText(pUser->m_nIndex,0)==""){
				m_cUserList.SetItemText(pUser->m_nIndex,0,pUser->m_szName);
			}else{
				m_cUserList.SetItemText(pUser->m_nIndex,0,"");				
			}
		}
	}
	CDialogEx::OnTimer(nIDEvent);
}

afx_msg LRESULT CchatDlg::OnSendback(WPARAM wParam, LPARAM lParam)
{
	return 0;
}

afx_msg LRESULT CchatDlg::OnMsgcome(WPARAM wParam, LPARAM lParam)
{
	CMsgContent* pMsg=(CMsgContent*)wParam; // todo: add to list or delete!
	char buf[1024];
	sprintf(buf,"Confirm\n%d\n",pMsg->m_nMsgID);
	send(m_pConn->m_hConn,buf,strlen(buf),0);
	
	CChatter *pUser=FindUser(pMsg->m_nFrom);
	if(pUser){
		pUser->m_arMsgLog.Add(pMsg);
		// show msg...
	/*	int nSel=m_cUserList.GetNextItem(-1,LVIS_SELECTED);
		CChatter* pCurUser=NULL;
		if(nSel>=0){
			pUser=m_arUsers[nSel];
		}*/
		if(m_nCurChatID==pUser->m_nUID){
			char show[4096];
			sprintf_s(show,4095,"%s(%s):\n%s\n",pUser->m_szName,pMsg->m_szTime,pMsg->m_pBuf);
			m_cCurChatLog.SetSel(-1,-1);
			m_cCurChatLog.ReplaceSel(show);
		}else{
			KillTimer(100+pUser->m_nIndex);
			SetTimer(100+pUser->m_nIndex,750,NULL);
		}
		TRACE("msg: from %s(%d) -- %s\n",pUser->m_szName,pUser->m_nUID,pMsg->m_pBuf);
	}else{
		// show msg...
		delete pMsg;
	}
	return 0;
}

afx_msg LRESULT CchatDlg::OnHbcome(WPARAM wParam, LPARAM lParam)
{
	time_t tm=time(NULL);
//	TRACE("Get heartbeat from server:%s\n",ctime(&tm));
	return 0;
}

int CchatDlg::OnCreate(LPCREATESTRUCT lpCreateStruct)
{
	if (CDialogEx::OnCreate(lpCreateStruct) == -1)
		return -1;

	m_pRD=AfxBeginThread(readproc,this);
	//const char *uinfo="GetUserInfo\n";
	//send(m_pConn->m_hConn, uinfo, strlen(uinfo),0);
	SetTimer(1,30*1000,NULL);
	//SetTimer(2,90*1000,NULL);
	return 0;
}

afx_msg LRESULT CchatDlg::OnUpdateoluser(WPARAM wParam, LPARAM lParam)
{
	char *p=(char*)wParam;
	TRACE("online user: %s\n",p);
	int nAll=m_arUsers.GetCount();
	for(int i=0;i<nAll;i++)
		m_arUsers[i]->m_bOnline=false;
    CArray<CString> fields;
    CString strTmp=strtok(p,"|");//(LPSTR)(LPCTSTR)将CString转char*
    while(1)
    {    
		fields.Add(strTmp);
        char *pRet=strtok(NULL,";");
        if (pRet==NULL)              
            break;
		strTmp=pRet;
        strTmp.TrimLeft(); 
	}
	int nCnt=fields.GetCount();
	for(int i=0;i<nCnt;i++){
		unsigned int uid=0;
		sscanf((const char*) fields[i],"%d",&uid);
		CChatter* pUser=NULL;
		BOOL bFind=m_mapUsers.Lookup(uid,pUser);
		if(bFind){
			pUser->m_bOnline=true;
		}// todo add timeout offline
	}
	delete[]p;
	for (int i=0;i<nAll;i++){
		CChatter *pUser=m_arUsers[i];
		LV_ITEM lvi;
		lvi.mask     = LVIF_IMAGE;
		lvi.iImage   = pUser->m_nFace+6*(pUser->m_bOnline?1:0);
		lvi.iItem    = i;
		lvi.iSubItem = 0;
		m_cUserList.SetItem(&lvi);
	}
	return 0;
}


BOOL CchatDlg::GetUserDetail(void)
{
	return 0;
}

CChatter* CchatDlg::ParseUserInfo(char *desc)
{
    CArray<CString> fields;
    CString strTmp=strtok(desc,";");//(LPSTR)(LPCTSTR)将CString转char*
    while(1)
    {    
		fields.Add(strTmp);
        char *pRet=strtok(NULL,";");
        if (pRet==NULL)              
            break;
		strTmp=pRet;
        strTmp.TrimLeft(); 
	}
	int nCnt;
	if((nCnt=fields.GetCount())<2)
		return NULL;
	CChatter* pUser=new CChatter;
	sscanf((const char*) fields[0],"%d",&pUser->m_nUID);
	strcpy(pUser->m_szName,fields[1]);
	pUser->m_bOnline=false;
	if(nCnt>2)
		strcpy(pUser->m_szDescr,fields[2]);
	if(nCnt>3)
		strcpy(pUser->m_szFace,fields[3]);
	if(nCnt>4)
		strcpy(pUser->m_szPhone,fields[4]);
	return pUser;
}

afx_msg LRESULT CchatDlg::OnUpdateuinfo(WPARAM wParam, LPARAM lParam)
{
	int nCnt=(int)wParam;
	char **msg=(char**)lParam;
	ClearUsers();
	m_cUserList.DeleteAllItems();
	int index=0;
	for(int i=0;i<nCnt;i++){
		CChatter* pUser=ParseUserInfo(msg[i]);
		pUser->m_nFace=i%6; // todo: according to received data in ParseUserInfo
		if(pUser){
			pUser->ReloadMsgs();

			CString strUser=pUser->m_szName;
			if(pUser->m_szDescr[0]!='\0')
				strUser+=pUser->m_szDescr;
			if (pUser->m_nUID!=m_nSelfID)
			{
				m_cUserList.InsertItem(i,strUser,pUser->m_nFace);
				m_arUsers.Add(pUser);
				pUser->m_nIndex=index++;
			}else{
				m_pSelf=pUser;
				SetWindowText(pUser->m_szName);
			}
			m_mapUsers[pUser->m_nUID]=pUser;
			TRACE("%d->%s\n",pUser->m_nUID,pUser->m_szName);
		}else{
			TRACE("Parse user error:%s\n",msg[i]);
		}
		delete[] msg[i];
	}
	delete[] msg;
	return 0;
}

void CchatDlg::ClearUsers(){
	int nCnt=m_arUsers.GetCount();
	for(int i=0;i<nCnt;i++){
		delete m_arUsers[i];
	}
	m_arUsers.RemoveAll();
	if(m_pSelf)
	{
		delete m_pSelf;
		m_pSelf=NULL;
	}
	m_mapUsers.RemoveAll();
}


CChatter* CchatDlg::FindUser(int nID)
{
	int nCnt=m_arUsers.GetCount();
	for(int i=0;i<nCnt;i++){
		if(m_arUsers[i]->m_nUID==nID)
			return m_arUsers[i];
	}
	return NULL;
}


void CchatDlg::ReloadChatLog(){
	// find original index first, if exists, keep state
	// then change original ChatID
	CString strInput;
	if(m_nCurChatID>=0){
		m_cCurInput.GetWindowText(strInput);
		m_mapCurChatLog[m_nCurChatID]=strInput;
	}
	int nSel=m_cUserList.GetNextItem(-1,LVIS_SELECTED);
	ASSERT(nSel>=0);
	CChatter *pUser=m_arUsers[nSel];
	m_nCurChatID=pUser->m_nUID;

	BOOL bFind=m_mapCurChatLog.Lookup(m_nCurChatID,strInput);
	m_cCurInput.SetWindowText(bFind?LPCTSTR(strInput):"");
	// finish current chat edit

	// reload chat history:
	m_cCurChatLog.SetWindowText("");
	int nLogCnt=pUser->m_arMsgLog.GetCount();
	for(int i=0;i<nLogCnt;i++)
	{
		CMsgContent* pMsg=m_arUsers[nSel]->m_arMsgLog[i];
		char show[4096];
		if(pMsg->m_nFrom==m_nSelfID)
			sprintf_s(show,4095,"我(%s):\n%s\n",pMsg->m_szTime,pMsg->m_pBuf);
		else
		{
			sprintf_s(show,4095,"%s(%s):\n%s\n",pUser->m_szName,pMsg->m_szTime,pMsg->m_pBuf);
		}
		m_cCurChatLog.SetSel(-1,-1);
		m_cCurChatLog.ReplaceSel(show);
	}
}

void CchatDlg::OnNMDblclkUsrlst(NMHDR *pNMHDR, LRESULT *pResult)
{
	LPNMITEMACTIVATE pNMItemActivate = reinterpret_cast<LPNMITEMACTIVATE>(pNMHDR);

	// TODO: 在此添加控件通知处理程序代码
	*pResult = 0;
}


void CchatDlg::OnNMClickUsrlst(NMHDR *pNMHDR, LRESULT *pResult)
{
	LPNMITEMACTIVATE pNMItemActivate = reinterpret_cast<LPNMITEMACTIVATE>(pNMHDR);
	// TODO: 在此添加控件通知处理程序代码
	int nSel=m_cUserList.GetNextItem(-1,LVIS_SELECTED);
	if(nSel>=0) {
		m_cSend.EnableWindow(TRUE);
		CChatter* pUser=m_arUsers[nSel];
		if(pUser->m_nUID!=m_nCurChatID)
		{
		//	m_nCurChatID=m_arUsers[nSel]->m_nUID;
			ReloadChatLog();
			KillTimer(pUser->m_nIndex+100);
			m_cUserList.SetItemText(pUser->m_nIndex,0,pUser->m_szName);
			m_cCurChat.SetWindowText(CString("当前对话：")+pUser->m_szName);
		}
	}
	
	*pResult = 0;
}

void CchatDlg::OnKeyDown(UINT nChar, UINT nRepCnt, UINT nFlags)
{
	// TODO: 在此添加消息处理程序代码和/或调用默认值
	TRACE("on key down");
	if(nChar==VK_RETURN &&(::GetKeyState(VK_CONTROL)<0))
	{
		TRACE("ctrl + enter");
		return ;
	}
	CDialogEx::OnKeyDown(nChar, nRepCnt, nFlags);
}


void CchatDlg::OnBnClickedSend()
{
	/*
	1. Get current chatter & text
	2. send text
	3. if send success add to logwindow and add to log array
	*/
	// TODO: 在此添加控件通知处理程序代码
	if(!m_cSend.IsWindowEnabled()) // actually, impossible
		return;
	char buf[4096];
	m_cCurInput.GetWindowText(buf,4095);
	if(buf[0]=='\0'){
		AfxMessageBox("Please input something first");
		return ;
	}
	CChatter* pUser=FindUser(m_nCurChatID);
	if(!pUser){
		return;
	}
	char sendcmd[4096];
	int nMsg=strlen(buf);
	sprintf(sendcmd,"SendMsg\n%ld %d %d %d\n%s\n",m_nMsgID++,pUser->m_nUID,1,nMsg,buf);
	int ret=send(m_pConn->m_hConn,sendcmd,strlen(sendcmd),0);
	if (ret>0){
		CMsgContent* pMsg=new CMsgContent;
		pMsg->m_nMsgID=-1;// self
		pMsg->m_nType=1;
		pMsg->m_nLen=nMsg;
		pMsg->m_nFrom=m_nSelfID;
		pMsg->m_nTo=pUser->m_nUID;
		SYSTEMTIME tm;
		GetLocalTime(&tm);
		sprintf_s(pMsg->m_szTime,50,"%d%02d%02d %02d:%02d:%02d",
			tm.wYear,tm.wMonth,tm.wDay,tm.wHour,tm.wMinute,tm.wSecond);
		pMsg->m_pBuf=new char[pMsg->m_nLen+1];
		strcpy(pMsg->m_pBuf,buf);		
		pUser->m_arMsgLog.Add(pMsg);

		char show[4096];
		sprintf_s(show,4095,"我(%s):\n%s\n",pMsg->m_szTime,pMsg->m_pBuf);
		m_cCurChatLog.SetSel(-1,-1);
		m_cCurChatLog.ReplaceSel(show);
		m_cCurInput.SetWindowText("");
	}
}
// todo: timeout reconnect

afx_msg LRESULT CchatDlg::OnReadError(WPARAM wParam, LPARAM lParam)
{
	DoReconn();
	return 0;
}


void CchatDlg::OnIdok()
{
	// TODO: 在此添加命令处理程序代码
	DoReconn();
}


void CchatDlg::OnIdcancel()
{
	// TODO: 在此添加命令处理程序代码
	OnBnClickedLogout();
}


BOOL CchatDlg::PreTranslateMessage(MSG* pMsg)
{
	// TODO: 在此添加专用代码和/或调用基类
	 if(pMsg->message==WM_KEYDOWN) {
		if(pMsg->wParam==VK_RETURN)
		{
			OnBnClickedSend();
			return true;
		/*	if(::GetKeyState(VK_CONTROL)<0)
			{
				
			}
			else{
				if(GetFocus()->m_hWnd==m_cCurInput.m_hWnd){
					m_cCurInput.SendMessage(WM_KEYUP,VK_RETURN,0);
					return true;
				}
			}*/
		}
	 }
	return CDialogEx::PreTranslateMessage(pMsg);
}
