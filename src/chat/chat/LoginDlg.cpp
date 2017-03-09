// LoginDlg.cpp : 实现文件
//

#include "stdafx.h"
#include "chat.h"
#include "LoginDlg.h"
#include "afxdialogex.h"

#define SHA256_ROTL(a,b) (((a>>(32-b))&(0x7fffffff>>(31-b)))|(a<<b))
#define SHA256_SR(a,b) ((a>>b)&(0x7fffffff>>(b-1)))
#define SHA256_Ch(x,y,z) ((x&y)^((~x)&z))
#define SHA256_Maj(x,y,z) ((x&y)^(x&z)^(y&z))
#define SHA256_E0(x) (SHA256_ROTL(x,30)^SHA256_ROTL(x,19)^SHA256_ROTL(x,10))
#define SHA256_E1(x) (SHA256_ROTL(x,26)^SHA256_ROTL(x,21)^SHA256_ROTL(x,7))
#define SHA256_O0(x) (SHA256_ROTL(x,25)^SHA256_ROTL(x,14)^SHA256_SR(x,3))
#define SHA256_O1(x) (SHA256_ROTL(x,15)^SHA256_ROTL(x,13)^SHA256_SR(x,10))
extern char* StrSHA256(const char* str, long long length, char* sha256){
    /*
    计算字符串SHA-256
    参数说明：
    str         字符串指针
    length      字符串长度
    sha256         用于保存SHA-256的字符串指针
    返回值为参数sha256
    */
    char *pp, *ppend;
    long l, i, W[64], T1, T2, A, B, C, D, E, F, G, H, H0, H1, H2, H3, H4, H5, H6, H7;
    H0 = 0x6a09e667, H1 = 0xbb67ae85, H2 = 0x3c6ef372, H3 = 0xa54ff53a;
    H4 = 0x510e527f, H5 = 0x9b05688c, H6 = 0x1f83d9ab, H7 = 0x5be0cd19;
    long K[64] = {
        0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
        0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
        0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
        0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
        0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
        0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
        0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
        0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
    };
    l = length + ((length % 64 >= 56) ? (128 - length % 64) : (64 - length % 64));
    if (!(pp = (char*)malloc((unsigned long)l))) return 0;
    for (i = 0; i < length; pp[i + 3 - 2 * (i % 4)] = str[i], i++);
    for (pp[i + 3 - 2 * (i % 4)] = 128, i++; i < l; pp[i + 3 - 2 * (i % 4)] = 0, i++);
    *((long*)(pp + l - 4)) = length << 3;
    *((long*)(pp + l - 8)) = length >> 29;
    for (ppend = pp + l; pp < ppend; pp += 64){
        for (i = 0; i < 16; W[i] = ((long*)pp)[i], i++);
        for (i = 16; i < 64; W[i] = (SHA256_O1(W[i - 2]) + W[i - 7] + SHA256_O0(W[i - 15]) + W[i - 16]), i++);
        A = H0, B = H1, C = H2, D = H3, E = H4, F = H5, G = H6, H = H7;
        for (i = 0; i < 64; i++){
            T1 = H + SHA256_E1(E) + SHA256_Ch(E, F, G) + K[i] + W[i];
            T2 = SHA256_E0(A) + SHA256_Maj(A, B, C);
            H = G, G = F, F = E, E = D + T1, D = C, C = B, B = A, A = T1 + T2;
        }
        H0 += A, H1 += B, H2 += C, H3 += D, H4 += E, H5 += F, H6 += G, H7 += H;
    }
    free(pp - l);
    sprintf(sha256, "%08x%08x%08x%08x%08x%08x%08x%08x", H0, H1, H2, H3, H4, H5, H6, H7);
    return sha256;
}
extern char* FileSHA256(const char* file, char* sha256){
    /*
    计算文件SHA-256
    参数说明：
    file        文件路径字符串指针
    sha256         用于保存SHA-256的字符串指针
    返回值为参数sha256
    */
    FILE* fh;
    char* addlp, T[64];
    long addlsize, j, W[64], T1, T2, A, B, C, D, E, F, G, H, H0, H1, H2, H3, H4, H5, H6, H7;
    long long length, i, cpys;
    void *pp, *ppend;
    H0 = 0x6a09e667, H1 = 0xbb67ae85, H2 = 0x3c6ef372, H3 = 0xa54ff53a;
    H4 = 0x510e527f, H5 = 0x9b05688c, H6 = 0x1f83d9ab, H7 = 0x5be0cd19;
    long K[64] = {
        0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
        0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
        0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
        0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
        0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
        0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
        0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
        0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
    };
    fh = fopen(file, "rb");
    fseek(fh, 0, SEEK_END);
    length = _ftelli64(fh);
    addlsize = (56 - length % 64 > 0) ? (64) : (128);
    if (!(addlp = (char*)malloc(addlsize))) return 0;
    cpys = ((length - (56 - length % 64)) > 0) ? (length - length % 64) : (0);
    j = (long)(length - cpys);
    if (!(pp = (char*)malloc(j))) return 0;
    fseek(fh, -j, SEEK_END);
    fread(pp, 1, j, fh);
    for (i = 0; i < j; addlp[i + 3 - 2 * (i % 4)] = ((char*)pp)[i], i++);
    free(pp);
    for (addlp[i + 3 - 2 * (i % 4)] = 128, i++; i < addlsize; addlp[i + 3 - 2 * (i % 4)] = 0, i++);
    *((long*)(addlp + addlsize - 4)) = length << 3;
    *((long*)(addlp + addlsize - 8)) = length >> 29;
    for (rewind(fh); 64 == fread(W, 1, 64, fh);){
        for (i = 0; i < 64; T[i + 3 - 2 * (i % 4)] = ((char*)W)[i], i++);
        for (i = 0; i < 16; W[i] = ((long*)T)[i], i++);
        for (i = 16; i < 64; W[i] = (SHA256_O1(W[i - 2]) + W[i - 7] + SHA256_O0(W[i - 15]) + W[i - 16]), i++);
        A = H0, B = H1, C = H2, D = H3, E = H4, F = H5, G = H6, H = H7;
        for (i = 0; i < 64; i++){
            T1 = H + SHA256_E1(E) + SHA256_Ch(E, F, G) + K[i] + W[i];
            T2 = SHA256_E0(A) + SHA256_Maj(A, B, C);
            H = G, G = F, F = E, E = D + T1, D = C, C = B, B = A, A = T1 + T2;
        }
        H0 += A, H1 += B, H2 += C, H3 += D, H4 += E, H5 += F, H6 += G, H7 += H;
    }
    for (pp = addlp, ppend = addlp + addlsize; pp < ppend; pp = (long*)pp + 16){
        for (i = 0; i < 16; W[i] = ((long*)pp)[i], i++);
        for (i = 16; i < 64; W[i] = (SHA256_O1(W[i - 2]) + W[i - 7] + SHA256_O0(W[i - 15]) + W[i - 16]), i++);
        A = H0, B = H1, C = H2, D = H3, E = H4, F = H5, G = H6, H = H7;
        for (i = 0; i < 64; i++){
            T1 = H + SHA256_E1(E) + SHA256_Ch(E, F, G) + K[i] + W[i];
            T2 = SHA256_E0(A) + SHA256_Maj(A, B, C);
            H = G, G = F, F = E, E = D + T1, D = C, C = B, B = A, A = T1 + T2;
        }
        H0 += A, H1 += B, H2 += C, H3 += D, H4 += E, H5 += F, H6 += G, H7 += H;
    }
    free(addlp); fclose(fh);
    sprintf(sha256, "%08X%08X%08X%08X%08X%08X%08X%08X", H0, H1, H2, H3, H4, H5, H6, H7);
    return sha256;
}

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
	char res[33]="";
	char szPass[65]="";
	StrSHA256((const char*)m_strPasswd.GetBuffer(),m_strPasswd.GetLength(),szPass);
	m_strPasswd.ReleaseBuffer();
	TRACE("%s\n",szPass);
/*	for(int i=0;i<32;++i)
		sprintf(szPass+i*2,"%02x",res[i]);*/
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
