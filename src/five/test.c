#include <stdio.h>
#include "robot.h"

static int frame[15][15];
int checkover(int x, int y){
if(x==6 && y==9)
	printf("checkover %d,%d\n",x,y);
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
		n++;
	}
	if (n>=4)
		return frame[x][y];

    for(i=1,n=0;x-i>=0 && y+i<15 && frame[x-i][y+i]==frame[x][y] && n<4; i++){
        n++;
    }
    for(i=1;x+i<15 && y-i>=0 && frame[x+i][y-i]==frame[x][y] && n<4; i++){
		n++;
	}
	if (n>=4)
		return frame[x][y];
	
	return 0;
}

void drawtxt(){
	int i,j;
	for(i=0;i<15;++i){
		if(i==0){
			printf("  ");
			for(j=0;j<15;++j)
				printf("%2d",j);
			printf("\n");
		}
		for(j=0;j<15;++j){
			if(j==0)
				printf("%-2d",i);
			if(frame[j][i]==0)
				printf(" .");
			else if(frame[j][i]==1)
				printf(" X");
			else
				printf(" O");
		}
		printf("\n");
	}
}


int main()
{
	int you=2;
	int com=1;
	
	struct STEP st;
	init_robot(1,0);
	while(1){
		int x,y;
		
		st=get_step();
		frame[st.x][st.y]=st.bw;
		drawtxt();
		if(checkover(st.x,st.y)){
			printf("computer win!\n");
			break;
		}
redo:
		printf("your turn: x y --");
		scanf("%d%d",&x,&y);
		if(frame[x][y])
			goto redo;
		frame[x][y]=you;
		st.x=x;
		st.y=y;
		st.bw=you;
		set_step(st);
		drawtxt();
		if(checkover(x,y)){
			printf("You win!\n");
			break;
		}

	}
}
