package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
)

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

func main() {

	http.HandleFunc("/address", func(w http.ResponseWriter, r *http.Request) {

		// クエリパラメータを取得
		postalCode := r.URL.Query().Get("postal_code")

		url := "https://geoapi.heartrails.com/api/json?method=searchByPostal&postal=" + postalCode

		response, err := http.Get(url)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer response.Body.Close()

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

	})

	// サーバーを起動
	serverAddr := ":8080"
	fmt.Printf("Server listening on %s...\n", serverAddr)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}

func calMaxDistance(locations []Location) float64 {
	r := 6371.0
	xTokyo := 139.7673068
	yTokyo := 35.6809591
	pi := math.Pi
	dMax := 0.0

	for _, loc := range locations {

		// x(緯度),y(経度)をfloat64型に変換
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
