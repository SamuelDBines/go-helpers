package webview
package app

import webview "github.com/webview/webview"

type WebViewConfig struct {
	URL       string
	Title     string
	Width     int
	Height    int
	Resizable bool
	Debug     bool
}

func NewWebView(cfg WebViewConfig) webview.WebView {
	w := webview.New(cfg.Debug)

	w.SetTitle(cfg.Title)

	hint := webview.HintNone
	if !cfg.Resizable {
		hint = webview.HintFixed
	}

	w.SetSize(cfg.Width, cfg.Height, hint)
	w.Navigate(cfg.URL)

	return w
}

func Run(w webview.WebView) {
	defer w.Destroy()
	w.Run()
}

func Terminate(w webview.WebView) {
	w.Terminate()
}