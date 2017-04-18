#include<stdio.h>
#include <assert.h>
#include <time.h>
#include <stdlib.h>

#define SCORE_INIT -20000

const int VAL_WIN=10000;

int frame[15][15];

int max_score;
int max_depth=3;// should be even
int curstep=0;
#define MAX_STEP 15*15

struct STEP{
	int x,y;
	int bw;
//	int n;
};


struct STEP max_route[100];
struct STEP current_path[100];

typedef struct _BWSCORE{
	int bscore;
	int wscore;
}BWCORE;

int computerbw=2;	// computer use white

void addcond(BWCORE *ps, int bw,int cnt,int left,int right, int midspace){
	int score=0;
	if(midspace==cnt)
		midspace=0;
	if( cnt>=5 ){
		if (midspace>=5 || cnt-midspace>=5)
			score=50000;
		else score=720;
	}
	if( cnt==4){
		if(!midspace){
			if( left && right)
				score=4320;
			else if(left || right)
				score=720;
		}else{
				score=720;
		}
		if(bw==1) score*=2;
	}
	if (cnt==3){
		if(left && right && left+right>2)
		{
			score=720;
			if (bw==1) score*=2;
		}
		else if(left+right>=2)
			score=120;
		if(midspace)
			score-=40;
	}

	if (cnt==2){
		if(left && right && left+right>3)
			score=120;
		else if(left +right>=3)
			score=20;
		if(bw==1) score+=10;
	}

	if(cnt==1){
		if(left && right && right+left>=4)
			score=20;
	}
	if(bw==1)
		ps->bscore+=score;
	else if(bw==2)
		ps->wscore+=score;
//    printf("addcount: bw %d, cnt %d\n",bw,cnt, score);
}

