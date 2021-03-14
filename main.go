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
	Id    int    `json:id`
	Title string `json:title`
	Body  string `json:body`
}

func main() {
	// CSVファイル
	csvText, err := readCSV()
	if err != nil {
		log.Fatal(err)
	}
	if err := generateWordCloud(csvText, csvOutput); err != nil {
		log.Fatal(err)
	}

	// DBから取得
	db, err := gormConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	bodies := getContents(db)
	if err := generateWordCloud(strings.Join(bodies, ""), dbOutput); err != nil {
		log.Fatal(err)
	}
}

func generateWordCloud(text string, filename string) error {
	words, err := parseToNode(text)
	if err != nil {
		return err
	}
	// fmt.Println(words)

	img := drawWordCloud(sortByValue(words))

	if err := fileOutput(img, filename); err != nil {
		return err
	}
	return nil
}

func gormConnect() (*gorm.DB, error) {
	CONNECT := "user:password@tcp(127.0.0.1:3307)/sample_db"
	db, err := gorm.Open("mysql", CONNECT)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getContents(db *gorm.DB) []string {
	contents := []content{}
	db.Find(&contents)

	bodies := []string{}
	for _, c := range contents {
		bodies = append(bodies, c.Body)
	}
	return bodies
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

func parseToNode(text string) (map[string]int, error) {

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
		if len(newWords) >= 150 {
			break
		}
	}
	fmt.Println(len(newWords))
	return newWords
}
