package vulinbox

import (
	"github.com/yaklang/yaklang/common/utils"
	"net/http"
	"strings"
)

const defaultRenderPage = `<!doctype html>
<html>
<head>
    <title>Example DEMO</title>

    <meta charset="utf-8" />
    <meta http-equiv="Content-type" content="text/html; charset=utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
	<style>
	.code-style {
		width: 100%; /* Set the width */
		background-color: #f8f8f8; /* Set the background color */
		border: 1px solid #ccc; /* Set the border */
		padding: 10px; /* Set the padding */
		font-family: "Courier New", monospace; /* Set the font */
		white-space: pre-wrap; /* Preserve white space and line wrapping */
		overflow-x: auto; /* If the content exceeds the width, display the scroll bar */
	}
	</style>

    <style type="text/css">

    body {
        background-color: #f0f0f2;
        margin: 0;
        padding: 0;
        font-family: -apple-system, system-ui, BlinkMacSystemFont, "Segoe UI", "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
        
    }
    div {
        width: 600px;
        margin: 5em auto;
        padding: 2em;
        background-color: #fdfdff;
        border-radius: 0.5em;
        box-shadow: 2px 3px 7px 2px rgba(0,0,0,0.02);
    }
    </style>    
</head>

<body>
<div>
	{{ .__innerHtml }}
</div>
</body>
</html>`

func DefaultRender(innerHtml any, writer http.ResponseWriter, request *http.Request, paramIns ...map[string]any) {
	DefaultRenderEx(false, innerHtml, writer, request, paramIns...)
}

func DefaultRenderEx(override bool, innerHtml any, writer http.ResponseWriter, request *http.Request, paramIns ...map[string]any) {
	var params = make(map[string]any)
	for _, p := range paramIns {
		if p == nil {
			continue
		}
		for k, v := range p {
			params[k] = v
		}
	}
	params["__innerHtml"] = innerHtml
	var page string
	if override {
		page = utils.InterfaceToString(innerHtml)
	} else {
		page = defaultRenderPage
	}
	unsafeTemplateRender(writer, request, page, params)
}

func block(title string, text string) string {
	raw, _ := unsafeTemplate(`<h2>{{ .title }}</h2> <br> <p class='code-style'>{{ .text }}</p> <br><br>`, map[string]any{
		"title": title, "text": text,
	})
	return string(raw)
}

func BlockContent(i ...string) string {
	return strings.Join(i, "<br>")
}
