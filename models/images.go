package models

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/signintech/gopdf"
)

//LookImg a
func LookImg(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	var name string
	if len(path) > 1 {
		name = path[len(path)-1]
	} else {
		name = "null.png"
	}
	fmt.Println(name)

	graphid := name
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache, value")

	client1 := &http.Client{nil, nil, JAR, 99999999999992}
	reqest1, err := http.NewRequest("GET", "https://zabbix.cactifans.com/chart2.php?graphid="+graphid+"&screenid=16&width=400&height=156&legend=1&profileIdx=web.screens.filter&profileIdx2=16&from=now-1h&to=now", nil)

	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(0)
	}
	response1, err := client1.Do(reqest1)

	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(0)
	}
	defer response1.Body.Close()
	// cookies1 := response1.Cookies()
	// for _, cookie := range cookies1 {
	// 	fmt.Println("cookie:", cookie.Name)
	// }

	if response1.StatusCode == 200 {
		var reader io.Reader

		switch response1.Header.Get("Content-Encoding") {
		case "gzip":
			//clog.Trace("Encoding: %+v", resp.Header.Get("Content-Encoding"))
			reader, _ = gzip.NewReader(response1.Body)
		default:
			reader = response1.Body
		}

		data, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Fatalln("读取响应数据失败: %+v", err)
		}

		w.Write(data)
	}
}

var ZabbixServer = beego.AppConfig.String("zabbix_server")

//SaveImagePDF 导出图形到PDF
func SaveImagePDF(hostids []string, start, end string) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}})
	pdf.AddPage()
	err = pdf.AddTTFFont("msty", "./msty.ttf")
	if err != nil {
		log.Print(err.Error())
		//	return
	}

	err = pdf.SetFont("msty", "", 8)
	if err != nil {
		log.Print(err.Error())
		//return
	}
	TimeStr := time.Now().Format("2006-01-02 15:04:05")
	pdf.CellWithOption(nil, "导出时间:"+TimeStr, gopdf.CellOption{Align: 0, Border: 0})
	pdf.SetX(10)                            // x coordinate specification
	pdf.SetY(20)                            // y coordinate specification
	pdf.Cell(nil, "导出周期:"+start+"----"+end) // Rect, String
	pdf.Line(10, 30, 585, 30)
	pdf.SetLineWidth(1)
	pdf.SetLineType("dashed")

	for _, vv := range hostids {
		rep, err := API.CallWithError("graph.get", Params{"output": "extend",
			"hostids": vv, "sortfiled": "name"})
		if err != nil {
			log.Println(err)
		}
		hba, err := json.Marshal(rep.Result)
		if err != nil {
			log.Println(err)
		}
		var hb []GraphInfo
		err = json.Unmarshal(hba, &hb)
		if err != nil {
			log.Println(err)
		}
		//轮训图形

		PdfResponses := make(chan gopdf.ImageHolder)
		var wg sync.WaitGroup
		all := 0
		for _, tt := range hb {
			wg.Add(1)
			all++
			go GetPdfImageHolder(tt, start, end, &wg, PdfResponses)
		}

		index := 0
		for response := range PdfResponses {
			pdf.ImageByHolder(response, 40, float64(index%2*400+60), nil)
			if index%2 == 1 {
				pdf.AddPage()
				pdf.CellWithOption(nil, "导出时间:"+TimeStr, gopdf.CellOption{Align: 0, Border: 0})
				pdf.SetX(10)                                // x coordinate specification
				pdf.SetY(20)                                // y coordinate specification
				pdf.Cell(nil, "导出周期:"+TimeStr+"——"+TimeStr) // Rect, String
				pdf.Line(10, 30, 585, 30)
				pdf.SetLineWidth(1)
				pdf.SetLineType("dashed")
			}
			index++
			if all == index {
				close(PdfResponses)
			}
		}
		wg.Wait()
		//pdf.AddPage()
	}
	var b bytes.Buffer
	err = pdf.Write(&b)
	if err != nil {
		log.Println(err)
	}
	return b.Bytes(), nil
}

//GetPdfImageHolder func
func GetPdfImageHolder(grupinfo GraphInfo, start, end string, wg *sync.WaitGroup, pdfHolder chan<- gopdf.ImageHolder) {
	defer wg.Done()
	//请求
	client1 := &http.Client{nil, nil, JAR, 99999999999992}
	reqest1, err := http.NewRequest("GET", ZabbixServer+"/chart2.php?graphid="+grupinfo.GraphID+"&from="+start+"&to="+end+
		"&profileIdx=web.graphs.filter&profileIdx2=200&width=800", nil)

	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(0)
	}
	response1, err := client1.Do(reqest1)
	if err != nil {
		beego.Error("Fatal error ", err.Error())
	}
	defer response1.Body.Close()
	if response1.StatusCode == 200 {
		var reader io.Reader
		switch response1.Header.Get("Content-Encoding") {
		case "gzip":
			//clog.Trace("Encoding: %+v", resp.Header.Get("Content-Encoding"))
			reader, _ = gzip.NewReader(response1.Body)
		default:
			reader = response1.Body
		}

		imgH2, err := gopdf.ImageHolderByReader(reader)
		if err != nil {
			log.Print("pdf3", err.Error())
			return
		}
		pdfHolder <- imgH2
	}
}
