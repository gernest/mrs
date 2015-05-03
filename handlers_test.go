package mrs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gernest/render"
	"github.com/gorilla/mux"
)

var (
	pids = []string{
		"db0668ac-7eba-40dd-56ee-0b1c0b9b415d",
		"e6917dfe-b4f6-49b8-5628-83dd2a430e9a",
		"bc5288cf-4120-4f3c-5957-b19e093a12f4",
	}
)

func TestHandlers_Home(t *testing.T) {
	opts := render.Options{Directory: "fixture"}
	cfg:=&Config{"imgs.db", "meta", "data"}
	handle := NewHandlers(cfg, &opts)
	defer handle.pm.store.DeleteDatabase()

	h := mux.NewRouter()
	h.HandleFunc("/profile/{id}", handle.Home)

	for _, id := range pids {
		// GET request expecting htm response
		r, err := http.NewRequest("GET", fmt.Sprintf("/profile/%s", id), nil)
		if err != nil {
			t.Error(err)
		}
		if err == nil {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			if w.Code != http.StatusNotFound {
				t.Errorf("Expected %d found %d", http.StatusNotFound, w.Code)
			}
			if w.Code == http.StatusNotFound {
				if !strings.Contains(w.Body.String(), ErrProfileNotFound.Error()) {
					t.Errorf("Expected %s to contain %s", w.Body.String(), ErrProfileNotFound)
				}
			}
		}

		// GET request expecting json resonse
		rj, err := http.NewRequest("GET", fmt.Sprintf("/profile/%s", id), nil)
		if err != nil {
			t.Error(err)
		}
		rj.Header.Set("X-Requested-With", "XMLHttpRequest")
		if err == nil {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, rj)

			if w.Code != http.StatusNotFound {
				t.Errorf("Expected %d found %d", http.StatusNotFound, w.Code)
			}
			if w.Code == http.StatusNotFound {
				if !strings.Contains(w.Body.String(), ErrProfileNotFound.Error()) {
					t.Errorf("Expected %s to contain %s", w.Body.String(), ErrProfileNotFound)
				}
			}
		}

	}

	// create the test profiles
	defer cleanUp()
	for _, id := range pids {
		profile := NewProfile(id)
		err := profile.Create()
		if err != nil {
			t.Error(err)
		}
	}

	for _, id := range pids {
		// GET request expecting htm response
		r, err := http.NewRequest("GET", fmt.Sprintf("/profile/%s", id), nil)
		if err != nil {
			t.Error(err)
		}
		if err == nil {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Errorf("Expected %d found %d", http.StatusOK, w.Code)
			}
			if w.Code == http.StatusOK {
				if !strings.Contains(w.Body.String(), id) {
					t.Errorf("Expected %s to contain %s", w.Body.String(), id)
				}
			}
		}

		// GET request expecting json resonse
		rj, err := http.NewRequest("GET", fmt.Sprintf("/profile/%s", id), nil)
		if err != nil {
			t.Error(err)
		}
		rj.Header.Set("X-Requested-With", "XMLHttpRequest")
		if err == nil {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, rj)

			if w.Code != http.StatusOK {
				t.Errorf("Expected %d found %d", http.StatusOK, w.Code)
			}
			if w.Code == http.StatusOK {
				if !strings.Contains(w.Body.String(), id) {
					t.Errorf("Expected %s to contain %s", w.Body.String(), id)
				}
			}
		}

	}

}

func TestHandlers_ProfilePic(t *testing.T) {
	opts := render.Options{Directory: "fixture"}
	cfg:=&Config{"imgs.db", "meta", "data"}
	handle := NewHandlers(cfg, &opts)
	defer handle.pm.store.DeleteDatabase()

	h := mux.NewRouter()
	h.HandleFunc("/profile/picture/{id}", handle.ProfilePic)
	bPath := "/profile/picture/"

	// there is no profile yet
	req := ajaxtWithFile(fmt.Sprintf("%s%s", bPath, pids[0]), "profile", t)
	if req != nil {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected %d actual %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), ErrProfileNotFound.Error()) {
			t.Errorf("Expected %s to contain %s", w.Body.String(), ErrProfileNotFound.Error())
		}

		// Create a new profile and try again
		profile := NewProfile(pids[0])
		err := profile.Create()
		if err != nil {
			t.Error(err)
		}
		if err == nil {
			w2 := httptest.NewRecorder()
			h.ServeHTTP(w2, req)
			if w2.Code != http.StatusOK {
				t.Errorf("Expected %d actual %d", http.StatusOK, w.Code)
			}
			if !strings.Contains(w2.Body.String(), profile.ID) {
				t.Errorf("Expected %s to contain %s", w2.Body.String(), profile.ID)
			}
		}

	}
	// There is np such field name
	req2 := ajaxtWithFile(fmt.Sprintf("%s%s", bPath, pids[0]), "profile_pic", t)
	if req2 != nil {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req2)
		if w.Code != http.StatusOK {
			t.Errorf("Expected %d actual %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), "trouble saving") {
			t.Errorf("Expected %s to contain trouble saving", w.Body.String())
		}
	}

}

