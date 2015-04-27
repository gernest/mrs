package mrs

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

func cleanUp() {
	os.RemoveAll("db")
}

func TestProfile_Create(t *testing.T) {
	ids := []string{
		"db0668ac-7eba-40dd-56ee-0b1c0b9b415d",
		"e6917dfe-b4f6-49b8-5628-83dd2a430e9a",
		"bc5288cf-4120-4f3c-5957-b19e093a12f4",
	}
	defer cleanUp()
	for _, id := range ids {
		profile := NewProfile(id)
		err := profile.Create()
		if err != nil {
			t.Error(err)
		}
	}
}
func TestProfile_Get(t *testing.T) {
	ids := []string{
		"db0668ac-7eba-40dd-56ee-0b1c0b9b415d",
		"e6917dfe-b4f6-49b8-5628-83dd2a430e9a",
		"bc5288cf-4120-4f3c-5957-b19e093a12f4",
	}
	defer cleanUp()
	for _, id := range ids {
		profile := NewProfile(id)
		err := profile.Create()
		if err != nil {
			t.Error(err)
		}
		gP, err := NewProfile(id).Get()
		if err != nil {
			t.Error(err)
		}
		if gP.ID != profile.ID {
			t.Errorf("Expected %s to Equal %s", gP.ID, profile.ID)
		}
	}
}

func TestProfile_Update(t *testing.T) {
	ids := []string{
		"db0668ac-7eba-40dd-56ee-0b1c0b9b415d",
		"e6917dfe-b4f6-49b8-5628-83dd2a430e9a",
		"bc5288cf-4120-4f3c-5957-b19e093a12f4",
	}
	defer cleanUp()
	for _, id := range ids {
		profile := NewProfile(id)
		err := profile.Create()
		if err != nil {
			t.Error(err)
		}
		gP, err := NewProfile(id).Get()
		if err != nil {
			t.Error(err)
		}
		if gP.ID != profile.ID {
			t.Errorf("Expected %s to Equal %s", gP.ID, profile.ID)
		}
		gP.City = "mwanza"
		err = gP.Update()
		if err != nil {
			t.Error(err)
		}
		up, err := profile.Get()
		if err != nil {
			t.Error(err)
		}
		if up.City != gP.City {
			t.Errorf("Expected %s actual %s", gP.City, up.City)
		}
	}
}

func TestProfile_Delete(t *testing.T) {
	ids := []string{
		"db0668ac-7eba-40dd-56ee-0b1c0b9b415d",
		"e6917dfe-b4f6-49b8-5628-83dd2a430e9a",
		"bc5288cf-4120-4f3c-5957-b19e093a12f4",
	}
	defer cleanUp()
	for _, id := range ids {
		profile := NewProfile(id)
		err := profile.Create()
		if err != nil {
			t.Error(err)
		}
		gP, err := NewProfile(id).Get()
		if err != nil {
			t.Error(err)
		}
		if gP.ID != profile.ID {
			t.Errorf("Expected %s to Equal %s", gP.ID, profile.ID)
		}
		gP.City = "mwanza"
		err = gP.Update()
		if err != nil {
			t.Error(err)
		}
		up, err := profile.Get()
		if err != nil {
			t.Error(err)
		}
		if up.City != gP.City {
			t.Errorf("Expected %s actual %s", gP.City, up.City)
		}
	}
	for _, id := range ids {
		profile := NewProfile(id)
		err := profile.Deleta()
		if err != nil {
			t.Error(err)
		}
		gP, err := profile.Get()
		if err == nil {
			t.Errorf("Expected an error got nil instead")
		}
		if gP != nil {
			t.Errorf("Expected nil got %v", gP)
		}
	}
}

func TestPhotoManager_GetSingleFile(t *testing.T) {
	pm := NewPhotoManager("media.db", "meta", "data")
	req, err := requestWithFile()
	if err != nil {
		t.Error(err)
	}
	up, err := pm.GetSingleFileUpload(req, "profile")
	if err != nil {
		t.Error(err)
	}
	if up == nil {
		t.Error("Expected FileUpload, got nil instead")
	}
	if up != nil {
		if up.Ext != "jpg" {
			t.Errorf("Expected jpg actual %s", up.Ext)
		}
	}
}

func TestPhotoManager_GetAllFiles(t *testing.T) {
	pm := NewPhotoManager("media.db", "meta", "data")
	req, err := requestMuliFile()
	if err != nil {
		t.Error(err)
	}
	up, err := pm.GetUploadFiles(req, "photos")
	if err != nil {
		t.Error(err)
	}
	if up == nil {
		t.Error("Expected FileUpload, got nil instead")
	}
	if up != nil {
		if len(up) != 3 {
			t.Errorf("Expected 3 actual %d", len(up))
		}
	}
}

func TestPhotoManager_SaveSingleFile(t *testing.T) {
	os.MkdirAll("db", 0700)
	profileID := "db0668ac-7eba-40dd-56ee-0b1c0b9b415d"
	pm := NewPhotoManager("db/media.db", "meta", "data")
	defer cleanUp()
	req, err := requestWithFile()
	if err != nil {
		t.Error(err)
	}
	up, err := pm.GetSingleFileUpload(req, "profile")
	if err != nil {
		t.Error(err)
	}
	if up == nil {
		t.Error("Expected FileUpload, got nil instead")
	}
	if up != nil {
		if up.Ext != "jpg" {
			t.Errorf("Expected jpg actual %s", up.Ext)
		}
		err := pm.SaveSingle(up, profileID)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestPhotoManager_SaveMultiple(t *testing.T) {
	os.MkdirAll("db", 0700)
	profileID := "db0668ac-7eba-40dd-56ee-0b1c0b9b415d"
	pm := NewPhotoManager("db/media.db", "meta", "data")
	defer cleanUp()
	req, err := requestMuliFile()
	if err != nil {
		t.Error(err)
	}
	up, err := pm.GetUploadFiles(req, "photos")
	if err != nil {
		t.Error(err)
	}
	if up == nil {
		t.Error("Expected FileUpload, got nil instead")
	}
	if up != nil {
		if len(up) != 3 {
			t.Errorf("Expected 3 actual %d", len(up))
		}
		err := pm.SaveMultiple(up, profileID)
		if err != nil {
			t.Error(err)
		}
	}
}
func requestWithFile() (*http.Request, error) {
	buf := new(bytes.Buffer)
	f, err := ioutil.ReadFile("me.jpg")
	if err != nil {
		return nil, err
	}
	w := multipart.NewWriter(buf)
	defer w.Close()
	ww, err := w.CreateFormFile("profile", "me.jpg")
	if err != nil {
		return nil, err
	}
	ww.Write(f)
	req, err := http.NewRequest("POST", "http://bogus.com", buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}

func requestMuliFile() (*http.Request, error) {
	buf := new(bytes.Buffer)
	f, err := ioutil.ReadFile("me.jpg")
	if err != nil {
		return nil, err
	}
	w := multipart.NewWriter(buf)
	defer w.Close()
	first, err := w.CreateFormFile("photos", "home.jpg")
	if err != nil {
		return nil, err
	}
	first.Write(f)
	second, err := w.CreateFormFile("photos", "baby.jpg")
	if err != nil {
		return nil, err
	}
	second.Write(f)
	third, err := w.CreateFormFile("photos", "wanker.jpg")
	if err != nil {
		return nil, err
	}
	third.Write(f)
	req, err := http.NewRequest("POST", "http://bogus.com", buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}
