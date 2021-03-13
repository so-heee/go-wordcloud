package main

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/psykhi/wordclouds"
	mecab "github.com/shogo82148/go-mecab"
)

const (
	ipadic = "/usr/local/lib/mecab/dic/mecab-ipadic-neologd"
	text   = "吾輩は猫である。名前はまだ無い。どこで生れたかとんと見当けんとうがつかぬ。何でも薄暗いじめじめした所でニャーニャー泣いていた事だけは記憶している。吾輩はここで始めて人間というものを見た。しかもあとで聞くとそれは書生という人間中で一番獰悪どうあくな種族であったそうだ。この書生というのは時々我々を捕つかまえて煮にて食うという話である。しかしその当時は何という考もなかったから別段恐しいとも思わなかった。ただ彼の掌てのひらに載せられてスーと持ち上げられた時何だかフワフワした感じがあったばかりである。掌の上で少し落ちついて書生の顔を見たのがいわゆる人間というものの見始みはじめであろう。この時妙なものだと思った感じが今でも残っている。第一毛をもって装飾されべきはずの顔がつるつるしてまるで薬缶やかんだ。その後ご猫にもだいぶ逢あったがこんな片輪かたわには一度も出会でくわした事がない。のみならず顔の真中があまりに突起している。そうしてその穴の中から時々ぷうぷうと煙けむりを吹く。どうも咽むせぽくて実に弱った。これが人間の飲む煙草たばこというものである事はようやくこの頃知った。"
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
	{0x1b, 0x1b, 0x1b, 0xff},
	{0x48, 0x48, 0x4B, 0xff},
	{0x59, 0x3a, 0xee, 0xff},
	{0x65, 0xCD, 0xFA, 0xff},
	{0x70, 0xD6, 0xBF, 0xff},
}

var defaultConf = wcConf{
	FontMaxSize:     200,
	FontMinSize:     10,
	RandomPlacement: false,
	FontFile:        "font",
	Colors:          wcColors,
	Width:           2048,
	Height:          2048,
	Mask: maskConf{"", color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 0,
	}},
}

func main() {
	wordCounts, err := parseToNode(text)
	if err != nil {
		panic(err)
	}
	fmt.Println(wordCounts)

	colors := make([]color.Color, 0)
	for _, c := range defaultConf.Colors {
		colors = append(colors, c)
	}

	var boxes []*wordclouds.Box

	w := wordclouds.NewWordcloud(wordCounts,
		wordclouds.FontFile(defaultConf.FontFile),
		wordclouds.FontMaxSize(defaultConf.FontMaxSize),
		wordclouds.FontMinSize(defaultConf.FontMinSize),
		wordclouds.Colors(colors),
		wordclouds.MaskBoxes(boxes),
		wordclouds.Height(defaultConf.Height),
		wordclouds.Width(defaultConf.Width),
		wordclouds.RandomPlacement(defaultConf.RandomPlacement))

	img := w.Draw()

	outputFile, err := os.Create("wordcloud.png")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	err = png.Encode(outputFile, img)
	if err != nil {
		err = fmt.Errorf("pngのエンコードに失敗しました。%w", err)
		panic(err)
	}
}

func parseToNode(text string) (map[string]int, error) {

	mecab, err := mecab.New(map[string]string{"dicdir": ipadic})
	if err != nil {
		return nil, err
	}
	defer mecab.Destroy()

	node, err := mecab.ParseToNode(text)
	wordMap := make(map[string]int)
	for ; !node.IsZero(); node = node.Next() {
		features := strings.Split(node.Feature(), ",")
		fmt.Printf("%s\t%s\n", node.Surface(), node.Feature())
		if features[0] == "名詞" && features[1] == "一般" || features[1] == "固有名詞" {
			//fmt.Printf("%s %s\n", node.Surface(), node.Feature())
			wordMap[node.Surface()]++
		}
	}

	return wordMap, nil
}
