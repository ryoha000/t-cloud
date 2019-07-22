package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/japanese"
	"bufio"
	"strconv"
	// "strings"
	"os"
	"log"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/srinathgs/mysqlstore"
	"golang.org/x/crypto/bcrypt"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type NullInt64 struct {    // 新たに型を定義
    sql.NullInt64
}

type NullString struct {    // 新たに型を定義
    sql.NullString
}

type Game struct {
	GameID      int    `json:"gameid,omitempty"  db:"gameid"`
	GameName   	NullString `json:"gamename,omitempty"  db:"gamename"`
	Sellday		NullString `json:"sellday,omitempty"  db:"sellday"`
	BrandID   	NullString `json:"brandid,omitempty"  db:"brandid"`
	Median		NullInt64	   `json:"median,omitempty"  db:"median"`
	Stdev	    NullInt64    `json:"stdev,omitempty"  db:"stdev"`
	Count2		NullInt64    `json:"count2,omitempty"  db:"count2"`
	Shoukai		NullString `json:"shoukai,omitempty"  db:"shoukai"`
	// NowIntention int	   `json:"nowintention,omitempty"  db:"nowintention"`
}

type GameIntention struct {
	GameID      int    `json:"gameid,omitempty"  db:"gameid"`
	GameName   	NullString `json:"gamename,omitempty"  db:"gamename"`
	Median		NullInt64	   `json:"median,omitempty"  db:"median"`
	NowIntention NullInt64	   `json:"nowintention,omitempty"  db:"nowintention"`
}

type AwsJan struct {
	GameID		int		`json:"gameid,omitempty"  db:"gameid"`
	Aws			string	`json:"aws,omitempty"  db:"aws"`
	Jan			NullString	`json:"jan,omitempty"  db:"jan"`
}

var (
	db *sqlx.DB
)

func main() {
	_db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOSTNAME"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE")))
	if err != nil {
		log.Fatalf("Cannot Connect to Database: %s", err)
	}
	db = _db

	store, err := mysqlstore.NewMySQLStoreFromConnection(db.DB, "sessions", "/", 60*60*24*14, []byte("secret-token"))
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
        log.Printf("failed to ping by error '%#v'", err)
        return
    }

	// store, err := mysqlstore.NewMySQLStoreFromConnection(db.DB, "sessions", "/", 60*60*24*14, []byte("secret-token"))
	// if err != nil {
	// 	panic(err)
	// }

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(session.Middleware(store))

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	e.GET("/games/:gameID", getGameInfoHandler)
	e.POST("/login", postLoginHandler)
	e.POST("/signup", postSignUpHandler)
	// e.POST("/title", searchTitleHandler)

	withLogin := e.Group("")
	withLogin.Use(checkLogin)
	// withLogin.GET("/cities/:cityName", getCityInfoHandler)
	withLogin.GET("/mypage", getIntentionHandler)
	withLogin.GET("/whoami", getWhoAmIHandler)
	// withLogin.POST("/rightButton", rightButtonHandler)
	// game := Game{}
	// if err = db.Get(&game, "SELECT gameid, gamename, sellday, brandid, median, stdev, count2, shoukai FROM gamelist WHERE gameid=27000"); err !=nil {
	// 	log.Printf("failed get by error '%#v'", err)
	// 	return
	// }
	// fmt.Println(game)
	
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

type SearchRequestBody struct {
	Word string `json:"word,omitempty" form:"word"`
}

type ButtonRequestBody struct {
	GameID int `json:"gameid,omitempty" form:"gameid"`
}

type Joutai struct {
	GameID int `json:"gameid,omitempty" form:"gameid"`
	Username   string `json:"username,omitempty"  db:"Username"`
	NowIntention NullInt64	   `json:"nowintention,omitempty"  db:"nowintention"`
}

type Amazon struct{
	Hontai		string
	Souryo		string
	URL			string
}

type Surugaya struct{
	Nedan		string
	URL			string
}

type Kekka struct {
	GameID 		int `json:"gameid,omitempty" form:"gameid"`
	GameName   	string `json:"gamename,omitempty"  db:"gamename"`
	Median		NullInt64	   `json:"median,omitempty"  db:"median"`
}