BWCORE evaluate(){
	BWCORE s={
		.bscore=0,
		.wscore=0
	};
	// --
	int i,j;
	for(i=0;i<15;++i){
		int incal=0,last=0; // 0,1,2
		int lsp=0,rsp=0,msp=0;
		int cnt=0;
		for(j=0;j<15;++j){
			int cur=frame[i][j];
			if(!incal){
				if(cur==0) // ---
					lsp++;
				else{
					incal=cur;// --*
					cnt=1;
					if(j==14)
						addcond(&s,incal,cnt,lsp,0,0);
				}
			}else{ // incal!=0
				if(cur==0){ // -
					if(last==0){ // --
						int k;
						assert(msp!=0); //  --
						if(msp==cnt)
							msp=0;
						for(k=j;k<15 && !frame[i][j];++k)
							rsp++;
						addcond(&s,incal,cnt,lsp,rsp,msp);
						cnt=rsp=incal=0;
						lsp=2;
					}else{ // last!=0 ***-
						if(j==14){
							addcond(&s,incal,cnt,lsp,1,msp);
							continue;
						}
						if(msp==0){ 
							msp=cnt;
							rsp=1;
						}else{ // msp!=0 && last!=0: *-**-
							if(frame[i][j+1]==0){ // *-**--
								addcond(&s,incal,cnt,lsp,2,msp);
								cnt=incal=rsp=0;
								lsp=1;
							}else{ // *-**-?
								addcond(&s,incal,cnt,lsp,1,msp);
								if(frame[i][j+1]==incal){ // *-**-*
									cnt=cnt-msp;
									lsp=1;
									msp=cnt;
									rsp=1;
								}else{ // *--**-x
									cnt=0;
									lsp=1;
									msp=0;
									rsp=0;
									incal=0;
								}
							}
						}
					}

				}else{	// incal!=0, cur!=0 : *x || * x ||  ** ||  *-*
					if(j==14){
						if(cur!=incal)
							addcond(&s,incal,cnt,lsp,rsp,msp);
						else
							addcond(&s,incal,cnt+1,lsp,0,msp);
						continue;
					}
					if(incal==cur){
						cnt++;
						rsp=0;
					}else{
						addcond(&s,incal,cnt,lsp,rsp,msp);
						incal=cur;
						lsp=rsp;
						rsp=0;
						msp=0;
						cnt=1;
					}
					
				}
			}
			last=cur;
		}
	}

// |
	for(i=0;i<15;++i){
		int incal=0,last=0; // 0,1,2
		int lsp=0,rsp=0,msp=0;
		int cnt=0;
		for(j=0;j<15;++j){
			int cur=frame[j][i];
			if(!incal){
				if(cur==0) // ---
					lsp++;
				else{
					incal=cur;// --*
					cnt=1;
					if(j==14)
						addcond(&s,incal,cnt,lsp,0,0);
				}
			}else{ // incal!=0
				if(cur==0){ // -
					if(last==0){ // --
						int k;
						assert(msp!=0); //  --
						if(msp==cnt)
							msp=0;
						for(k=j;k<15 && !frame[i][j];++k)
							rsp++;
						addcond(&s,incal,cnt,lsp,rsp,msp);
						cnt=rsp=incal=0;
						lsp=2;
					}else{ // last!=0 ***-
						if(j==14){
							addcond(&s,incal,cnt,lsp,1,msp);
							continue;
						}
						if(msp==0){ 
							msp=cnt;
							rsp=1;
						}else{ // msp!=0 && last!=0: *-**-
							if(frame[j+1][i]==0){ // *-**--
								addcond(&s,incal,cnt,lsp,2,msp);
								cnt=incal=rsp=0;
								lsp=1;
							}else{ // *-**-?
								addcond(&s,incal,cnt,lsp,1,msp);
								if(frame[j+1][i]==incal){ // *-**-*
									cnt=cnt-msp;
									lsp=1;
									msp=cnt;
									rsp=1;
								}else{ // *--**-x
									cnt=0;
									lsp=1;
									msp=0;
									rsp=0;
									incal=0;
								}
							}
						}
					}
				}else{	// incal!=0, cur!=0 : *x || * x ||  ** ||  *-*
					if(j==14){
						if(cur!=incal)
							addcond(&s,incal,cnt,lsp,rsp,msp);
						else
							addcond(&s,incal,cnt+1,lsp,0,msp);
						continue;
					}
					if(incal==cur){
						cnt++;
						rsp=0;
					}else{
						addcond(&s,incal,cnt,lsp,rsp,msp);
						incal=cur;
						lsp=rsp;
						rsp=0;
						msp=0;
						cnt=1;
					}
					
				}
			}
			last=cur;
		}
	}

/*
	for(i=0;i<15;++i){
		int incal=0,last=0; // 0,1,2
		int lsp=0,rsp=0,msp=0;
		int cnt=0;
		for(j=0;;++j){
			int cur=frame[j][i];
			if(!incal){
				if(cur==0) // ---
					lsp++;
				else{
					incal=cur;// --*
					cnt=1;
					if(j==14)
						addcond(&s,incal,cnt,lsp,0,0);
				}
			}else{ // incal!=0
				if(cur==0){ // -
					if(last==0){ // --
						int k;
						assert(msp!=0); //  --
						if(msp==cnt)
							msp=0;
						for(k=j;k<15 && !frame[i][j];++k)
							rsp++;
						addcond(&s,incal,cnt,lsp,rsp,msp);
						cnt=rsp=incal=0;
						lsp=2;
					}else{ // last!=0 ***-
						if(j==14){
							addcond(&s,incal,cnt,lsp,1,msp);
							continue;
						}
						if(msp==0){ 
							msp=cnt;
							rsp=1;
						}else{ // msp!=0 && last!=0: *-**-
							if(frame[j+1][i]==0){ // *-**--
								addcond(&s,incal,cnt,lsp,2,msp);
								cnt=incal=rsp=0;
								lsp=1;
							}else{ // *-**-?
								addcond(&s,incal,cnt,lsp,1,msp);
								if(frame[j+1][i]==incal){ // *-**-*
									cnt=cnt-msp;
									lsp=1;
									msp=cnt;
									rsp=1;
								}else{ // *--**-x
									cnt=0;
									lsp=1;
									msp=0;
									rsp=0;
									incal=0;
								}
							}
						}
					}
				}else{	// incal!=0, cur!=0 : *x || * x ||  ** ||  *-*
					if(j==14){
						if(cur!=incal)
							addcond(&s,incal,cnt,lsp,rsp,msp);
						else
							addcond(&s,incal,cnt+1,lsp,0,msp);
						continue;
					}
					if(incal==cur){
						cnt++;
						rsp=0;
					}else{
						addcond(&s,incal,cnt,lsp,rsp,msp);
						incal=cur;
						lsp=rsp;
						rsp=0;
						msp=0;
						cnt=1;
					}
					
				}
			}
			last=cur;
		}
	}
*/
	return s;
}

int getpossible(struct STEP st[MAX_STEP]){
	int i,j;
	int n=0;
	if(curstep==0){
		st[0].x=7;
		st[0].y=7;
		return 1;
	}
	for(i=0;i<15;++i)
		for(j=0;j<15;j++){
			if(frame[i][j]==0){
				int x,y;
				int find;
				for(x=-2,find=0;x<=2;++x){
					if(i+x<0 || i+x>14) continue;
					for(y=-2;y<=2;y++){
						if(j+y<0 || j+y>14 || (x==0 && y==0))
							continue;	
						if(frame[x+i][y+j]){
							find=1;
							st[n].x=i;
							st[n].y=j;
							st[n].bw=(curstep%2)?2:1;
							n++;
							break;
						}
					}
					if(find)
						break;
				}
			}
		}
	return n;
}

