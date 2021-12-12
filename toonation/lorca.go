package toonation

import (
	"github.com/zserge/lorca"
	"livteam/toonationpapago/util"
	"log"
	"net/url"
	"os"
)

var (
	Papago     lorca.UI
	PapagoList lorca.UI
)

func PapagoWindow() {
	var err error
	if UserKey == nil {
		Papago, err = lorca.New("http://localhost:5200/ui/index.html", "", 1280, 960)
		if Papago == nil {
			log.Println("papago Window Create Error")
			util.Log_Error.Sugar.Errorf("papago Window Create Error : %v", err)
			return
		}
	} else {
		Papago, err = lorca.New("https://papago.naver.com/", "", 1280, 960)
		if Papago == nil {
			log.Println("papago Window Create Error")
			util.Log_Error.Sugar.Errorf("papago Window Create Error : %v", err)
			return
		}
	}

	//papagoListWindow()
	go func() {
		<-Papago.Done()
		os.Exit(0)
	}()
}
func papagoListWindow() {
	PapagoList, _ = lorca.New("", "", 800, 600)
	if Papago == nil {
		log.Println("papago Window Create Error")
		util.Log_Error.Sugar.Errorf("papago Window Create Error : %v")
		return
	}

	PapagoList.Load("data:text/html," + url.PathEscape(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Title</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            height: 100%;
        }
        .container {
            width: 800px;
            height: 100%;
            border: 1px solid black;
        }
        .sponsorlist {
            display: flex;
        }
        .sponsortext {
            width: max-content;
            margin-right: 10px;
        }
        .remove {
            margin: auto 0 0 auto;
        }
    </style>
</head>
<body>
<div class="container">

</div>
</body>
<script>
    const conatiner = document.querySelector(".container");
</script>
</html>
	`))
	PapagoList.Eval("conatiner")
}
