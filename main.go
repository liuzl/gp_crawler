package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/crawlerclub/x/downloader"
	"github.com/crawlerclub/x/parser"
	"github.com/crawlerclub/x/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

var permissions = []string{"read phone status and identity", "directly call phone numbers", "reroute outgoing calls", "read your contacts", "draw over other apps"}

func getPermissions(pkg string, substrs []string) map[string]bool {
	req := &types.HttpRequest{Url: "https://play.google.com/store/xhr/getdoc", Method: "POST", UseProxy: false, Platform: "pc"}
	req.PostData = fmt.Sprintf("ids=%s&xhr=1", pkg)
	res := downloader.Download(req)

	if res.Error != nil {
		log.Println(res.Error)
		return nil
	}
	ret := make(map[string]bool)
	for _, s := range substrs {
		ret[s] = strings.Contains(res.Text, s)
	}
	return ret
}

func crawlInfo(pkg string, parseConf *types.ParseConf) ([]string, []string) {
	pageUrl := fmt.Sprintf("https://play.google.com/store/apps/details?id=%s", pkg)
	req := &types.HttpRequest{Url: pageUrl, Method: "GET", UseProxy: false, Platform: "pc"}
	fmt.Println(pageUrl)
	res := downloader.Download(req)
	if res.Error != nil {
		log.Println(res.Error)
		return nil, nil
	}

	html := strings.Replace(res.Text, "<head>", "<head><meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">", 1)
	_, retItems, err := parser.Parse([]byte(html), pageUrl, parseConf)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	if len(retItems) != 1 {
		return nil, nil
	}
	ret := getPermissions(pkg, permissions)
	item := retItems[0]
	for k, v := range ret {
		item[k] = v
	}
	keys := make([]string, len(item))
	values := make([]string, len(item))
	i := 0
	for k, _ := range item {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for i, k := range keys {
		values[i] = fmt.Sprintf("%v", item[k])
	}
	return keys, values
}

func main() {
	confFile := flag.String("conf", "./google_play.json", "crawler configure file")
	pkgFile := flag.String("pkgs", "./pkgs.txt", "package name file")
	resultFile := flag.String("results", "./results.csv", "results file, csv format")
	isTest := flag.Bool("test", false, "is test mode")
	flag.Parse()

	conf, _ := ioutil.ReadFile(*confFile)
	var parseConf types.ParseConf
	err := json.Unmarshal(conf, &parseConf)
	if err != nil {
		log.Fatal(err)
	}

	pkgs, err := os.Open(*pkgFile)
	if err != nil {
		log.Fatal(err)
	}
	defer pkgs.Close()

	br := bufio.NewReader(pkgs)
	first := true
	f, err := os.OpenFile(*resultFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		i++
		if *isTest && i > 10 {
			break
		}
		keys, values := crawlInfo(string(line), &parseConf)
		if first {
			log.Println(strings.Join(keys, "\t"))
			fmt.Fprintln(f, strings.Join(keys, "\t"))
			first = false
		}
		log.Println(strings.Join(values, "\t"))
		fmt.Fprintln(f, strings.Join(values, "\t"))
	}
}
