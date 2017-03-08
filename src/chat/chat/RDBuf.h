#pragma once

struct BUFFER{
	char m_arBuf[4096];
	int m_nCap;
	BUFFER(){
		m_nCap=0;
	}
};

class CRDBuf
{
public:
	CRDBuf(SOCKET hConn,BUFFER* strInit=NULL);

	// blocked, or return true for successly read a line or nMax bytes, or read error on EOF
	int ReadLine(char *szBuf,int nMax); 	// return <=0 , error or closed
	
	int Read(char* szBuf,int nBytes);
	int ReadNext();
	~CRDBuf(void);
	BOOL TryGetLine(char *szBuf,int *pEnd,int nMax);
	SOCKET m_hConn;
	time_t m_nLastRead;
private:
	CList<BUFFER *> m_lstBuffer;
	int m_nCurChar;

};