int checkover(int x, int y){
	int i,n;
	for(n=0,i=x-1;i>=0 && frame[i][y]==frame[x][y] && n<4; i--){
		n++;
	}
	for(i=x+1;i<15 && frame[i][y]==frame[x][y] && n<4; i++){
		n++;
	}
	if (n>=4)
		return frame[x][y];

	for(i=y-1,n=0;i>=0 && frame[x][i]==frame[x][y] && n<4; i--){
		n++;
	}
	for(i=y+1;i<15 && frame[x][i]==frame[x][y] && n<4; i++){
		n++;
	}
	if(n>=4)
		return frame[x][y];

    for(i=1,n=0;x-i>=0 && y-i>=0 && frame[x-i][y-i]==frame[x][y] && n<4; i++){
        n++;
    }
    for(i=1;x+i<15 && y+i<15 && frame[x+i][y+i]==frame[x][y] && n<4; i++){
	}
	if (n>=4)
		return frame[x][y];

    for(i=1,n=0;x-i>=0 && y+i<15 && frame[x-i][y+i]==frame[x][y] && n<4; i++){
        n++;
    }
    for(i=1;x+i<15 && y-i>=0 && frame[x+i][y-i]==frame[x][y] && n<4; i++){
	}
	if (n>=4)
		return frame[x][y];
	
	return 0;
}

void applystep(struct STEP st){
	assert(frame[st.x][st.y]==0);
	frame[st.x][st.y]=st.bw;
}

void undostep(struct STEP st){
	frame[st.x][st.y]=0;
}

void searchsteps(int depth){
	if (depth>max_depth){
		return;
	}
	struct STEP steps[MAX_STEP];
	int npos=getpossible(steps);
	int i;
	for(i=0;i<npos;++i){
		int over;
		applystep(steps[i]);
		current_path[depth-1]=steps[i];
/*		over=checkover();
		if(over==computerbw){
			int j;
			for(j=0;j<depth;++j)
				max_route[j]=current_path[j];
			max_score=VAL_WIN;
			undostep(steps[i]);
			return;
		}else if(over!=0){
		}*/
		if(depth==max_depth){
			int i;
			BWCORE score=evaluate();
			int compare=computerbw==1?score.bscore-score.wscore:score.wscore-score.bscore;
			if(compare>max_score){
				for(i=0;i<max_depth;i++){
					max_route[i]=current_path[i];
				}
				max_score=compare;
			}
		} else
			searchsteps(depth+1);
		undostep(steps[i]);
	}
}

void algol1(){
	int nret;
	struct STEP ret[MAX_STEP];
	struct STEP steps[MAX_STEP];
	int max=SCORE_INIT;
	int npos=getpossible(steps);
	int i;
	int samecnt=0;
	for(i=0;i<npos;i++){
		BWCORE sc;
		int score;
		applystep(steps[i]);
		sc=evaluate();
printf("%d,%d: b: %d, w %d\n",steps[i].x,steps[i].y,sc.bscore,sc.wscore);
		score=sc.wscore-sc.bscore;
		if(score>max){
			max=score;
			ret[0]=steps[i];
			samecnt=1;
		}else if(score==max){
			ret[samecnt++]=steps[i];
		}
		undostep(steps[i]);
	}
	npos=0;
	if(samecnt>1)
		npos=rand()%samecnt;
    applystep(ret[npos]);
	printf("%d,%d, evaluate %d\n",ret[npos].x,ret[npos].y,max);
}

void compute()
{
	max_score=SCORE_INIT;
//	searchsteps(1);
	algol1();
//	frame[max_route[0].x][max_route[0].y]=computerbw;
//	printf("step: %d,%d, evaluate: %d\n", max_route[0].x,max_route[0].y,max_score);
}

int main()
{
	int you=computerbw==1?2:1;
	srand(time(NULL));
	while(1){
		int x,y;
		printf("your turn: x y --");
		scanf("%d%d",&x,&y);
		frame[x][y]=you;
		curstep++;
		if(checkover(x,y)){
			printf("You win!\n");
			break;
		}
		compute();
		curstep++;
		if(checkover(max_route[0].x,max_route[0].y)){
			printf("computer win!\n");
			break;
		}
	}
}
