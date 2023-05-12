package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Printf("CurrentTime in milliseconds: %d \n", time.Now().UnixMilli())
}
