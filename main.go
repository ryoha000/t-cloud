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

type GameInfo struct {
	Game		Game	`json:"game,omitempty"`
	AmaSuru		[]AmaSuru `json:"amasuru,omitempty"`
	// NowIntention int	   `json:"nowintention,omitempty"  db:"nowintention"`
}

type Intention struct {
	Intention	int		`json:"intention,omitempty"  db:"intention"`
}

type ProfileGame struct {
	GameID      int    `json:"gameid,omitempty"  db:"gameid"`
	GameName   	NullString `json:"gamename,omitempty"  db:"gamename"`
	Median		NullInt64	   `json:"median,omitempty"  db:"median"`
}

type GameIntention struct {
	GameID      int    `json:"gameid,omitempty"  db:"gameid"`
	GameName   	NullString `json:"gamename,omitempty"  db:"gamename"`
	Median		NullInt64	   `json:"median,omitempty"  db:"median"`
	NowIntention	int		`json:"intention,omitempty"  db:"intention"`
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
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(session.Middleware(store))

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	e.POST("/login", postLoginHandler)
	e.POST("/signup", postSignUpHandler)
	// e.POST("/title", searchTitleHandler)

	withLogin := e.Group("")
	withLogin.Use(checkLogin)
	withLogin.GET("/games/:gameID", getGameInfoHandler)
	// withLogin.GET("/cities/:cityName", getCityInfoHandler)
	withLogin.GET("/mypage", getIntentionHandler)
	withLogin.GET("/whoami", getWhoAmIHandler)
	withLogin.POST("/rightButton", rightButtonHandler)
	withLogin.POST("/leftButton", leftButtonHandler)
	withLogin.POST("/title", searchTitleHandler)
	withLogin.POST("/brand", searchBrandHandler)
	withLogin.POST("/bought", boughtHandler)
	withLogin.POST("/ari", ariHandler)
	withLogin.POST("/imahax", imahaxHandler)
	withLogin.POST("/nai", naiHandler)

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
	Word string `json:"word,omitempty"`
}

type ButtonRequestBody struct {
	GameID int `json:"gameid,omitempty"`
}

type Amazon struct{
	AmaP		string	`json:"amap,omitempty"`
	Souryo		string	`json:"souryo,omitempty"`
	URL			string	`json:"url,omitempty"`
}

type Surugaya struct{
	SuruP		string	`json:"surup,omitempty"`
	URL			string	`json:"url,omitempty"`
}

type AmaSuru struct{
	Ama			Amazon		`json:"ama,omitempty"`
	Suru		[]Surugaya	`json:"amap,omitempty"`
}

type Kekka struct {
	GameID 		int `json:"gameid,omitempty" form:"gameid"`
	GameName   	string `json:"gamename,omitempty"  db:"gamename"`
	Median		NullInt64	   `json:"median,omitempty"  db:"median"`
	Brandname	NullString	 `json:"brandname,omitempty"  db:"brandname"`
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

type Me struct {
	Username string `json:"username,omitempty"  db:"username"`
}

func getWhoAmIHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, Me{
		Username: c.Get("userName").(string),
	})
}

func getIntentionHandler(c echo.Context) error {
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	userName := sess.Values["userName"]
	conditions := []GameIntention{}
	db.Select(&conditions,"SELECT gamelist.gameid, gamename, median, intention FROM gamelist inner JOIN intention_table ON gamelist.gameid = intention_table.gameid WHERE username=?", userName)


	return c.JSON(http.StatusOK, conditions)
}

func getGameInfoHandler(c echo.Context) error {
	gameID := c.Param("gameID")
	AJ := []AwsJan{}
	AS := []AmaSuru{}
	db.Select(&AJ,"SELECT aws, jan FROM a_j WHERE gameid=?", gameID)
	for i := 0; i < len(AJ); i++ {
		as := AJ[i].Aws
		a := amazon(as)
		jan := AJ[i].Jan
		s := surugaya(jan)
		AS = append(AS,AmaSuru{a,s})
	}
	game := Game{}
	db.Get(&game, "SELECT gameid, gamename, sellday, brandid, median, stdev, count2, shoukai FROM gamelist WHERE gameid=?", gameID)
	// if game.GameName == "" {
	// 	return c.NoContent(http.StatusNotFound)
	// }
	gameInfo := GameInfo{game,AS}
	return c.JSON(http.StatusOK, gameInfo)
}

func rightButtonHandler(c echo.Context) error {
	req := ButtonRequestBody{}
	c.Bind(&req)
	gameid := req.GameID
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	userName := sess.Values["userName"]
	db.Exec("UPDATE intention_table SET intention = intention -1 WHERE username=? and gameid = ?", userName, gameid)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	return c.NoContent(http.StatusOK)
}

func leftButtonHandler(c echo.Context) error {
	req := ButtonRequestBody{}
	c.Bind(&req)
	gameid := req.GameID
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	userName := sess.Values["userName"]
	db.Exec("UPDATE intention_table SET intention = intention +1 WHERE username=? and gameid = ?", userName, gameid)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	return c.NoContent(http.StatusOK)
}

func searchTitleHandler(c echo.Context) error {
	req := SearchRequestBody{}
	c.Bind(&req)
	word := req.Word
	fmt.Printf(word)
	kensaku := []Kekka{}
	if err := db.Select(&kensaku,"SELECT gameid, brandname, gamename, gamelist.median FROM gamelist inner join brandlist on brandlist.brandid = gamelist.brandid WHERE gamename like ?","'%" + word + "%'" ); err != nil{
		log.Printf("failed to ping by error '%#v'", err)
	}
	return c.JSON(http.StatusOK, kensaku)
}

