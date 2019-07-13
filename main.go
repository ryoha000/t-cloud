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
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/srinathgs/mysqlstore"
	"golang.org/x/crypto/bcrypt"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Game struct {
	ID          int    `json:"id,omitempty"  db:"id"`
	GameName   	string `json:"gamename,omitempty"  db:"gamename"`
	Sellday		string `json:"sellday,omitempty"  db:"sellday"`
	BrandName   string `json:"brandname,omitempty"  db:"brandname"`
	Median		int	   `json:"median,omitempty"  db:"median"`
	Stdev	    int    `json:"stdev,omitempty"  db:"stdev"`
	Count2		int    `json:"count2,omitempty"  db:"count2"`
	Shoukai		string `json:"shoukai,omitempty"  db:"shoukai"`
}

var (
	db *sqlx.DB
)

func main() {
	db, err := sql.Open("mysql", "root@/mydb1")
    if err != nil {
        panic(err.Error())
    }
    defer db.Close()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(session.Middleware(store))

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	e.GET("/games/:gameID", getGameInfoHandler)
	e.POST("/login", postLoginHandler)
	e.POST("/signup", postSignUpHandler)
	e.POST("/title", searchTitleHandler)

	withLogin := e.Group("")
	withLogin.Use(checkLogin)
	// withLogin.GET("/cities/:cityName", getCityInfoHandler)
	withLogin.GET("/mypage", getIntentionHandler)
	withLogin.POST("/rightButton", rightButtonHandler)

	e.Start(":4000")
}

type LoginRequestBody struct {
	Username string `json:"username,omitempty" form:"username"`
	Password string `json:"password,omitempty" form:"password"`
}

type User struct {
	ID          int    `json:"id,omitempty"  db:"ID"`
	Username   string `json:"username,omitempty"  db:"Username"`
	HashedPass string `json:"-"  db:"HashedPass"`
}

func postSignUpHandler(c echo.Context) error {
	req := LoginRequestBody{}
	c.Bind(&req)

	// もう少し真面目にバリデーションするべき
	if req.Password == "" || req.Username == "" {
		// エラーは真面目に返すべき
		return c.String(http.StatusBadRequest, "項目が空です")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("bcrypt generate error: %v", err))
	}

	// ユーザーの存在チェック
	var count int

	err = db.Get(&count, "SELECT COUNT(*) FROM users WHERE Username=?", req.Username)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err))
	}

	if count > 0 {
		return c.String(http.StatusConflict, "ユーザーが既に存在しています")
	}

	_, err = db.Exec("INSERT INTO users (Username, HashedPass) VALUES (?, ?)", req.Username, hashedPass)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err))
	}
	return c.NoContent(http.StatusCreated)
}

func postLoginHandler(c echo.Context) error {
	req := LoginRequestBody{}
	c.Bind(&req)

	user := User{}
	err := db.Get(&user, "SELECT * FROM users WHERE username=?", req.Username)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err))
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPass), []byte(req.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return c.NoContent(http.StatusForbidden)
		} else {
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	sess, err := session.Get("sessions", c)
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusInternalServerError, "something wrong in getting session")
	}
	sess.Values["userName"] = req.Username
	sess.Save(c.Request(), c.Response())

	return c.NoContent(http.StatusOK)
}

func checkLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}

		if sess.Values["userName"] == nil {
			return c.String(http.StatusForbidden, "please login")
		}
		c.Set("userName", sess.Values["userName"].(string))

		return next(c)
	}
}

func getIntentionHandler(c echo.Context) error {
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	
	userName := sess.Values["userName"]

	condition := db.Query("SELECT gamename, median, nowintention FROM gamelist JOIN intention ON id = gameid WHERE username=?", userName)
	if condition == "" {
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, condition)
}

func getGameInfoHandler(c echo.Context) error {
	gameID := c.Param("gameID")
	rows, err := db.Query("SELECT aws, jan FROM aws_jan WHERE gameid=?", gameID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var aws string
		var jan string 
		if err := rows.Scan(&aws, &jan); err != nil {
			log.Fatal(err)
		}
		fmt.Println(aws, jan)
		amazon(aws)
		surugaya(jan)
	}
	game := Game{}
	db.Get(&game, "SELECT gameid, gamename, sellday, brandid, median, stdev, count2, shoukai FROM gamelist WHERE gameid=?", gameID)
	if game.GameName == "" {
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, game)
}

