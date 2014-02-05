package hookworm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/codegangsta/martini-contrib/render"
)

var (
	logTimeFmt   = "2/Jan/2006:15:04:05 -0700" // "%d/%b/%Y:%H:%M:%S %z"
	testFormHTML = template.Must(template.New("test_form").Parse(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Hookworm test page</title>
    <link rel="shortcut icon" href="../favicon.ico">
    <style type="text/css">
      body { font-family: sans-serif; }
    </style>
  </head>
  <body>
    <article>
      <h1>Hookworm test page</h1>
      <pre>{{.ProgVersion}}</pre>
      <hr>
      {{if .Debug}}
      <section id="debug_links">
        <h2>debugging </h2>
        <ul>
          <li><a href="vars">vars</a></li>
          <li><a href="pprof/">pprof</a></li>
        </ul>
      </section>
      {{end}}
      <section id="github_test">
        <h2>github test</h2>
        <form name="github" action="{{.GithubPath}}" method="post">
          <textarea name="payload" cols="80" rows="20"
                    placeholder="github payload JSON here"></textarea>
          <input type="submit" value="POST" />
        </form>
      </section>
      <section id="travis_test">
        <h2>travis test</h2>
        <form name="travis" action="{{.TravisPath}}" method="post">
          <textarea name="payload" cols="80" rows="20"
                    placeholder="travis payload JSON here"></textarea>
          <input type="submit" value="POST" />
        </form>
      </section>
    </article>
  </body>
</html>
`))
	hookwormIndex = `
   oo           ___        ___       ___  __   ___
   |"     |__| |__  \ /     |  |__| |__  |__) |__
   |      |  | |___  |      |  |  | |___ |  \ |___
 --'
--------------------------------------------------
`
	hookwormFaviconBytes []byte
)

const (
	boomExplosionsJSON = `{"error":"BOOM EXPLOSIONS"}`
	ctypeText          = "text/plain; charset=utf-8"
	ctypeJSON          = "application/json; charset=utf-8"
	ctypeHTML          = "text/html; charset=utf-8"
	ctypeIcon          = "image/vnd.microsoft.icon"

	hookwormFaviconBase64 = `
AAABAAEAEBAQAAAAAAAoAQAAFgAAACgAAAAQAAAAIAAAAAEABAAAAAAAgAAAAAAAAAAAAAAAEAAAAAAA
AAB7/wAAgushAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAABEREQAAEQABEREREAAREREREREQABERERAAERAAARERAAAR
EAAAERAAABEQAAAAAAAAERAAAAAAAAAREAAAAAAAABERAAAAAAAAEREQAAAAAAARERAAAAAAAAERAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAA`
)

type testFormContext struct {
	GithubPath  string
	TravisPath  string
	ProgVersion string
	Debug       bool
}

func init() {
	hookwormFaviconBytes, _ = base64.StdEncoding.DecodeString(hookwormFaviconBase64)
}

func handleIndex() string {
	return fmt.Sprintf("%s\n%s\n", hookwormIndex, progVersion())
}

func handleTestPage(cfg *HandlerConfig, w http.ResponseWriter) (int, string) {
	status := http.StatusOK
	body := ""

	var bodyBuf bytes.Buffer

	err := testFormHTML.Execute(&bodyBuf, &testFormContext{
		GithubPath:  strings.TrimLeft(cfg.GithubPath, "/"),
		TravisPath:  strings.TrimLeft(cfg.TravisPath, "/"),
		ProgVersion: progVersion(),
		Debug:       cfg.Debug,
	})
	if err != nil {
		status = http.StatusInternalServerError
		body = fmt.Sprintf("<!DOCTYPE html><html><head></head><body><h1>%+v</h1></body></html>", err)
	} else {
		body = string(bodyBuf.Bytes())
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	return status, body
}

func handleConfig(cfg *HandlerConfig, r render.Render) {
	r.JSON(http.StatusOK, cfg)
}

func handleGithubPayload(pipeline Handler, l *hookwormLogger, r *http.Request) (int, string) {
	return handlePayload("github", pipeline, l, r)
}

func handleTravisPayload(pipeline Handler, l *hookwormLogger, r *http.Request) (int, string) {
	return handlePayload("travis", pipeline, l, r)
}

func handlePayload(which string, pipeline Handler, l *hookwormLogger, r *http.Request) (int, string) {
	status, payload, err := prepPayloadForPipeline(l, r)
	if err != nil {
		return status, payload
	}

	if pipeline == nil {
		status, payload := reportNoPipeline(l)
		return status, payload
	}

	l.Debugf("Sending %s payload down pipeline: %+v\n", which, payload)

	if which == "github" {
		_, err = pipeline.HandleGithubPayload(payload)
	} else if which == "travis" {
		_, err = pipeline.HandleTravisPayload(payload)
	}

	return handlePayloadErrors(err)
}

func prepPayloadForPipeline(l *hookwormLogger, r *http.Request) (int, string, error) {
	payload, err := extractPayload(l, r)
	if err != nil {
		l.Printf("Error extracting payload: %v\n", err)
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusBadRequest, string(errJSON), err
		}
		return http.StatusBadRequest, boomExplosionsJSON, err
	}
	return 200, payload, nil
}

func handlePayloadErrors(err error) (int, string) {
	if err != nil {
		errJSON, err := json.Marshal(err)
		if err != nil {
			return http.StatusInternalServerError, string(errJSON)
		}
		return http.StatusInternalServerError, boomExplosionsJSON
	}

	return http.StatusNoContent, ""
}

func reportNoPipeline(l *hookwormLogger) (int, string) {
	l.Debugf("No pipeline present, so doing nothing.\n")
	return http.StatusNoContent, ""
}
