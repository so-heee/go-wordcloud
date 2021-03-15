package main

import (
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/psykhi/wordclouds"
	mecab "github.com/shogo82148/go-mecab"
)

const (
	ipadicDir = "/usr/local/lib/mecab/dic/mecab-ipadic-neologd"
	csvOutput = "csvWordCloud.png"
	dbOutput  = "dbWordCloud.png"
)

type maskConf struct {
	File  string     `json:"file"`
	Color color.RGBA `json:"color"`
}

type wcConf struct {
	FontMaxSize     int          `json:"font_max_size"`
	FontMinSize     int          `json:"font_min_size"`
	RandomPlacement bool         `json:"random_placement"`
	FontFile        string       `json:"font_file"`
	Colors          []color.RGBA `json:"colors"`
	Width           int          `json:"width"`
	Height          int          `json:"height"`
	Mask            maskConf     `json:"mask"`
}

var wcColors = []color.RGBA{
	{0x0f, 0x9d, 0x58, 0xFF},
	{0x42, 0x85, 0xf4, 0xFF},
	{0xdb, 0x44, 0x37, 0xFF},
	{0xf4, 0xb4, 0x00, 0xFF},
}

var defaultConf = wcConf{
	FontMaxSize:     700,
	FontMinSize:     10,
	RandomPlacement: false,
	FontFile:        "font_1_honokamarugo_1.1.ttf",
	Colors:          wcColors,
	Width:           2048,
	Height:          2048,
	Mask: maskConf{"mask.png", color.RGBA{
		R: 0x00,
		G: 0x00,
		B: 0xff,
		A: 0xFF,
	}},
}

type content struct {
	Id        int       `gorm:"primaryKey"`
	Title     string    `gorm:"column:title"`
	Body      string    `gorm:"column:body"`
	createdAt time.Time `gorm:"column:created_at"`
}

func main() {

	// CSVファイル
	csvText, err := readCSV()
	if err != nil {
		log.Fatal(err)
	}
	csvWordMap, err := parseToNodeMap(csvText)
	if err != nil {
		log.Fatal(err)
	}

	csvImg := drawWordCloud(sortByValue(csvWordMap))

	if err := fileOutput(csvImg, csvOutput); err != nil {
		log.Fatal(err)
	}

	// DBから
	db, err := gormConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbWordMap, err := getContentsWords(db)
	if err != nil {
		log.Fatal(err)
	}

	dbImg := drawWordCloud(sortByValue(dbWordMap))

	if err := fileOutput(dbImg, dbOutput); err != nil {
		log.Fatal(err)
	}
}

func readCSV() (string, error) {
	file, err := os.Open("sample.csv")
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var line []string
	var textLine string

	for {
		line, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		textLine = textLine + strings.Join(strings.Fields(strings.Join(line, ",")), "")
	}

	return textLine, nil
}

