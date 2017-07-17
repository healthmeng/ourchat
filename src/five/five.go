package main

/*
requirement:
1. common AI
2. concurrent computing
3. remote service
*/


import (
"fmt"
)

const(
	max_step=15*15
)

type StepInfo struct{
	x,y int
	bw int
}

type AIPlayer struct{
	frame [15][15] int
	level int
	steps []StepInfo
	curstep int
}

func InitServer(color int, level int) (* AIPlayer,error){
	player:=new (AIPlayer)
	player.level=level
	player.steps=make([]StepInfo,max_step,max_step)
	return player, nil
}

func (player* AIPlayer)SetStep(x int,y int){
}

func (player* AIPlayer)GetStep()(x int,y int){
	return
}

func main(){

	fmt.Println("Start:")
	if err:=InitServer;err!=nil{
		fmt.Println("Init server error:",err)
	}
}
