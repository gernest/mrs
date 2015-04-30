package mrs

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

var (
	// ErrProfileNotFound is the message when the user profile is not found
	ErrProfileNotFound = errors.New("sorry: the requested profile cannot be found")
)

// Handlers user profile centric handlers
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

// ProfilePic hadles fileupload for a profile picture. This is inteded to work in
// ajax only requests.
func (h *Handlers) ProfilePic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid := vars["id"]
	if r.Method == "POST" {
		p, err := NewProfile(pid).Get()
		if h.isAjax(r) {
			if err != nil {
				h.rendr.JSON(w, http.StatusOK, &jsonErr{Msg: ErrProfileNotFound.Error()})
				return
			}
			if h.isUpload(r) {
				up, err := h.pm.GetSingleFileUpload(r, "profile")
				if err != nil {
					h.rendr.JSON(w, http.StatusOK, &jsonErr{Msg: "trouble saving"})
					return
				}
				pic, err := h.pm.SaveSingle(up, p.ID)
				if err != nil {
					h.rendr.JSON(w, http.StatusNotFound, &jsonErr{Msg: "trouble saving"})
					return
				}
				p.Picture = pic.ID
				err = p.Update()
				if err != nil {
					// TODO (gernest): log this error
				}
				h.rendr.JSON(w, http.StatusOK, p)
				return
			}

		}
	}
}

// FileUploads handlers multiple file uploads by a given user.
func (h *Handlers) FileUploads(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid := vars["id"]
	if r.Method == "POST" {
		p, err := NewProfile(pid).Get()
		if h.isAjax(r) {
			if err != nil {
				h.rendr.JSON(w, http.StatusOK, &jsonErr{Msg: ErrProfileNotFound.Error()})
				return
			}
			if h.isUpload(r) {
				up, err := h.pm.GetUploadFiles(r, "photos")
				if err != nil {
					h.rendr.JSON(w, http.StatusOK, &jsonErr{Msg: "trouble saving"})
					return
				}
				ups, err := h.pm.SaveMultiple(up, p.ID)
				if err != nil {
					h.rendr.JSON(w, http.StatusOK, &jsonErr{Msg: "trouble saving"})
					return
				}
				h.rendr.JSON(w, http.StatusOK, ups)
				return
			}

		}
	}
}
func (h *Handlers) isAjax(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

func (h *Handlers) isUpload(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data")
}