func gormConnect() (*gorm.DB, error) {
	CONNECT := "user:password@tcp(127.0.0.1:3307)/sample_db?parseTime=true"
	db, err := gorm.Open("mysql", CONNECT)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getContentsWords(db *gorm.DB) (map[string]int, error) {

	utc, _ := time.LoadLocation("UTC")
	lastWeek := time.Now().In(utc).AddDate(0, 0, -7).Truncate(24 * time.Hour)
	// lastWeek2 := time.Now().In(utc).AddDate(0, 0, -222).Truncate(24 * time.Hour)

	fmt.Println(lastWeek)
	// fmt.Println(lastWeek2)

	contents := []content{}
	db.Where("created_at > ?", lastWeek).Find(&contents)
	// db.Table("spot_review").Where("created_date_time BETWEEN ? AND ?", lastWeek2, lastWeek).Find(&reviews)

	comments := [][]string{}
	for _, c := range contents {
		words, err := parseToNode(c.Body)
		if err != nil {
			return nil, err
		}
		comments = append(comments, unique(words))
	}
	fmt.Println(len(comments))
	wordMap := make(map[string]int)
	for _, wl := range comments {
		for _, w := range wl {
			wordMap[w]++
		}
	}
	return wordMap, nil
}

func parseToNode(text string) ([]string, error) {
	mecab, err := mecab.New(map[string]string{"dicdir": ipadicDir})
	if err != nil {
		return nil, err
	}
	defer mecab.Destroy()

	node, err := mecab.ParseToNode(text)
	if err != nil {
		return nil, err
	}
	words := []string{}
	for ; !node.IsZero(); node = node.Next() {
		features := strings.Split(node.Feature(), ",")
		//fmt.Printf("%s\t%s\n", node.Surface(), node.Feature())
		if features[0] == "名詞" && features[1] == "一般" || features[1] == "固有名詞" {
			//fmt.Printf("%s %s\n", node.Surface(), node.Feature())
			if !contains(StopWordJPN, node.Surface()) {
				words = append(words, node.Surface())
			}
		}
	}
	return words, nil
}

func parseToNodeMap(text string) (map[string]int, error) {
	mecab, err := mecab.New(map[string]string{"dicdir": ipadicDir})
	if err != nil {
		return nil, err
	}
	defer mecab.Destroy()

	node, err := mecab.ParseToNode(text)
	if err != nil {
		return nil, err
	}
	wordMap := make(map[string]int)
	for ; !node.IsZero(); node = node.Next() {
		features := strings.Split(node.Feature(), ",")
		//fmt.Printf("%s\t%s\n", node.Surface(), node.Feature())
		if features[0] == "名詞" && features[1] == "一般" || features[1] == "固有名詞" {
			//fmt.Printf("%s %s\n", node.Surface(), node.Feature())
			wordMap[node.Surface()]++
		}
	}

	return wordMap, nil
}

func drawWordCloud(words map[string]int) image.Image {

	colors := make([]color.Color, 0)
	for _, c := range defaultConf.Colors {
		colors = append(colors, c)
	}

	var boxes []*wordclouds.Box
	// boxes := wordclouds.Mask(
	// 	defaultConf.Mask.File,
	// 	defaultConf.Width,
	// 	defaultConf.Height,
	// 	defaultConf.Mask.Color,
	// )

	w := wordclouds.NewWordcloud(words,
		wordclouds.FontFile(defaultConf.FontFile),
		wordclouds.FontMaxSize(defaultConf.FontMaxSize),
		wordclouds.FontMinSize(defaultConf.FontMinSize),
		wordclouds.Colors(colors),
		wordclouds.MaskBoxes(boxes),
		wordclouds.Height(defaultConf.Height),
		wordclouds.Width(defaultConf.Width),
		wordclouds.RandomPlacement(defaultConf.RandomPlacement))

	return w.Draw()
}

func fileOutput(img image.Image, filename string) error {

	outputFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = png.Encode(outputFile, img)
	if err != nil {
		return err
	}
	return nil
}

func sortByValue(words map[string]int) map[string]int {
	names := make([]string, 0, len(words))
	for name := range words {
		names = append(names, name)
	}

	sort.Slice(names, func(i, j int) bool {
		return words[names[i]] > words[names[j]]
	})

	newWords := map[string]int{}
	for _, name := range names {
		newWords[name] = words[name]
		fmt.Printf("%-7v %v\n", name, words[name])
		if len(newWords) >= 300 {
			break
		}
	}
	fmt.Println(len(newWords))
	return newWords
}

func unique(who []string) []string {
	m := make(map[string]struct{})

	newList := make([]string, 0)

	for _, element := range who {
		// mapでは、第二引数にその値が入っているかどうかの真偽値が入っている
		if _, ok := m[element]; !ok {
			m[element] = struct{}{}
			newList = append(newList, element)
		}
	}
	return newList
}

func contains(sl []string, s string) bool {

	for _, v := range sl {
		if s == v {
			return true
		}
	}
	return false
}
