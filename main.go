package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var chunkFlag = flag.Int("chunk", 10, "Per chunk")
var commandFlag = flag.String("command", "", "Command to fanout")

// go install github.com/ketan-10/go-fanout@commit
// ~/go/bin/go-fanout --chunk=20 --command="/home/ketan/go/bin/goimports -w" --chunk=40 -- *.go
func main() {

	fmt.Println(os.Args)

	flag.Parse()
	chunkSize := *chunkFlag
	fmt.Printf("chunkSize: %d \n", chunkSize)
	if chunkSize <= 0 {
		panic("chunk must be greater than zero")
	}
	if *commandFlag == "" {
		panic("command flag must be provided")
	}

	idx := contains(os.Args, "--")

	if idx == -1 {
		panic("command arguments not provided after --")
	}

	arguments := os.Args[idx+1:]

	fan := newFanOut(runtime.NumCPU())

	chunks := chunkThis(arguments, chunkSize)
	for idx := range chunks {
		fan.Run(func(i any) {
			chunk := chunks[i.(int)]
			commands := append(strings.Split(*commandFlag, " "), chunk...)
			fmt.Println(commands)

			err := exec.Command(commands[0], commands[1:]...).Run()
			if err != nil {
				fmt.Println(err)
			}
		}, idx)
	}

	fan.Wait()

}

func contains(list []string, value string) int {
	for idx, v := range list {
		if v == value {
			return idx
		}
	}
	return -1
}

func chunkThis(list []string, chunkSize int) [][]string {
	var out [][]string
	for i := 0; i < len(list); i += chunkSize {
		end := i + chunkSize
		if end > len(list) {
			end = len(list)
		}
		out = append(out, list[i:end])
	}
	return out
}

func newFanOut(cpuSize int) *FanOut {
	if cpuSize <= 0 {
		cpuSize = 1
	}
	pool := FanOut{
		cpuLock: make(chan bool, cpuSize),
	}

	return &pool
}

type FanOut struct {
	cpuLock chan bool
}

func (fan *FanOut) Run(callback func(any), data any) {

	// Here we aq
	fan.cpuLock <- true
	go func() {
		callback(data)
		<-fan.cpuLock
	}()
}

func (fan *FanOut) Wait() {
	for i := 0; i < cap(fan.cpuLock); i++ {
		fan.cpuLock <- true
	}
}
