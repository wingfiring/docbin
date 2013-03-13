package main

import (
	"flag"
	"fmt"
	"os"
	"encoding/json"
	"net/http"
	"net/http/fcgi"
)

type Config struct{
	Root string
	Docs map[string]string
}
type FastCGIServer struct{
	configFile string
}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%+v\n", req)
	fmt.Printf("A Request..\n")
}

func main() int {
	b := new(FastCGIServer)
	bconfigFile := flag.String("config", "docbin.conf", "docbin config file path")
	flag.Parse()

	f, err := os.Open(*configFile)
	if err != nil {
		fmt.Print("Failed to open config file %s\n", configFile)
		return 1
	}
	defer f.Close()

	jdecoder := json.NewDecoder(f)

	var cfg Config
	err = jdecoder.Decode(&cfg)
	if err != nil {
		fmt.Printf("Failed to parse config file %s, %+v\n", configFile, err)
		return 1
	}

//	fmt.Printf("%+v\n", cfg)


	fmt.Printf("Starting server")
	fcgi.Serve(nil, b)

	return 0
}
