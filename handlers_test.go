package mrs

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/unrolled/render"
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
	handle := NewHandlers("imgs.db", "meta", "data", &opts)
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
