package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

func ProxyScrape(mode string) []string {
	url := "https://api.openproxylist.xyz/" + mode + ".txt"
	req, _ := http.Get(url)
	body, _ := ioutil.ReadAll(req.Body)
	proxylist := strings.Split(string(body), "\n")
	return proxylist
}

func CheckProxy(proxy string, mode string) bool {
	proxyurl, err := url.Parse(mode + "://" + proxy)
	if err != nil {
		return false
	}
	client := &http.Client{}
	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyurl),
	}
	client.Timeout = 3 * time.Second
	req, err := client.Get("http://google.com/")
	if err != nil {
		return false
	}
	req.Body.Close()
	return true
}

func main() {
	fmt.Println("1.http\n2.socks4\n3.socks5\n")
	fmt.Print("ScrapeMode: ")
	var ScrapeMode string
	var Threads int
	fmt.Scan(&ScrapeMode)
	switch ScrapeMode {
	case "1":
		ScrapeMode = "http"
	case "2":
		ScrapeMode = "socks4"
	case "3":
		ScrapeMode = "socks5"
	default:
		fmt.Println("error")
		return
	}
	fmt.Print("Threads: ")
	fmt.Scan(&Threads)
	proxys := ProxyScrape(ScrapeMode)
	file, _ := os.OpenFile(ScrapeMode+".txt", os.O_CREATE|os.O_SYNC, 0664)
	ch := make(chan bool, Threads)
	wg := &sync.WaitGroup{}
	for _, proxy := range proxys {
		ch <- true
		wg.Add(1)
		go func(proxy string) {
			if CheckProxy(proxy, ScrapeMode) {
				fmt.Sprintln(fmt.Fprintln(color.Output, color.HiGreenString("[+]"+proxy))) //fast???
				// fmt.Fprintln(color.Output, color.HiGreenString("[+]"+proxy))
				file.Write([]byte(proxy + "\n"))

			} else {
				fmt.Sprintln(fmt.Fprintln(color.Output, color.HiRedString("[-]"+proxy))) //fast??
				// fmt.Fprintln(color.Output, color.HiRedString("[-]"+proxy))
			}
			<-ch
			wg.Done()
		}(proxy)
	}
	wg.Wait()
	file.Close()
}