func TestHandlers_FileUploads(t *testing.T) {
	opts := render.Options{Directory: "fixture"}
	cfg:=&Config{"imgs.db", "meta", "data"}
	handle := NewHandlers(cfg, &opts)
	defer handle.pm.store.DeleteDatabase()

	h := mux.NewRouter()
	h.HandleFunc("/profile/uploads/{id}", handle.FileUploads)
	bPath := "/profile/uploads/"

	req := ajaxWithMultipleFiles(fmt.Sprintf("%s%s", bPath, pids[2]), "photos", t)
	if req == nil {
		t.Error("Expected 7http.Request got nil instead")
	}
	if req != nil {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected %d actual %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), ErrProfileNotFound.Error()) {
			t.Errorf("Expected %s to contain %s", w.Body.String(), ErrProfileNotFound.Error())
		}
		// Create a new profile and try again
		profile := NewProfile(pids[2])
		err := profile.Create()
		if err != nil {
			t.Error(err)
		}
		if err == nil {
			w2 := httptest.NewRecorder()
			h.ServeHTTP(w2, req)
			if w2.Code != http.StatusOK {
				t.Errorf("Expected %d actual %d", http.StatusOK, w.Code)
			}
			if !strings.Contains(w2.Body.String(), profile.ID) {
				t.Errorf("Expected %s to contain %s", w2.Body.String(), profile.ID)
			}
		}
		// There is np such field name
		req2 := ajaxtWithFile(fmt.Sprintf("%s%s", bPath, pids[2]), "profile_pic", t)
		if req2 != nil {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req2)
			if w.Code != http.StatusOK {
				t.Errorf("Expected %d actual %d", http.StatusOK, w.Code)
			}
			if !strings.Contains(w.Body.String(), "trouble saving") {
				t.Errorf("Expected %s to contain trouble saving", w.Body.String())
			}
		}
	}

}
func ajaxtWithFile(path, fname string, t *testing.T) *http.Request {
	buf := new(bytes.Buffer)
	f, err := ioutil.ReadFile("me.jpg")
	if err != nil {
		t.Error(err)
	}
	if err == nil {
		w := multipart.NewWriter(buf)
		defer w.Close()
		ww, werr := w.CreateFormFile(fname, "me.jpg")
		if werr != nil {
			t.Error(werr)
		}
		ww.Write(f)
		req, rerr := http.NewRequest("POST", path, buf)
		if rerr != nil {
			t.Error(rerr)
		}
		if rerr == nil {
			req.Header.Set("Content-Type", w.FormDataContentType())
			req.Header.Set("X-Requested-With", "XMLHttpRequest")
			return req
		}
	}
	return nil
}

func ajaxWithMultipleFiles(path, fname string, t *testing.T) *http.Request {
	buf := new(bytes.Buffer)
	f, err := ioutil.ReadFile("me.jpg")
	if err != nil {
		t.Error(err)
	}
	if err == nil {
		w := multipart.NewWriter(buf)
		defer w.Close()
		first, err := w.CreateFormFile(fname, "home.jpg")
		if err != nil {
			t.Error(err)
		}
		first.Write(f)
		second, err := w.CreateFormFile(fname, "baby.jpg")
		if err != nil {
			t.Error(err)
		}
		second.Write(f)
		third, err := w.CreateFormFile(fname, "wanker.jpg")
		if err != nil {
			t.Error(err)
		}
		third.Write(f)
		req, rerr := http.NewRequest("POST", path, buf)
		if rerr != nil {
			t.Error(rerr)
		}
		if rerr == nil {
			req.Header.Set("Content-Type", w.FormDataContentType())
			req.Header.Set("X-Requested-With", "XMLHttpRequest")
			return req
		}
	}
	return nil
}
