package main

import (
	"fmt"
	"math"
	"sync"
)

type Direction struct {
	xDir int   //-1,0,1
	yDir int   //-1,0,1
}

func getPath(m *[][]tile, from *tile) ([]*tile, bool) {

	// map to keep track of the final path
	var parentOf map[*tile]*tile
	parentOf = make(map[*tile]*tile)

	//initialise 'costqueue', start-0, other-infinite
	costQueue := queue{}

	for i, list := range *m {
		for j, _ := range list {
			costQueue.Add(&(*m)[i][j], float32(math.Inf(1)))	
		}
	}

	costQueue.Update(from, 0)

//	checkedQueue := queue{}  // TODO: implement this later for a more efficient algorithm: dummer! går inte fortare..

	v := float32(0)
	current := tileCost{&tile{}, &v}
	currentDir := Direction{1,1}
	
	for len(costQueue) != 0 && !current.tile.door {
	//	fmt.Println("----")
		current = (&costQueue).Pop()
		currentDir = getDir(parentOf[current.tile], current.tile)  // for reference!
	//	fmt.Println(currentDir)
		//neighbors := getNeighbors(current.tile, costQueue)
		neighbors := getNeighborsPruned(current.tile, currentDir)
		var wg sync.WaitGroup
		wg.Add(len(neighbors))
		var mutex = &sync.Mutex{}
		for _, neighbor := range neighbors {
		//	fmt.Println(neighbor)
			go func(n *tile) {			
				defer wg.Done()			
				cost := *current.cost + stepCost(*n)
				if Diagonal(current.tile, n) {cost += float32(math.Sqrt(2)) - 1}
				if n.occupied.IsWaiting() {cost += 1}

				// TODO: 1 default cost improve!? depending on heat, smoke etc
				mutex.Lock()
				if cost < costQueue.costOf(n) {
					
					parentOf[n] = current.tile
					costQueue.Update(n, cost)				
				}
				mutex.Unlock()
			}(neighbor)		
		}
		wg.Wait()	
		//	checkedQueue.AddTC(current)	
	}	
	return compactPath(parentOf, from, current.tile)
}

func contains(tiles []*tile, t *tile) bool {
	for _, ti := range tiles {
		if ti == t {
			return true
		}
	}
	return false
}

func stepCost(t tile) float32 {
	cost := float32(1)
	cost += float32(t.heat) / 5 //TODO how much cost for fire etc??
	if t.fireLevel > 0 {
		cost = float32(math.Inf(1))
	}
	return cost
}


/*
func getJumpPoint(m *[][]tile, current *tile, dir Direction, from *tile, to *tile) *tile {
	//from+to onödig(?)
	nextX := current.xCoord + dir.xDir
	nextY := current.yCoord + dir.yDir
	nextTile := GetTile(m, nextX, nextY)

	if nextTile == nil {return nil}
	
	if nextTile.door {return nextTile}

	//	if 
}*/

