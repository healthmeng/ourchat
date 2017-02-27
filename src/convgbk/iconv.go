package convgbk

import (
"errors"
"unsafe"
)

/*
#include <stdio.h>
#include <string.h>
#include <iconv.h>

enum CONV{
 utf2gb, gb2utf
};

int gbutf8(int c, char* src, char* dest, int maxsize){
 size_t lsrc=strlen(src);
 size_t lout=maxsize-1;
 int ret=0;
 int tmp=0;
 iconv_t td;
 if(src ==NULL || dest==NULL)
	return -1;
 if(lsrc==0)
 {
	*dest='\0';
	return 0;
 }
 if(c==utf2gb)
	td=iconv_open("gbk","utf8");
 else if(c==gb2utf)
	td=iconv_open("utf8","gbk");
 else return -1;
 
 if (td<=0){
	printf("iconv open error\n");
	return -1;
 }
 while(lsrc>0){
	tmp=iconv(td,&src,&lsrc,&dest,&lout);
	if (tmp<0) {
	    ret=-1;
            perror("conv error:");
	    break;
	}
 }
 if(!ret && *dest!='\0')
	*++dest='\0';
 iconv_close(td);
 return ret;
}

*/
import "C"

func UTF2GB(src string) (string,error){
	slen:=len(src)+1
	dst:=make([]byte,slen,slen)
	cptr:=(*C.char)((unsafe.Pointer)(&dst[0]))
	if ret:=C.gbutf8(C.int(0),C.CString(src),cptr,C.int(slen));ret!=0{
		return "",errors.New("Convert error")
	}
	return C.GoString(cptr),nil
}

func GB2UTF(src string)(string,error){
	slen:=len(src)+1
	dst:=make([]byte,slen,slen)
	cptr:=(*C.char)((unsafe.Pointer)(&dst[0]))
	if ret:=C.gbutf8(C.int(1),C.CString(src),cptr,C.int(slen));ret!=0{
		return "",errors.New("Convert error")
	}
	return C.GoString(cptr),nil
}

