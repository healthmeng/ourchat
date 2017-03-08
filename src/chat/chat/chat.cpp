
// chat.cpp : ����Ӧ�ó��������Ϊ��
//

#include "stdafx.h"
#include "chat.h"
#include "chatDlg.h"
#include "LoginDlg.h"

#ifdef _DEBUG
#define new DEBUG_NEW
#endif

#pragma comment(lib, "ws2_32.lib")

// CchatApp

BEGIN_MESSAGE_MAP(CchatApp, CWinApp)
	ON_COMMAND(ID_HELP, &CWinApp::OnHelp)
END_MESSAGE_MAP()


// CchatApp ����

CchatApp::CchatApp()
{
	// TODO: �ڴ˴���ӹ�����룬
	// ��������Ҫ�ĳ�ʼ�������� InitInstance ��
}


// Ψһ��һ�� CchatApp ����

CchatApp theApp;


// CchatApp ��ʼ��

BOOL CchatApp::InitInstance()
{
//TODO: call AfxInitRichEdit2() to initialize richedit2 library.
	// ���һ�������� Windows XP �ϵ�Ӧ�ó����嵥ָ��Ҫ
	// ʹ�� ComCtl32.dll �汾 6 ����߰汾�����ÿ��ӻ���ʽ��
	//����Ҫ InitCommonControlsEx()�����򣬽��޷��������ڡ�
	INITCOMMONCONTROLSEX InitCtrls;
	InitCtrls.dwSize = sizeof(InitCtrls);
	// ��������Ϊ��������Ҫ��Ӧ�ó�����ʹ�õ�
	// �����ؼ��ࡣ
	InitCtrls.dwICC = ICC_WIN95_CLASSES;
	InitCommonControlsEx(&InitCtrls);

	CWinApp::InitInstance();
	WSADATA Ws;
	if ( WSAStartup(MAKEWORD(2,2), &Ws) != 0 )
	{
		AfxMessageBox(IDP_SOCKETS_INIT_FAILED);
		return FALSE;
	}

	AfxEnableControlContainer();

	// ���� shell ���������Է��Ի������
	// �κ� shell ����ͼ�ؼ��� shell �б���ͼ�ؼ���
//	CShellManager *pShellManager = new CShellManager;

	// ���Windows Native���Ӿ����������Ա��� MFC �ؼ�����������
//	CMFCVisualManager::SetDefaultManager(RUNTIME_CLASS(CMFCVisualManagerWindows));

	// ��׼��ʼ��
	// ���δʹ����Щ���ܲ�ϣ����С
	// ���տ�ִ���ļ��Ĵ�С����Ӧ�Ƴ�����
	// ����Ҫ���ض���ʼ������
	// �������ڴ洢���õ�ע�����
	// TODO: Ӧ�ʵ��޸ĸ��ַ�����
	// �����޸�Ϊ��˾����֯��
	SetRegistryKey(_T("Ӧ�ó��������ɵı���Ӧ�ó���"));
	INT_PTR nResponse;
	do{
	CLoginDlg ldlg;
	if(ldlg.DoModal()!=IDOK){
		return FALSE;
	}
	CchatDlg dlg(ldlg.m_pBuf,ldlg.m_nSelfID);
//	dlg.SetWindowText(ldlg.m_strUser);
//	m_pMainWnd = &dlg;
	nResponse = dlg.DoModal();

	TRACE("return %d\n",nResponse);
	}while(nResponse!=IDOK);
	if (nResponse == -1)
	{
		TRACE(traceAppMsg, 0, "����: �Ի��򴴽�ʧ�ܣ�Ӧ�ó���������ֹ��\n");
		TRACE(traceAppMsg, 0, "����: ������ڶԻ�����ʹ�� MFC �ؼ������޷� #define _AFX_NO_MFC_CONTROLS_IN_DIALOGS��\n");
	}

	// ɾ�����洴���� shell ��������
/*	if (pShellManager != NULL)
	{
		delete pShellManager;
	}*/

	// ���ڶԻ����ѹرգ����Խ����� FALSE �Ա��˳�Ӧ�ó���
	//  ����������Ӧ�ó������Ϣ�á�
	return FALSE;
}