func getNeighborsPruned(current *tile, dir Direction) []*tile{
	neighbors := []*tile{}

	north := validTile(current.neighborNorth) 
	east := validTile(current.neighborEast)
	west := validTile(current.neighborWest)
	south := validTile(current.neighborSouth)   // replace !?
	
	if dir.yDir == 0 {  // horisontal/vertical? hur vare med coordsen..
		if dir.xDir == -1 {  // går rakt uppåt
			if north {
				neighbors = append(neighbors, current.neighborNorth)
				//	if !west {neighbors = append(neighbors, current.neighborNW )}
				//	if !east {neighbors = append(neighbors, current.neighborNE )}
			
			}
			if !validTile(current.neighborSW) && west {
				neighbors = append(neighbors, current.neighborWest)
				if north && validTile(current.neighborNW) {neighbors = append(neighbors, current.neighborNW)}
			}
			if !validTile(current.neighborSE) && east {
				neighbors = append(neighbors, current.neighborEast)
				if north && validTile(current.neighborNE) {neighbors = append(neighbors, current.neighborNE)}
			}
			
		} else { // går rakt neråt
			if validTile(current.neighborSouth) {
				neighbors = append(neighbors, current.neighborSouth)
				//	if !validTile(current.neighborWest) {neighbors = append(neighbors, current.neighborSW )}
				//	if !validTile(current.neighborEast) {neighbors = append(neighbors, current.neighborSE )}				
			}
			if !validTile(current.neighborNW) && west {
				neighbors = append(neighbors, current.neighborWest)
				if south && validTile(current.neighborSW) {neighbors = append(neighbors, current.neighborSW)}	
			}
			if !validTile(current.neighborNE) && east {
				neighbors = append(neighbors, current.neighborEast)
				if south && validTile(current.neighborSE) {neighbors = append(neighbors, current.neighborSE)}
			}  // done so far!			
		}		
	} else if dir.yDir == 1 { 
		if dir.xDir == 1 { // går SE
			if validTile(current.neighborEast) {
				neighbors = append(neighbors, current.neighborEast)			
			}
			if validTile(current.neighborSouth) {
				neighbors = append(neighbors, current.neighborSouth)
			//	if !validTile(current.neighborNW) && validTile(current.neighborNW) {
			//		neighbors = append(neighbors, current.neighborNW)}
			}
			if east && south && validTile(current.neighborSE) {neighbors = append(neighbors, current.neighborSE)}

			
		} else if dir.xDir == -1 { // går NE    
			if east {neighbors = append(neighbors, current.neighborEast)}
			if east && north && validTile(current.neighborSE) {neighbors = append(neighbors, current.neighborNE)}
			if north {
				neighbors = append(neighbors, current.neighborNorth)
				//	if !west && validTile(current.neighborSW) {neighbors = append(neighbors, current.neighborSW)}
			}
			
		} else {  // ydir = 0,  går höger   //fixed!
			if east {
				neighbors = append(neighbors, current.neighborEast)
			}
			if !validTile(current.neighborNW) && north {
				neighbors = append(neighbors, current.neighborNorth)
				if east && validTile(current.neighborNE) {neighbors = append(neighbors, current.neighborNE)}
			}
			if !validTile(current.neighborSW) && south {
				neighbors = append(neighbors, current.neighborSouth)
				if east && validTile(current.neighborSE) {neighbors = append(neighbors, current.neighborSE)}
			}
		}

	} else { //xdir = -1
		if dir.xDir == 1 { // går NW
			if west {neighbors = append(neighbors, current.neighborWest)}
			if north {neighbors = append(neighbors, current.neighborNorth)}
			if west && north && validTile(current.neighborNW) {neighbors = append(neighbors, current.neighborNW)}	
		} else if dir.xDir == -1 { // går SW
			if west {neighbors = append(neighbors, current.neighborWest)}
			if south {neighbors = append(neighbors, current.neighborSouth)}
			if west && south && validTile(current.neighborSW) {neighbors = append(neighbors, current.neighborSW)}
		} else {  // ydir = 0,  går vänster
			if west {neighbors = append(neighbors, current.neighborWest)}
			if !validTile(current.neighborNE) && north {
				neighbors = append(neighbors, current.neighborNorth)
				if west && validTile(current.neighborNW) {neighbors = append(neighbors, current.neighborNW)}
			}
			if !validTile(current.neighborSE) && south {
				neighbors = append(neighbors, current.neighborSouth)
				if west && validTile(current.neighborSW) {neighbors = append(neighbors, current.neighborSW)}
			}
			
		}
	}

	return neighbors
}


func getNeighbors(current *tile, costQueue queue) []*tile {
	neighbors := []*tile{}

	north := validTile(current.neighborNorth) 
	east := validTile(current.neighborEast)
	west := validTile(current.neighborWest)
	south := validTile(current.neighborSouth)

	if north {
		neighbors = append(neighbors, current.neighborNorth)
		if west && validTile(current.neighborNW) {
			neighbors = append(neighbors, current.neighborNW)}
		if east && validTile(current.neighborNE) {
			neighbors = append(neighbors, current.neighborNE)}
	}
	if east {neighbors = append(neighbors, current.neighborEast)}
	if west {neighbors = append(neighbors, current.neighborWest)}
	if south {
		neighbors = append(neighbors, current.neighborSouth)
		if west && validTile(current.neighborSW) {
			neighbors = append(neighbors, current.neighborSW)}
		if east && validTile(current.neighborSE) {
			neighbors = append(neighbors, current.neighborSE)}	
	}

	//  --this--
	// nedanför kollar om värdet finns i costQueue också.. tar längre tid för 100*100 och 100*200 iaf
	
/*	north := validTile(current.neighborNorth) && costQueue.inQueue(current.neighborNorth)
	east := validTile(current.neighborEast) && costQueue.inQueue(current.neighborEast)
	west := validTile(current.neighborWest) && costQueue.inQueue(current.neighborWest)
	south := validTile(current.neighborSouth) && costQueue.inQueue(current.neighborSouth)

	if north {
		neighbors = append(neighbors, current.neighborNorth)
		if west && validTile(current.neighborNW) && costQueue.inQueue(current.neighborNW) {
			neighbors = append(neighbors, current.neighborNW)}
		if east && validTile(current.neighborNE) && costQueue.inQueue(current.neighborNE){
			neighbors = append(neighbors, current.neighborNE)}
	}
	if east {neighbors = append(neighbors, current.neighborEast)}
	if west {neighbors = append(neighbors, current.neighborWest)}
	if south {
		neighbors = append(neighbors, current.neighborSouth)
		if west && validTile(current.neighborSW) && costQueue.inQueue(current.neighborSW){
			neighbors = append(neighbors, current.neighborSW)}
		if east && validTile(current.neighborSE) && costQueue.inQueue(current.neighborSE){
			neighbors = append(neighbors, current.neighborSE)}	
	} */
//  --this--
	
	/*
	
	if validTile(current.neighborNorth) && costQueue.inQueue(current.neighborNorth){
		neighbors = append(neighbors, current.neighborNorth)
	}
	if validTile(current.neighborEast) && costQueue.inQueue(current.neighborEast){
		neighbors = append(neighbors, current.neighborEast)
	}
	if validTile(current.neighborWest) && costQueue.inQueue(current.neighborWest){
		neighbors = append(neighbors, current.neighborWest)
	}
	if validTile(current.neighborSouth) && costQueue.inQueue(current.neighborSouth){
		neighbors = append(neighbors, current.neighborSouth)
	}
	//
	if validTile(current.neighborNW) && costQueue.inQueue(current.neighborNW){
		neighbors = append(neighbors, current.neighborNW)
	}
	if validTile(current.neighborNE) && costQueue.inQueue(current.neighborNE){
		neighbors = append(neighbors, current.neighborNE)
	}
	if validTile(current.neighborSE) && costQueue.inQueue(current.neighborSE){
		neighbors = append(neighbors, current.neighborSE)
	}
	if validTile(current.neighborSW) && costQueue.inQueue(current.neighborSW){
		neighbors = append(neighbors, current.neighborSW)
	}
	*/
	//

	return neighbors
}



