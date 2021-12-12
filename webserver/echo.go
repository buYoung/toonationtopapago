package webserver

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"livteam/toonationpapago/toonation"
	"log"
	"net/http"
	"strings"
)

var (
	upgrader = websocket.Upgrader{}
)

func getFileSystem(fsd embed.FS) http.FileSystem {
	fsys, err := fs.Sub(fsd, "template/assets")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

func Start(fs embed.FS) {
	e := echo.New()
	client := resty.New()
	req := client.R()
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	assetHandler := http.FileServer(getFileSystem(fs))
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseFS(fs, "template/*.html")),
	}
	e.GET("/ui/index.html", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{})
	})

	e.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/assets/", assetHandler)))
	e.POST("/check", func(c echo.Context) error {
		m := map[string]string{}
		err := c.Bind(&m)
		if err != nil {
			return c.String(http.StatusOK, "error:"+err.Error())
		}
		get, err := req.Get(fmt.Sprintf("https://toon.at/widget/alertbox/%s", m["key"]))
		if err != nil {
			return c.String(http.StatusOK, "error1:"+err.Error())
		}
		if !strings.Contains(get.String(), "widget_malformed_url_desc") {
			toonation.UserKey = &toonation.ToonationKey{
				WigetUrl: fmt.Sprintf("https://toon.at/widget/alertbox/%s", m["key"]),
			}
			data, err := json.Marshal(toonation.UserKey)
			if err != nil {
				return c.String(http.StatusOK, "fail2")
			}
			err = ioutil.WriteFile("user.json", data, 0644)
			if err != nil {
				return c.String(http.StatusOK, "fail3")
			}
			toonation.Papago.Load("https://papago.naver.com/")
			return c.String(http.StatusOK, "success")
		}
		return c.String(http.StatusOK, "fail")
	})
	e.GET("/check", func(c echo.Context) error {
		if toonation.UserKey != nil {
			return c.String(http.StatusOK, "true")
		} else {
			return c.String(http.StatusOK, "false")
		}
	})

	go func() {
		err := e.Start(":5200")
		if err != nil {
			log.Println("Webserver Start Error", err)
		}
	}()
}

type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