func searchBrandHandler(c echo.Context) error {
	req := SearchRequestBody{}
	c.Bind(&req)
	word := req.Word
	kensaku := []Kekka{}
	db.Select(&kensaku,"SELECT gameid, brandname, gamename, gamelist.median FROM gamelist inner join brandlist on brandlist.brandid = gamelist.brandid WHERE brandname like ?","'%" + word + "%'" )
	return c.JSON(http.StatusOK, kensaku)
}

func boughtHandler(c echo.Context) error {
	req := ButtonRequestBody{}
	c.Bind(&req)
	gameid := req.GameID
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	userName := sess.Values["userName"]
	db.Exec("insert into intention_table (username,gameid,intention) values (?,?,3) on DUPLICATE KEY UPDATE intention = values(intention)", userName, gameid)
	return c.NoContent(http.StatusOK)
}

func ariHandler(c echo.Context) error {
	req := ButtonRequestBody{}
	c.Bind(&req)
	gameid := req.GameID
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	userName := sess.Values["userName"]
	db.Exec("insert into intention_table (username,gameid,intention) values (?,?,2) on DUPLICATE KEY UPDATE intention = values(intention)", userName, gameid)
	return c.NoContent(http.StatusOK)
}

func imahaxHandler(c echo.Context) error {
	req := ButtonRequestBody{}
	c.Bind(&req)
	gameid := req.GameID
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	userName := sess.Values["userName"]
	db.Exec("insert into intention_table (username,gameid,intention) values (?,?,1) on DUPLICATE KEY UPDATE intention = values(intention)", userName, gameid)
	return c.NoContent(http.StatusOK)
}

func naiHandler(c echo.Context) error {
	req := ButtonRequestBody{}
	c.Bind(&req)
	gameid := req.GameID
	sess, err := session.Get("sessions", c)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
	userName := sess.Values["userName"]
	db.Exec("insert into intention_table (username,gameid,intention) values (?,?,0) on DUPLICATE KEY UPDATE intention = values(intention)", userName, gameid)
	return c.NoContent(http.StatusOK)
}


func amazon(as string)(Ama Amazon) {
	//スクレイピング対象URLを設定
	url := "https://www.amazon.co.jp/gp/offer-listing/" + as + "/ref=dp_olp_used?ie=UTF8&condition=used"
	//goquery、ページを取得
	res, err := http.Get(url)
	if err != nil {
    	// handle error
	}
	defer res.Body.Close()
 
	utfBody := transform.NewReader(bufio.NewReader(res.Body), japanese.ShiftJIS.NewDecoder())
 
	doc, err := goquery.NewDocumentFromReader(utfBody)
 
	hontai := doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > span").Text()
	souryo := doc.Find("#olpOfferList > div > div > div:nth-child(3) > div.a-column.a-span2.olpPriceColumn > p > span > span.olpShippingPrice").Text()
	if len(hontai) < 1 {
		fmt.Printf("中古なし")
		hontai = "中古なし"
		souryo = "URLをクリックしてみたらあるかも"
		Ama = Amazon{hontai,souryo,url}
		
	} else {
		fmt.Println("Amazon：￥" + hontai[23:])
		if len(souryo) < 6 {
			fmt.Printf("送料無料")
			Ama = Amazon{hontai[23:32],"送料無料",url}
		} else {
			fmt.Println("送料：￥" + souryo[7:15])
			Ama = Amazon{hontai[23:32],souryo[7:15],url}
		}
		fmt.Printf("AmazonURL:" + url)
	}
	
	return Ama
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

func surugaya(jan NullString) (Suru []Surugaya) {
		//スクレイピング対象URLを設定
		// Suru = make([]Surugaya, 0)
		if !jan.Valid {			
			Suru = append(Suru,Surugaya{"JAN不明",""})			
			return Suru
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
				if i == 1{
					if exists != true {
						Suru = append(Suru,Surugaya{"駿河屋無し",""})
					}
				}
				if exists != true {
				} else {
				if len(kakaku.Text()) < 5 {
				} else {
					nedan := kakaku.Text()
					urls := "https://www.suruga-ya.jp/search?category=&search_word=&bottom_detail_search_bookmark=1&gtin=" + jan.String + "&id_s=&jan10=&mpn=" + link
					fmt.Println("駿河屋：￥" + nedan[6:])
					fmt.Println("駿河屋URL：" + urls)
					Suru = append(Suru,Surugaya{"￥"+ nedan[6:],urls})
				}
				}	
			}
			for i := 1; i < 4; i++ {
				var v string
				v = strconv.Itoa(i)
				kakaku := doc.Find("#search_result > div:nth-child(2) > div:nth-child(" + v + ") > div.item_price > p:nth-child(1) > span > strong")
				link,exists := doc.Find("#search_result > div:nth-child(2) > div:nth-child(" + v + ") > div.item_detail > p.title > a").Attr("href")
				if exists != true {
				} else {
				if len(kakaku.Text()) < 5 {
				} else {
					nedan := kakaku.Text()
					urls := "https://www.suruga-ya.jp/search?category=&search_word=&bottom_detail_search_bookmark=1&gtin=" + jan.String + "&id_s=&jan10=&mpn=" + link
					fmt.Println(nedan[6:])
					fmt.Println("駿河屋URL：" + urls)
					Suru = append(Suru,Surugaya{"￥" + nedan[6:],urls})
				}
				}
			}
			return Suru
		}
}
