package mrs

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gernest/warlock"
	"github.com/gorilla/context"
	"github.com/monoculum/formam"

	"github.com/gernest/render"
	"github.com/gorilla/mux"
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
func NewHandlers(cfg *Config, opt *render.Options, args ...interface{}) *Handlers {
	p := &Handlers{pm: NewPhotoManager(cfg.DB, cfg.MetaBucket, cfg.DataBucket)}
	if len(args) > 0 {
		switch t := args[0].(type) {
		case *render.Render:
			p.rendr = t
			return p
		}
	}
	if opt != nil {
		p.rendr = render.New(*opt)
		return p
	}
	p.rendr = render.New()

	return p
}

// Home handles the profile home page. It expects in the url path to have the param
// id which is a uuid v4 string.using gorilla mux the url  should be as follows.
//	/profile/{id:^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$}
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid := vars["id"]
	if r.Method == "GET" {
		var p *Profile
		var err error
		p, err = NewProfile(pid).Get()
		if p == nil && h.isAllowed(r, pid) {
			p = NewProfile(pid)
			err = p.Create()
		}
		if h.isAjax(r) {
			if err != nil {
				h.rendr.JSON(w, http.StatusNotFound, &jsonErr{Msg: ErrProfileNotFound.Error()})
				return
			}
			h.rendr.JSON(w, http.StatusOK, p)
			return
		}
		data := render.NewTemplateData()
		data.Add("user", h.getCurrentUser(r))
		if err != nil {
			data.Add("error", ErrProfileNotFound)
			h.rendr.HTML(w, http.StatusNotFound, "404", data)
			return
		}
		data.Add("profile", p)
		h.rendr.HTML(w, http.StatusOK, "profile/home", data)
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
func (h *Handlers) View(w http.ResponseWriter, r *http.Request) {
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
		data := render.NewTemplateData()
		data.Add("user", h.getCurrentUser(r))
		if err != nil {
			data.Add("error", ErrProfileNotFound)
			h.rendr.HTML(w, http.StatusNotFound, "404", data)
			return
		}
		data.Add("profile", p)
		h.rendr.HTML(w, http.StatusOK, "profile/view", data)
		return
	}
}
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid := vars["id"]
	data := render.NewTemplateData()
	data.Add("user", h.getCurrentUser(r))
	p, err := NewProfile(pid).Get()
	if r.Method == "GET" {
		if h.isAjax(r) {
			if err != nil {
				h.rendr.JSON(w, http.StatusNotFound, &jsonErr{Msg: ErrProfileNotFound.Error()})
				return
			}
			h.rendr.JSON(w, http.StatusOK, p)
			return
		}
		if err != nil {
			data.Add("error", ErrProfileNotFound)
			h.rendr.HTML(w, http.StatusNotFound, "404", data)
			return
		}
		data.Add("profile", p)
		h.rendr.HTML(w, http.StatusOK, "profile/update", data)
		return
	}
	if r.Method == "POST" {
		r.ParseForm()
		if err := formam.Decode(r.Form, p); err != nil {
			data.Add("error", err.Error())
			data.Add("profile", p)
			h.rendr.HTML(w, http.StatusOK, "profile/update", data)
			return
		}
		err := p.Update()
		if err != nil {
			data.Add("error", err.Error())
			data.Add("profile", p)
			h.rendr.HTML(w, http.StatusOK, "profile/update", data)
			return
		}
		data.Add("success", "profile was updated")
		data.Add("profile", p)
		h.rendr.HTML(w, http.StatusOK, "profile/update", data)
		return
	}
}
func (h *Handlers) isAjax(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

func (h *Handlers) isUpload(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data")
}

func (h *Handlers) isAllowed(r *http.Request, profileID string) bool {
	u := h.getCurrentUser(r)
	if u != nil {
		if u.ID == profileID {
			return true
		}
	}
	return false
}

func (h *Handlers) getCurrentUser(r *http.Request) *warlock.User {
	u := context.Get(r, "user")
	if u != nil {
		usr := u.(*warlock.User)
		return usr
	}
	return nil
}

func (h *Handlers) MeOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pid := vars["id"]
		if h.isAllowed(r, pid) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "you have no permission to acess this page", http.StatusForbidden)
			return
		}
	}
}
