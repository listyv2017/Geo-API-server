package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// MySQLの接続情報
const (
	DBUsername = "root"
	DBPassword = ""
	DBHost     = ""
	DBPort     = ""
	DBName     = "finatext"
)

//データベースを作成するSQL文
const createDatabaseSQL = `
CREATE DATABASE IF NOT EXISTS finatext;
`

// アクセスログを保存するためのテーブルを作成するSQL文
const createTableSQL = `
CREATE TABLE IF NOT EXISTS access_logs (
id INT AUTO_INCREMENT PRIMARY KEY,
    postal_code VARCHAR(8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP	
);

`

var db *sql.DB


type Location struct {
	City       string `json:"city"`
	CityKana   string `json:"city_kana"`
	Town       string `json:"town"`
	TownKana   string `json:"town_kana"`
	X          string `json:"x"`
	Y          string `json:"y"`
	Prefecture string `json:"prefecture"`
	Postal     string `json:"postal"`
}

type Response struct {
	Location []Location `json:"location"`
}

type PostalResponse struct {
	PostalCode       string  `json:"postal_code"`
	HitCount         int     `json:"hit_count"`
	Address          string  `json:"address"`
	TokyoStaDistance float64 `json:"tokyo_sta_distance"`
}

// PostalRequestCounter 構造体の定義
type PostalRequestCounter struct {
	PostalCode string `json:"postal_code"`
	Count      int    `json:"request_count"`
}

type AccessLogResponse struct {
	AccessLogs []PostalRequestCounter `json:"access_logs"`
}

func main() {

	//MySQLに接続
	connectMysql()

	//ログファイルの作成
	logFile, err := os.Create("server.log")
	if err != nil {
		log.Fatal("Error creating log file:", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// ログメッセージの出力
	log.Println("Server starting...")

	http.HandleFunc("/address", addressSearch)

	// サーバーを起動
	serverAddr := ":8080"
	fmt.Printf("Server listening on %s...\n", serverAddr)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	logFile.Close()

	// サーバー終了時にログメッセージを出力
	log.Println("Server shutting down...")
}

// アクセスログをデータベースに保存する関数
func saveAccessLogToDB(postalCode string) {
	createdAt := time.Now()

	// データベースにアクセスログを挿入するSQL文
	insertSQL := `
	INSERT INTO access_logs (postal_code, created_at)
	VALUES (?, ?);
	`

	if db == nil {
		fmt.Printf("error nil")
		return
	}

	// SQL文を実行してアクセスログを保存
	_, err := db.Exec(insertSQL, postalCode, createdAt)
	if err != nil {
		log.Printf("Error inserting access log: %v", err)
	}
}

//MySQLに接続
func connectMysql(){

	// MySQLデータベースに接続
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/", DBUsername, DBPassword, DBHost, DBPort)
	var err error // エラー変数を宣言
	db, err = sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to the MySQL database: %v", err)
	}
	defer db.Close()

	// データベース作成
    _, err = db.Exec(createDatabaseSQL)
    if err != nil {
        log.Fatalf("Error creating database: %v", err)
    }
    fmt.Println("Database created successfully.")

	// データベースを使用
	_, err = db.Exec("USE " + DBName)
	if err != nil {
    	log.Fatalf("Error using database: %v", err)
	}

	// アクセスログテーブルを作成
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
	fmt.Println("Database and table created successfully.")

}

//住所検索するイベントハンドラ
func addressSearch(w http.ResponseWriter, r *http.Request) {

	// クエリパラメータを取得
	queryParam := r.URL.Query().Get("postal_code")

	// クエリパラメータの値を変数に格納
	postalCode := queryParam

	// アクセスログを保存
	saveAccessLogToDB(postalCode)

	url := "https://geoapi.heartrails.com/api/json?method=searchByPostal&postal=" + postalCode

	// GETリクエストを送信してレスポンスを取得
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	// response.BodyをUTF-8に変換
	utf8Reader := transform.NewReader(response.Body, unicode.UTF8.NewDecoder())
	utf8ResponseBody, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		panic(err)
	}

	var Data struct {
		Response Response `json:"response"`
	}

	//decode　JSON→Go
	err = json.Unmarshal(utf8ResponseBody, &Data)
	if err != nil {
		panic(err)
	}

	if len(Data.Response.Location) == 0 {
		fmt.Println("error no location")
		return
	}

	subAddress := Data.Response.Location[0].Prefecture + Data.Response.Location[0].City
	commonTownPrefix := extractCommonTown(Data.Response.Location)
	hitCount := len(Data.Response.Location)
	tokyoStaDistance := calMaxDistance(Data.Response.Location)

	postalResponse := PostalResponse{
		PostalCode:       postalCode,
		HitCount:         hitCount,
		Address:          subAddress + commonTownPrefix,
		TokyoStaDistance: tokyoStaDistance,
	}

	//encode Go→JSON
	responseData, err := json.MarshalIndent(postalResponse, "", "    ")
	if err != nil {
		panic(err)

	}
	fmt.Println(string(responseData))

}

//レスポンスで返ってきた住所のうち東京まで一番遠い場所との距離
func calMaxDistance(locations []Location) float64 {
	r := 6371.0
	xTokyo := 139.7673068
	yTokyo := 35.6809591
	pi := math.Pi
	dMax := 0.0

	// x(緯度),y(経度)をfloat64型に変換
	for _, loc := range locations {
		x, err := strconv.ParseFloat(loc.X, 64)
		if err != nil {
			fmt.Println("変換エラー:", err)
			return 0
		}
		y, err := strconv.ParseFloat(loc.Y, 64)
		if err != nil {
			fmt.Println("変換エラー:", err)
			return 0
		}

		term1 := (x - xTokyo) * math.Cos(pi*(y+yTokyo)/360)
		term2 := (y - yTokyo)
		term3 := math.Pow(term1, 2) + math.Pow(term2, 2)

		d := (pi * r / 180) * math.Sqrt(term3)
		if d > dMax {
			dMax = d
		}

		dMax = math.Round(d*10) / 10
	}

	return dMax
}

//レスポンスで返ってきた住所のうち共通の部分の出力
func extractCommonTown(locations []Location) string {
	if len(locations) == 0 {
		return ""
	}

	commonTown := locations[0].Town

	for _, loc := range locations {

		commonTown = findCommonPrefix(commonTown, loc.Town)
	}

	return commonTown
}

func findCommonPrefix(s1, s2 string) string {

	commonPart := ""

	rune1 := []rune(s1)
	rune2 := []rune(s2)

	minLen := len(rune1)
	if len(rune2) < minLen {
		minLen = len(rune2)
	}

	for i := 0; i < minLen && rune1[i] == rune2[i]; i++ {
		commonPart += string(rune1[i])
	}
	return commonPart

}
