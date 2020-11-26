package models

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/signintech/gopdf"
)

//SaveImagePDF 导出图形到PDF
func SaveImagePDF(hostids []string, start, end string) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}})
	pdf.AddPage()
	err = pdf.AddTTFFont("msty", "./msty.ttf")
	if err != nil {
		beego.Error(err.Error())
	}

	err = pdf.SetFont("msty", "", 8)
	if err != nil {
		beego.Error(err.Error())
	}
	//表头配置
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
			beego.Error(err.Error())
		}
		hba, err := json.Marshal(rep.Result)
		if err != nil {
			beego.Error(err.Error())
		}
		var hb []GraphInfo
		err = json.Unmarshal(hba, &hb)
		if err != nil {
			beego.Error(err.Error())
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
			if (index%2 == 1) && (index+1 < all) {
				//表头配置
				pdf.AddPage()
				pdf.CellWithOption(nil, "导出时间:"+TimeStr, gopdf.CellOption{Align: 0, Border: 0})
				pdf.SetX(10)                            // x coordinate specification
				pdf.SetY(20)                            // y coordinate specification
				pdf.Cell(nil, "导出周期:"+start+"----"+end) // Rect, String
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
	}
	var b bytes.Buffer
	err = pdf.Write(&b)
	if err != nil {
		beego.Error(err.Error())
	}
	return b.Bytes(), nil
}

//GetPdfImageHolder func
func GetPdfImageHolder(grupinfo GraphInfo, start, end string, wg *sync.WaitGroup, pdfHolder chan<- gopdf.ImageHolder) {
	defer wg.Done()
	//请求图形
	ZabbixWeb := beego.AppConfig.String("zabbix_web")
	client1 := &http.Client{nil, nil,
		JAR, 99999999999992}
	imgurl := ZabbixWeb + "/chart2.php?"
	data := url.Values{}
	URL, err := url.Parse(imgurl)
	if err != nil {
		beego.Error("Fatal error ", err.Error())
	}
	data.Set("graphid", grupinfo.GraphID)
	data.Set("from", start)
	data.Set("to", end)
	data.Set("profileIdx", "web.graphs.filter")
	data.Set("profileIdx2", "200")
	data.Set("width", "800")
	//Encode rul
	URL.RawQuery = data.Encode()
	urlPath := URL.String()
	reqest1, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		beego.Error("Fatal error ", err.Error())
	}
	reqest1.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest1.Header.Add("Accept-Encoding", "gzip, deflate")
	reqest1.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	reqest1.Header.Add("Connection", "keep-alive")
	reqest1.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
	response1, err := client1.Do(reqest1)
	if err != nil {
		beego.Error("Fatal error ", err.Error())
	}
	defer response1.Body.Close()
	if response1.StatusCode == 200 {
		var reader io.Reader
		switch response1.Header.Get("Content-Encoding") {
		case "gzip":
			reader, _ = gzip.NewReader(response1.Body)
		default:
			reader = response1.Body
		}
		imgH2, err := gopdf.ImageHolderByReader(reader)
		if err != nil {
			beego.Error("Fatal error ", err.Error())
		}
		pdfHolder <- imgH2
	}
}
