package mrs

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

var (
	ErrProfileNotFound = errors.New("sorry: the requested profile cannot be found")
)

type Handlers struct {
	pm    *PhotoManager
	rendr *render.Render
}

type jsonErr struct {
	Msg string `json:"msg"`
}

// NewHandlers initialize a new Handlers instance.
func NewHandlers(db, meta, data string, opt *render.Options) *Handlers {
	r := render.New()
	if opt != nil {
		r = render.New(*opt)
	}
	return &Handlers{pm: NewPhotoManager(db, meta, data), rendr: r}
}

// Home handles the profile home page. It expects in the url path to have the param
// id which is a uuid v4 string.using gorilla mux the url  should be as follows.
//	/profile/{id:^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$}
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid := vars["id"]
	if r.Method == "GET" {
		p, err := NewProfile(pid).Get()
		if h.isAjax(r) {
			if err != nil {
				h.rendr.JSON(w, http.StatusNotFound, &jsonErr{Msg: ErrProfileNotFound.Error()})
				return
			}
			h.rendr.JSON(w, http.StatusOK, p)
			return
		}
		data := make(map[string]interface{})
		if err != nil {
			data["error"] = ErrProfileNotFound
			h.rendr.HTML(w, http.StatusNotFound, "404", data)
			return
		}
		data["profile"] = p
		h.rendr.HTML(w, http.StatusOK, "profile_home", data)
		return
	}

}

func (h *Handlers) isAjax(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}
