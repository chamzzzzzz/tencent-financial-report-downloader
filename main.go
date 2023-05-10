package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chamzzzzzz/supersimplesoup"
)

func main() {
	log.Printf("start download at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.tencent.com/zh-cn/investors/financial-reports.html", nil)
	if err != nil {
		log.Printf("new request failed, err:%v\n", err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.61 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("get http reponse failed, err:%v\n", err)
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read body failed, err:%v\n", err)
		return
	}

	dom, err := supersimplesoup.Parse(bytes.NewReader(b))
	if err != nil {
		log.Printf("parse html failed, err:%v\n", err)
		return
	}

	items := dom.QueryAll("a", "class", "ten_report_item")
	if len(items) == 0 {
		log.Printf("find no item\n")
		return
	}

	os.MkdirAll("reports/tencent", 0755)
	for _, item := range items {
		span, err := item.Find("span")
		if err != nil {
			log.Printf("find span failed, err:%v\n", err)
			return
		}
		name := strings.ReplaceAll(span.Text(), " ", "")
		filename := fmt.Sprintf("reports/tencent/%s.pdf", name)
		if _, err := os.Stat(filename); err == nil {
			log.Printf("[%s] already downloaded, skip.\n", name)
			continue
		}
		err = download(item.Href(), filename)
		if err != nil {
			log.Printf("[%s] download failed, err:%v\n", name, err)
			continue
		}
		log.Printf("[%s] downloaded.\n", name)
	}

	log.Printf("finish download at %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

func download(URL, file string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code:%d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(b) != int(resp.ContentLength) {
		return fmt.Errorf("content length not match %d, %d", resp.ContentLength, len(b))
	}
	return os.WriteFile(file, b, 0755)
}
