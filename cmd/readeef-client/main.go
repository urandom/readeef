package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/zserge/webview"
)

var (
	cfg config
)

func main() {
	var err error
	cfg, err = readConfig()
	if err != nil {
		log.Println("Error reading config file:", err)
	}

	var w webview.WebView
	if cfg.URL == "" {
		w = openSettingsWindow()
	} else {
		w = openReadeefWindow(cfg.URL)
	}

	w.Run()
}

func settings(url string, width, height int) webview.Settings {
	return webview.Settings{
		Width:     width,
		Height:    height,
		Resizable: true,
		Title:     "readeef",
		URL:       url,
		Debug:     false,
		ExternalInvokeCallback: webviewRPC,
	}
}

func openSettingsWindow() webview.WebView {
	return webview.New(settings("data:text/html, "+url.PathEscape(settingsHTML), 640, 480))
}

func openReadeefWindow(u string) webview.WebView {
	w := webview.New(settings(u, 1280, 720))
	time.AfterFunc(time.Second, func() {
		onReadeefLoad(w)
	})

	return w
}

func webviewRPC(w webview.WebView, data string) {
	parts := strings.SplitN(data, ":", 2)
	switch parts[0] {
	case "set-url":
		u, err := url.Parse(parts[1])
		if err != nil {
			w.Dialog(webview.DialogTypeAlert, webview.DialogFlagError, "Error", fmt.Sprintf("Error: %s", err))
			return
		}
		if !u.IsAbs() {
			w.Dialog(webview.DialogTypeAlert, webview.DialogFlagError, "Error", fmt.Sprintf("URL not absolute: %s", parts[1]))
			return
		}
		go w.Exit()

		cfg.URL = u.String()
		writeConfig(cfg)

		openReadeefWindow(cfg.URL).Run()
	case "open":
		if err := open(parts[1]); err != nil {
			w.Dialog(webview.DialogTypeAlert, webview.DialogFlagError, "Error", err.Error())
		}
	}
}

func onReadeefLoad(w webview.WebView) {
	w.Dispatch(func() {
		w.Eval(`
document.body.addEventListener('open-link', (event) => {
	event.preventDefault();
	window.external.invoke("open:" + event.detail);
})
`)
	})
}

const settingsHTML = `
<!doctype html>
<html>
	<head>
		<script
			  src="https://code.jquery.com/jquery-3.3.1.slim.min.js"
			  integrity="sha256-3edrmyuQ0w65f8gfBsqowzjJe2iM6n0nKciPUp8y+7E="
			  crossorigin="anonymous"></script>
	</head>
	<body>
		<form>
			<h1>Please enter the url of your readeef server:</h1>
			<input name="server" placeholder="readeef url"><input type="submit"><br>
		</form>
		<script>

var rpcData = {
	targetURL: null,
};
$('form').on('submit', function(event) {
	event.preventDefault();
	window.external.invoke("set-url:" + $("[name=server]").val());
	setTimeout(processResult, 10);
});

function processResult() {
	if (rpcData.targetURL) {
		location = rpcData.targetURL;
	}
}
		</script>
	</body>
</html>
`
