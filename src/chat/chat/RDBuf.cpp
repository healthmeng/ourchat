#include "stdafx.h"
#include "RDBuf.h"

#include "chat.h"
#include "chatDlg.h"
#include "LoginDlg.h"
CRDBuf::CRDBuf(SOCKET hSocket,BUFFER *pInit)
{
	
	m_hConn=hSocket;
	m_nLastRead=time(NULL);
	m_nCurChar=0;
	if(!pInit)
		pInit=new BUFFER;
	m_lstBuffer.AddTail(pInit);
}

// -1 for get \n  or  buffer full, anyway ,finished, >=0 for not finish,return read bytes
int CRDBuf::TryGetLine(char *szBuf,int *pEnd,int nMax) // get line in current buffer
{
	int nBuf=m_lstBuffer.GetCount();
	if(nBuf<=0)
	{
		return 0;
	}
	int nPageUsed=0;
	int nCopied=0;
	bool bFind=false;
	POSITION pos=m_lstBuffer.GetHeadPosition();
	BUFFER *buf=NULL;
	for(int i=0;i<nBuf && !bFind;i++)
	{
		buf=m_lstBuffer.GetNext(pos);
		int j;
		if(nPageUsed>0){
			m_nCurChar=j=0;
		}else{
			j=m_nCurChar;
		}
		for(;j<buf->m_nCap;j++){
			if(buf->m_arBuf[j]=='\n'){
				m_nCurChar=++j;
				bFind=true;
				break;
			}else{
				szBuf[nCopied]=buf->m_arBuf[j]; // copy byte
				++*pEnd;	// nBytes copied total
				nCopied++;	// copied in this function
				m_nCurChar++;	// point to the first available byte in first page
				if(nCopied>=nMax){
					bFind=true;
					j++;
					break;
				}
			}
		}
		if (j==buf->m_nCap) 
			nPageUsed++;
	}
	if (buf && m_nCurChar>=buf->m_nCap)
		m_nCurChar=0; // already end of this page buffer, point to next
	for(int i=0;i<nPageUsed;i++)
	{
		BUFFER* tmp=m_lstBuffer.RemoveHead();
		delete tmp;
	}
	if(bFind)
		return -1;	
	return nCopied;
}

int CRDBuf::ReadLine(char *szBuf,int nMax) // max chars, not include \0
{
	bool rdfail=false;
	int nBufEnd=0;
	int nRead;
	while((nRead=TryGetLine(szBuf,&nBufEnd,nMax))>=0 && !rdfail) {
		nBufEnd+=nRead;
		nMax-=nRead;
		if(ReadNext()<=0)
			rdfail=true;
		
	}
	if (nBufEnd==0 && rdfail) 
		return -1;
	szBuf[nBufEnd]='\0';
	m_nLastRead=time(NULL);
	return nBufEnd;
}

int CRDBuf::Read(char* szBuf,int nBytes)
{


	return 0;
}

int CRDBuf::ReadNext() // read one page
{
	ASSERT(m_lstBuffer.IsEmpty());
	BUFFER* pBuf=new BUFFER;
	int nRead= recv(m_hConn,pBuf->m_arBuf,4096,0);
	if(nRead>0)
	{
		pBuf->m_nCap=nRead;
		m_lstBuffer.AddTail(pBuf);
	}else{
		delete pBuf;
	}
	return nRead;
}

CRDBuf::~CRDBuf(void)
{
	POSITION pos=m_lstBuffer.GetHeadPosition();
	while(pos!=NULL){
		delete[] m_lstBuffer.GetNext(pos);
	}
}
