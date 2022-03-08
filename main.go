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
	var urls []string
	var undupcheckproxylist []string
	var proxylist []string
	dupcheck := make(map[string]bool)
	urls = append(urls, "https://api.openproxylist.xyz/"+mode+".txt")
	urls = append(urls, "https://api.proxyscrape.com/v2/?request=getproxies&protocol="+mode)
	if mode == "http" {
		urls = append(urls, "https://www.proxyscan.io/download?type=http")
		urls = append(urls, "https://www.proxyscan.io/download?type=https")
	} else {
		urls = append(urls, "https://www.proxyscan.io/download?type="+mode)
	}

	for _, url := range urls {
		req, _ := http.Get(url)
		body, _ := ioutil.ReadAll(req.Body)
		// proxylist := strings.Split(string(body), "\n")
		undupcheckproxylist = append(undupcheckproxylist, strings.Split(string(body), "\n")...)
	}
	//delete dup
	for _, proxy := range undupcheckproxylist {
		if !dupcheck[proxy] {
			dupcheck[proxy] = true
			proxylist = append(proxylist, proxy)
		}
	}
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
	// fmt.Println("1.http\n2.socks4\n3.socks5\n")
	fmt.Println("1.http")
	fmt.Println("2.socks4")
	fmt.Println("3.socks5")
	fmt.Println("")
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
