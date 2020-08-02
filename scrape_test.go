package main

import (
	"testing"
)

func TestScrape(t *testing.T) {
	var tests = []struct {
		url       string
		title     string
		author    string
		price     int
		publisher string
		isbn      string
	}{
		{
			"https://www.kinokuniya.co.jp/f/dsg-01-9784065170069",
			"詳解 確率ロボティクス―Pythonによる基礎アルゴリズムの実装",
			"上田 隆一【著】",
			3900,
			"講談社",
			"9784065170069"},
		{
			"https://www.kinokuniya.co.jp/f/dsg-02-9783662575604",
			"グラフ理論(テキスト・第5版)Graph Theory (Graduate Texts in Mathematics .173) (5. Aufl. 2018. xviii, 428 S. 119 SW-Abb. 235 mm)",
			"Diestel, Reinhard",
			7178,
			"SPRINGER, BERLIN; SPRINGER BERLIN HEIDELBERG; SPRINGE",
			"9783662575604",
		},
		{
			"https://www.kinokuniya.co.jp/f/dsg-01-9784873119175",
			"Effective Python―Pythonプログラムを改良する90項目 (第2版)",
			"ブレット・スラットキン、黒川利明",
			3600,
			"オライリー・ジャパン",
			"9784873119175",
		},
		{
			"https://www.kinokuniya.co.jp/f/dsg-01-9784621300251",
			"ADDISON-WESLEY PROFESSIONAL CO プログラミング言語Go",
			"ドノバン,アラン・A.A.〈Donovan,Alan A.A.〉、カーニハン,ブライアン・W.【著】〈Kernighan,Brian W.〉、柴田 芳樹【訳】",
			3800,
			"丸善出版",
			"9784621300251",
		},
	}

	const cacheFilePath = ".cache"
	url := make([]string, 1)
	for _, test := range tests {
		url[0] = test.url
		books, err := scrape(url, cacheFilePath)
		if err != nil {
			t.Errorf("%v failed: %v", test.url, err)
		} else if test.title != books[0].Title {
			t.Errorf("Scraping %v results in title %q, want %q", url, books[0].Title, test.title)
		} else if test.author != books[0].Author {
			t.Errorf("Scraping %v results in author %q, want %q", url, books[0].Author, test.author)
			// } else if test.price != books[0].Price {
			// 	t.Errorf("Scraping %v results in price %v, want %v", url, books[0].Price, test.price)
		} else if test.publisher != books[0].Publisher {
			t.Errorf("Scraping %v results in publisher %q, want %q", url, books[0].Publisher, test.publisher)
		} else if test.isbn != books[0].Isbn {
			t.Errorf("Scraping %v results in isbn %q, want %q", url, books[0].Isbn, test.isbn)
		}
	}
}