func postSignUpHandler(c echo.Context) error {
	db, err := sqlx.Open("mysql", "root@/mydb1")

    if err != nil {
        panic(err.Error())
    }
	defer db.Close()
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
	db,err := sqlx.Open("mysql", "root@/mydb1")
	defer db.Close()
	req := LoginRequestBody{}
	c.Bind(&req)

	user := User{}
	err = db.Get(&user, "SELECT * FROM users WHERE username=?", req.Username)
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

type Me struct {
	Username string `json:"username,omitempty"  db:"username"`
}

func getWhoAmIHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, Me{
		Username: c.Get("userName").(string),
	})
}

func getIntentionHandler(c echo.Context) error {
	db, err := sqlx.Open("mysql", "root@/mydb1")

    if err != nil {
        panic(err.Error())
    }
	defer db.Close()
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	
	userName := sess.Values["userName"]
	conditions := []GameIntention{}
	db.Select(&conditions,"SELECT gameid, gamename, median, nowintention FROM gamelist JOIN intention ON id = gameid WHERE username=?", userName)
	// defer rows.Close()
	// for rows.Next() {
	// 	condition := GameIntention{}
	// 	var gameid int
	// 	var gamename string
	// 	var median int
	// 	var NowIntention int
	// 	if err := rows.Scan(&gameid, &gamename, &median,&NowIntention); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	conditions = append(conditions,condition)
	// }
	// if err := rows.Err(); err != nil {
	// 	log.Fatal(err)
	// }
	// // if condition == nil {
	// // 	return c.NoContent(http.StatusNotFound)
	// // }

	return c.JSON(http.StatusOK, conditions)
	// fmt.Fprint(w, GameIntention(conditions))
	// return
}

func getGameInfoHandler(c echo.Context) error {

	gameID := c.Param("gameID")
	AJ := []AwsJan{}
	// Ama := []Amazon{}
	// Suru := []Surugaya{}
	amaP := []string{}
	amaS := []string{}
	amaU := []string{}
	db.Select(&AJ,"SELECT aws, jan FROM a_j WHERE gameid=?", gameID)
	for i := 0; i < len(AJ); i++ {
		as := AJ[i].Aws
		var hontai string
		var souryo string
		var url string
		url = "https://www.amazon.co.jp/gp/offer-listing/" + as + "/ref=dp_olp_used?ie=UTF8&condition=used"
		amaU = append(amaU, url)
		//goquery、ページを取得
		res, err := http.Get(url)
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
			amaP = append(amaP,"中古なし")
			amaS = append(amaS,"")
		} else {
			fmt.Println("Amazon：￥" + hontai[23:])
			amaP = append(amaP,"￥" + hontai[23:])
			if len(souryo) < 6 {
				fmt.Printf("送料無料")
				amaS = append(amaS,"送料無料")
			} else {
				fmt.Println("送料：￥" + souryo[7:])
				amaS = append(amaS,"送料：￥" + souryo[7:])
			}
			fmt.Printf("AmazonURL:" + url)
		}



		// jan := AJ[i].Jan
		// Suru[i] = surugaya(jan)
	}
	game := Game{}
	db.Get(&game, "SELECT gameid, gamename, sellday, brandid, median, stdev, count2, shoukai FROM gamelist WHERE gameid=?", gameID)
	// if game.GameName == "" {
	// 	return c.NoContent(http.StatusNotFound)
	// }

	return c.JSON(http.StatusOK, amaP)
}

// func rightButtonHandler(c echo.Context) error {
// 	req := ButtonRequestBody{}
// 	c.Bind(&req)
// 	gameid := req.GameID
// 	sess, err := session.Get("sessions", c)
// 		if err != nil {
// 			fmt.Println(err)
// 			return c.String(http.StatusInternalServerError, "something wrong in getting session")
// 		}
	
// 	userName := sess.Values["userName"]
// 	nowIntenton := Joutai{}
// 	nowIntention = db.QueryRow("SELECT nowintention FROM intention WHERE username=? and gameid=?", userName, gameid)
// 	result, err := db.Exec("UPDATE intention SET nowintention = ? WHERE username=? and gameid = ?", nowIntention.NowIntention-1,userName, gameid)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	return c.NoContent(http.StatusOK)
// }

