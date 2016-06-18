package sftps

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	LOGFILE = "./sftps.log"
)

type Output struct {
	State   string `json:"state"`
	Command string `json:"command"`
	Result  string `json:"result"`
}

func Err(cmd string, err error, f func()) {
	if err != nil {
		Last(cmd, err, f)
	}
	return
}

func Last(cmd string, param interface{}, f func()) {

	out := new(Output)
	logger := GetLogger()
	if f != nil {
		f()
	}
	if p, ok := param.([]*Entity); ok {
		bytes, err := json.Marshal(p)
		if err != nil {
			fmt.Println("{\"ERROR\":\"Failed, convert the slice of Entity to json\"}")
			os.Exit(0)
		}
		listJson := string(bytes)
		res := fmt.Sprintf("{\"State\":\"DONE\",\"Cmd\":\"GetList\",\"Result\":%s}", listJson)
		fmt.Printf("%s\n", res)
		logger.Printf("[%s] %s\n", "DONE", "GetList")
		os.Exit(0)
	}
	if p, ok := param.(error); ok {
		out.State = "ERROR"
		out.Command = cmd
		out.Result = fmt.Sprintf("%v", p)
		logger.Printf("[%s] %v\n", "ERROR", p)
	}
	if p, ok := param.(string); ok {
		out.State = "DONE"
		out.Command = cmd
		out.Result = fmt.Sprintf("%s", p)
		logger.Printf("[%s] %s\n", "DONE", p)
	}

	bytes, err := json.Marshal(out)
	if err != nil {
		fmt.Println("{\"ERROR\":\"Failed, convert result value to json\"}")
		os.Exit(0)
	}
	fmt.Printf("%s\n", string(bytes))
	os.Exit(0)
}

func GetLogger() (logger *log.Logger) {

	lf, err := os.OpenFile(LOGFILE, os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("{\"ERROR\":\"%v\"}", err)
		os.Exit(0)
	}
	logger = log.New(lf, "[SFTPS", log.Ldate|log.Ltime|log.Lshortfile)

	return
}
