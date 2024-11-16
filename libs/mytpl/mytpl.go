package mytpl;

import (
    "radio_site/libs/myerr"

    "html/template"
    "log"
)

var Tpl *template.Template

func Template_init() {
    var err error

    Tpl, err = Tpl.ParseGlob("./templates/*.html")

    myerr.Check_err(err)

    log.Println("Parsed templates:")
    for _, tmpl := range Tpl.Templates() {
        log.Println(" - ", tmpl.Name())
    }
}
