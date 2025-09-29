// package handlers

// import (
//     "html/template"
//     "net/http"
// )

// func IndexHandler(tmpl *template.Template) http.HandlerFunc {
//     return func(w http.ResponseWriter, r *http.Request) {
//         tmpl.ExecuteTemplate(w, "login.html", nil)
//     }
// }
package handlers

import (
  "html/template"
  "net/http"
)

func IndexHandler(tmpl *template.Template) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    _ = tmpl.ExecuteTemplate(w, "login", nil)
  }
}
