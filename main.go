package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/japanese"
	"bufio"
	"strconv"
	// "regexp"
	"strings"
	// "encoding/csv"
	"log"
	// "os"
)

func amazon(as string) {
	//スクレイピング対象URLを設定
	sc_url := "https://www.amazon.co.jp/gp/offer-listing/" + as + "/ref=dp_olp_used?ie=UTF8&condition=used"
 
	//goquery、ページを取得
	res, err := http.Get(sc_url)
	if err != nil {
    	// handle error
	}
	defer res.Body.Close()
 
	utfBody := transform.NewReader(bufio.NewReader(res.Body), japanese.ShiftJIS.NewDecoder())
 
	doc, err := goquery.NewDocumentFromReader(utfBody)
 
	hontai := doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > span.a-size-large.a-color-price.olpOfferPrice.a-text-bold").Text()
	souryo := doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > p > span > span.olpShippingPrice").Text()
	fmt.Println("Amazon：￥" + hontai[23:])
	if len(souryo) < 6 {
		fmt.Printf("送料無料")
	} else {
		fmt.Println("送料：￥" + souryo[7:])
	}
}

func sofmap(jan string) {
		//スクレイピング対象URLを設定
		sc_url := "https://a.sofmap.com/search_result.aspx?mode=SEARCH&gid=&gid=&gid=&keyword_and=&keyword_or=&keyword_not=&product_maker=&product_name=&product_code=&jan_code=" + jan + "&product_type=NEW&product_type=USED&price_from=&price_to=&sale_date_from_year=&sale_date_from_month=&sale_date_from_day=&sale_date_to_year=&sale_date_to_month=&sale_date_to_day=&reserve_date_from_year=&reserve_date_from_month=&reserve_date_from_day=&reserve_date_to_year=&reserve_date_to_month=&reserve_date_to_day=&order_by=DEFAULT&styp=p_kwsk"
	 
		//goquery、ページを取得
		res, err := http.Get(sc_url)
		if err != nil {
			// handle error
		}
		defer res.Body.Close()
	 
		utfBody := transform.NewReader(bufio.NewReader(res.Body), japanese.ShiftJIS.NewDecoder())
	 
		doc, err := goquery.NewDocumentFromReader(utfBody)
		if err != nil{
		  panic(err)
		}
	 
		// 掲載イベントURL一覧を取得
		// doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn").Each(func(i int, s *goquery.Selection) {
		// 	// ブログのタイトルとタグを取得
		// 	title := s.Find("span").Text()
		
		// 	fmt.Printf(title)
		//   })
		a := doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > span.a-size-large.a-color-price.olpOfferPrice.a-text-bold").Text()
		b := doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > p > span > span.olpShippingPrice").Text()
		fmt.Println(a[23:])
		c := strings.Index(b, "5")
		if c == -1 {
			fmt.Printf("")
		} else {
			fmt.Println(b[7:])
		}
}

func failOnError(err error) {
		if err != nil {
			log.Fatal("Error:", err)
		}
}

func surugaya(jan string) {
		//スクレイピング対象URLを設定
		url := "https://www.suruga-ya.jp/search?category=&search_word=&bottom_detail_search_bookmark=1&gtin=" + jan + "&id_s=&jan10=&mpn="
	 
		//goquery、ページを取得
		res, err := http.Get(url)
		if err != nil {
			panic(err)		
		}
		defer res.Body.Close()
	 
		utfBody := transform.NewReader(bufio.NewReader(res.Body), japanese.ShiftJIS.NewDecoder())
	 
		doc, err := goquery.NewDocumentFromReader(utfBody)
		if err != nil{
			panic(err)
		}

	 

		var urls [6]string
		for i := 1; i < 4; i++ {
			var s string
			s = strconv.Itoa(i)
			kakaku := doc.Find("#search_result > div > div:nth-child(" + s + ") > div.item_price > p:nth-child(1) > span > strong")
			if len(kakaku.Text()) < 5 {
				urls[i-1] = ""
			} else {
				urls[i-1] = kakaku.Text()
				fmt.Println("駿河屋：￥" + urls[i-1][6:])
				           
			}
			
		}
		for i := 1; i < 4; i++ {
			var v string
			v = strconv.Itoa(i)
			b := doc.Find("#search_result > div:nth-child(2) > div:nth-child(" + v + ") > div.item_price > p:nth-child(1) > span > strong")
			if len(b.Text()) < 5 {
				urls[i-1] = ""
			} else {
				urls[i+2] = b.Text()
				fmt.Println(urls[i+2][6:])
			}
		}
}


func main() {
	var asin [4]string
	asin[0] = "B006LE5LA8"
	asin[1] = "B01M03YDPX"
	asin[2] = "B01N5OI9Z1"
	asin[3] = "B079TNNZ2R"

	var jan [4]string
	jan[0] = "4571382081018"
	jan[1] = "4571382081452"
	jan[2] = "4571382081568"
	jan[3] = "4571382081773"

	for i := 0; i < len(asin); i++ {
		amazon(asin[i])
		surugaya(jan[i])
	}
}


	
		
	



// func toAbsUrl(baseurl *url.URL, weburl string) string {
// 	relurl, err := url.Parse(weburl)
// 	if err != nil {
// 		return ""
// 	}
// 	absurl := baseurl.ResolveReference(relurl)
// 	return absurl.String()
// }