func validTile(t *tile) bool {
	if t == nil {
		return false
	}
	return !t.wall && !t.outOfBounds
}

func compactPath(parentOf map[*tile]*tile, from *tile, to *tile) ([]*tile, bool) {
	path := []*tile{to}

	current := to

	for current.xCoord != from.xCoord || current.yCoord != from.yCoord {
		path = append([]*tile{parentOf[current]}, path...)
		
		ok := true
		current, ok = parentOf[current]
	
		if !ok {
			return nil, false
		}
	}
	return path, true
}

func printPath(path []*tile) {
	if path == nil {
		fmt.Println("No valid path exists")
	}
	for i, t := range path {
		if (t == nil) {
			fmt.Println("End")
		} else {fmt.Println(i, ":", t.xCoord, ",", t.yCoord)}
	}
}

func mainPath() {

	workingPath()
	fmt.Println("--------------")
/*	blockedPath()
	fmt.Println("--------------")
	firePath()*/
	fmt.Println("--------------")
	doorsPath()
}

func workingPath() {
	matrix := [][]int{
		{0, 1, 2, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 0},
		{0, 0, 1, 0}}
	testmap := TileConvert(matrix)

	path, _ := getPath(&testmap, &testmap[0][0])

	fmt.Println("\nWorking path:")
	printPath(path)
}

func blockedPath() {
	matrix := [][]int{
		{0, 1, 2, 0},
		{0, 0, 1, 0},
		{0, 0, 1, 0},
		{0, 0, 1, 0}}
	testmap := TileConvert(matrix)

	path, _ := getPath(&testmap, &testmap[0][0])

	fmt.Println("\nBlocked path:")
	printPath(path)

}

func firePath() {
	matrix := [][]int{
		{0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 2, 0, 0, 0}}
	testmap := TileConvert(matrix)
	SetFire(&(testmap[3][2]))
	for i := 0; i < 10; i++ {
		FireSpread(testmap)
	}

	path, _ := getPath(&testmap, &testmap[0][3])
	fmt.Println("\nFire path:")
	printPath(path)
}

func doorsPath() {
	matrix := [][]int{
		{0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0},
		{2, 0, 0, 1, 0, 0, 0}}

	testmap := TileConvert(matrix)

	path, _ := getPath(&testmap, &testmap[0][0])
	fmt.Println("\nDoors path:")
	printPath(path)
}



// new funcs

func getDir(from *tile, to *tile) Direction{
	if from == nil{ 
//		fmt.Println("nil!")
		//fmt.Println(to.xCoord, to.yCoord)
		//fmt.Println(from.xCoord, from.yCoord)
		return Direction{1,1}}
//	fmt.Println(from.xCoord, from.yCoord)
//	fmt.Println(to.xCoord, to.yCoord)
//	fmt.Println("not nil!")
	return Direction {to.xCoord - from.xCoord, to.yCoord - from.yCoord}
}

func getJumpPoint(current *tile, dir Direction) *tile{
	if current.door {return current}
	if dir.xDir == 0 {
		if dir.yDir == 1 { // höger
			if validTile(current.neighborEast) {
				return getJumpPoint(current.neighborEast, dir)
			} else {return current}
		}
	}
	return nil
}