func rightButtonHandler(c echo.Context) error {
	req := gameid
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	
	userName := sess.Values["userName"]
	nowIntenton := db.QueryRow("SELECT nowintention FROM intention WHERE username=? and gameid=?", userName, gameid)
	result, err := db.Exec("UPDATE intention SET nowintention = ? WHERE username=? and gameid = ?", nowintention-1,userName, gameid)
	if err != nil {
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK)
}

func searchTitleHandler(c echo.Context) error {
	req := word
	rows,verr := db.Query("SELECT id, gamename, median FROM gamelist WHERE gamename=?", word)
	if err != nil {
		log.Fatal(err)
	}
	if kekka == "" {
		return c.NoContent(http.StatusNotFound)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var gamename string 
		var median string
		if err := rows.Scan(&id, &gamename, &median); err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, gamename, median)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return c.JSON(http.StatusOK, id, gamename, median)
}

func amazon(as string)(hontai string,souryo string) {
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
 
	hontai = doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > span").Text()
	souryo = doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > p > span > span.olpShippingPrice").Text()
	if len(hontai) < 1 {
		fmt.Printf("中古なし")
	} else {
		fmt.Println("Amazon：￥" + hontai[23:])
		if len(souryo) < 6 {
			fmt.Printf("送料無料")
		} else {
			fmt.Println("送料：￥" + souryo[7:])
		}
		fmt.Printf("AmazonURL:" + sc_url)
	}
	
	return hontai,souryo
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

func surugaya(jan string) (nedan[6] string,urls[6] string) {
		//スクレイピング対象URLを設定
		url := "https://www.suruga-ya.jp/search?category=&search_word=&bottom_detail_search_bookmark=1&gtin=" + jan + "&id_s=&jan10=&mpn="
	 
		//goquery、ページを取得
		res, err := http.Get(url)
		if err != nil {
		}
		defer res.Body.Close()
	 
		utfBody := transform.NewReader(bufio.NewReader(res.Body), japanese.ShiftJIS.NewDecoder())
	 
		doc, err := goquery.NewDocumentFromReader(utfBody)
		if err != nil{
			panic(err)
		}

	 

		// var nedan [6]string
		for i := 1; i < 4; i++ {
			var s string
			s = strconv.Itoa(i)
			kakaku := doc.Find("#search_result > div > div:nth-child(" + s + ") > div.item_price > p:nth-child(1) > span > strong")
			link,exists := doc.Find("#search_result > div.item_box.first_item > div:nth-child(" + s + ") > div.item_detail > p.title > a").Attr("href")
			if exists != true {
				nedan[i-1] = ""
				urls[i-1] = ""
			} else {
			if len(kakaku.Text()) < 5 {
				nedan[i-1] = ""
				urls[i-1] = ""
			} else {
				nedan[i-1] = kakaku.Text()
				urls[i-1] = "https://www.suruga-ya.jp/search?category=&search_word=&bottom_detail_search_bookmark=1&gtin=" + jan + "&id_s=&jan10=&mpn=" + link
				fmt.Println("駿河屋：￥" + nedan[i-1][6:])
				fmt.Println("駿河屋URL：" + urls[i-1])
				           
			}
			}
			
		}
		for i := 1; i < 4; i++ {
			var v string
			v = strconv.Itoa(i)
			kakaku := doc.Find("#search_result > div:nth-child(2) > div:nth-child(" + v + ") > div.item_price > p:nth-child(1) > span > strong")
			link,exists := doc.Find("#search_result > div:nth-child(2) > div:nth-child(" + v + ") > div.item_detail > p.title > a").Attr("href")
			if exists != true {
				nedan[i+2] = ""
				urls[i+2] = ""
			} else {
			if len(kakaku.Text()) < 5 {
				nedan[i+2] = ""
				urls[i+2] = ""
			} else {
				nedan[i+2] = kakaku.Text()
				urls[i+2] = "https://www.suruga-ya.jp/search?category=&search_word=&bottom_detail_search_bookmark=1&gtin=" + jan + "&id_s=&jan10=&mpn=" + link
				fmt.Println(nedan[i+2][6:])
				fmt.Println("駿河屋URL：" + urls[i+2])
			}
			}
		}
		return nedan,urls
}
