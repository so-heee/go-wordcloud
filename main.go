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
	text   = "政府は新型コロナウイルス特別措置法に基づき首都圏１都３県に発令中の緊急事態宣言について、延長後の期限通り、２１日までで解除する方向だ。再延長した理由だった病床の指標が改善傾向にあるため。週明け以降の感染状況を見極めたうえで、１８日にもコロナ対策本部を開いて決定する。【グラフ】コロナで亡くなった著名人と国内累計死者数推移　内閣官房の集計で延長前後（４日、１１日）の病床使用率を比較すると、東京３０％→２６％▽埼玉４１％→４０％▽千葉４６％→４２％▽神奈川２８％→２６％－と、いずれも緩やかながら改善傾向にあり、解除の目安である「ステージ３」の上限５０％を切り、下限の２０％に近づいている。こうした状況を踏まえ政府高官は「今のままなら大丈夫だ」として、３度目の宣言延長は見送る考えを示す。　一方、新規感染者数は下げ止まって「横ばいから微増傾向」（西村康稔経済再生担当相）に転じている。感染力が強いとされる変異株は全国的に広がりをみせており、主要駅や繁華街での人出増も懸念材料だ。　とはいえ、政府や専門家の間では、現在の対策ではこれ以上の改善は見込めないとの見方が強い。関係閣僚の一人は「宣言はもう効かない。早く解除するしかない」と語る。厚生労働省に助言する専門家組織が１１日に行った非公式の会合では、主要メンバーから「もう打つ手がない」との意見が出たという。　政府は解除後を見据えた対策を急ぐ。宣言を解除した地域の繁華街などで無症状者へのモニタリング検査を始めており、北海道や沖縄、首都圏でも実施する予定。感染再拡大の予兆があれば、改正コロナ特措法で新設した「蔓延（まんえん）防止等重点措置」を適用する構えだ。　また「第４波」に備え、都道府県に病床確保計画の見直しを要請する。田村憲久厚労相は第３波ピーク時の２倍の確保を例示している。"
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
	wordCounts, err := parseToNode("吾輩は猫である。名前はまだ無い。どこで生れたかとんと見当けんとうがつかぬ。何でも薄暗いじめじめした所でニャーニャー泣いていた事だけは記憶している。吾輩はここで始めて人間というものを見た。しかもあとで聞くとそれは書生という人間中で一番獰悪どうあくな種族であったそうだ。この書生というのは時々我々を捕つかまえて煮にて食うという話である。しかしその当時は何という考もなかったから別段恐しいとも思わなかった。ただ彼の掌てのひらに載せられてスーと持ち上げられた時何だかフワフワした感じがあったばかりである。掌の上で少し落ちついて書生の顔を見たのがいわゆる人間というものの見始みはじめであろう。この時妙なものだと思った感じが今でも残っている。第一毛をもって装飾されべきはずの顔がつるつるしてまるで薬缶やかんだ。その後ご猫にもだいぶ逢あったがこんな片輪かたわには一度も出会でくわした事がない。のみならず顔の真中があまりに突起している。そうしてその穴の中から時々ぷうぷうと煙けむりを吹く。どうも咽むせぽくて実に弱った。これが人間の飲む煙草たばこというものである事はようやくこの頃知った。")
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