func searchTitleHandler(c echo.Context) error {
	db, err := sqlx.Open("mysql", "root@/mydb1")

    if err != nil {
        panic(err.Error())
    }
	defer db.Close()
	req := SearchRequestBody{}
	c.Bind(&req)
	word := req.Word
	kensaku := []Kekka{}
	db.Select(&kensaku,"SELECT gameid, gamename, median FROM gamelist WHERE gamename=?", word)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if rows == "" {
	// 	return c.NoContent(http.StatusNotFound)
	// }
	// defer rows.Close()
	// kensaku := []Kekka{}
	// for rows.Next() {
	// 	kekka := Kekka{}
	// 	var gameid int
	// 	var gamename string
	// 	var median int
	// 	if err := rows.Scan(&gameid, &gamename, &median); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	kensaku = append(kensaku,kekka)
	// }
	// if err := rows.Err(); err != nil {
	// 	log.Fatal(err)
	// }
	return c.JSON(http.StatusOK, kensaku)
}

func amazon(as string)(hontai string,souryo string,url string) {
	//スクレイピング対象URLを設定
	url = "https://www.amazon.co.jp/gp/offer-listing/" + as + "/ref=dp_olp_used?ie=UTF8&condition=used"
 
	//goquery、ページを取得
	res, err := http.Get(url)
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
		hontai = ""
		souryo = ""
	} else {
		fmt.Println("Amazon：￥" + hontai[23:])
		if len(souryo) < 6 {
			fmt.Printf("送料無料")
		} else {
			fmt.Println("送料：￥" + souryo[7:])
		}
		fmt.Printf("AmazonURL:" + url)
	}
	
	return hontai,souryo,url
}

// func sofmap(jan string) {
// 		//スクレイピング対象URLを設定
// 		sc_url := "https://a.sofmap.com/search_result.aspx?mode=SEARCH&gid=&gid=&gid=&keyword_and=&keyword_or=&keyword_not=&product_maker=&product_name=&product_code=&jan_code=" + jan + "&product_type=NEW&product_type=USED&price_from=&price_to=&sale_date_from_year=&sale_date_from_month=&sale_date_from_day=&sale_date_to_year=&sale_date_to_month=&sale_date_to_day=&reserve_date_from_year=&reserve_date_from_month=&reserve_date_from_day=&reserve_date_to_year=&reserve_date_to_month=&reserve_date_to_day=&order_by=DEFAULT&styp=p_kwsk"
	 
// 		//goquery、ページを取得
// 		res, err := http.Get(sc_url)
// 		if err != nil {
// 			// handle error
// 		}
// 		defer res.Body.Close()
	 
// 		utfBody := transform.NewReader(bufio.NewReader(res.Body), japanese.ShiftJIS.NewDecoder())
	 
// 		doc, err := goquery.NewDocumentFromReader(utfBody)
// 		if err != nil{
// 		  panic(err)
// 		}
	 
// 		// 掲載イベントURL一覧を取得
// 		// doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn").Each(func(i int, s *goquery.Selection) {
// 		// 	// ブログのタイトルとタグを取得
// 		// 	title := s.Find("span").Text()
		
// 		// 	fmt.Printf(title)
// 		//   })
// 		a := doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > span.a-size-large.a-color-price.olpOfferPrice.a-text-bold").Text()
// 		b := doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > p > span > span.olpShippingPrice").Text()
// 		fmt.Println(a[23:])
// 		c := strings.Index(b, "5")
// 		if c == -1 {
// 			fmt.Printf("")
// 		} else {
// 			fmt.Println(b[7:])
// 		}
// }

func surugaya(jan NullString) (nedan[6] string,urls[6] string) {
		//スクレイピング対象URLを設定
		if !jan.Valid {
			for i := 0; i < 6; i++{
				nedan[i] = ""
				urls[i] = ""
			}
			return nedan,urls
		} else {
			url := "https://www.suruga-ya.jp/search?category=&search_word=&bottom_detail_search_bookmark=1&gtin=" + jan.String + "&id_s=&jan10=&mpn="
		
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
					urls[i-1] = "https://www.suruga-ya.jp/search?category=&search_word=&bottom_detail_search_bookmark=1&gtin=" + jan.String + "&id_s=&jan10=&mpn=" + link
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
					urls[i+2] = "https://www.suruga-ya.jp/search?category=&search_word=&bottom_detail_search_bookmark=1&gtin=" + jan.String + "&id_s=&jan10=&mpn=" + link
					fmt.Println(nedan[i+2][6:])
					fmt.Println("駿河屋URL：" + urls[i+2])
				}
				}
			}
			return nedan,urls
		}
}
