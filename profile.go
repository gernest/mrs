package mrs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/gernest/nutz"
	u "github.com/nu7hatch/gouuid"
)

const defaultMaxMemory = 32 << 20 //32MB

type Profile struct {
	store     nutz.Storage `json:"-"`
	ID        string       `json:"id"`
	Picture   string       `json:"picture"`
	Age       int          `json:"age"`
	BirthDate time.Time    `json:"birth_date"`
	Height    int          `json:"height"`
	Weight    int          `json:"weight"`
	Hobies    []string     `json:"hobies"`
	Photos    []string     `json:"photos"`
	City      string       `json:"city"`
	Country   string       `json:"country"`
	Street    string       `json:"street"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"update_at"`
}

type Photo struct {
	ID         string    `json:!id"`
	Type       string    `json:"type"`
	Size       int       `json:"size"`
	UploadedBy string    `json:"uploaded_by"`
	UploadedAt time.Time `json:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PhotoManager struct {
	store      nutz.Storage
	MetaBucket string
	DataBucket string
}

type FileUpload struct {
	Body *multipart.File
	Ext  string
}

func NewProfile(userID string) *Profile {
	p := new(Profile)
	p.store = nutz.NewStorage("db/"+userID+".db", 0600, nil)
	p.ID = userID
	err := os.MkdirAll("db", 0700)
	if err != nil {
		log.Println(err)
	}
	return p
}

func (p *Profile) Create() error {
	p.CreatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	r := p.store.Create(p.ID, p.ID, data)
	return r.Error
}

func (p *Profile) Get() (*Profile, error) {
	s := p.store.Get(p.ID, p.ID)
	if s.Error != nil {
		return nil, s.Error
	}
	err := json.Unmarshal(s.Data, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Profile) Update() error {
	p.UpdatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	r := p.store.Update(p.ID, p.ID, data)
	return r.Error
}

func (p *Profile) Deleta() error {
	d := p.store.Delete(p.ID, p.ID)
	return d.Error
}

func NewPhotoManager(db, meta, data string) *PhotoManager {
	return &PhotoManager{
		store:      nutz.NewStorage(db, 0600, nil),
		MetaBucket: meta,
		DataBucket: data,
	}
}

func (p *PhotoManager) NewPhoto(profileID string) *Photo {
	uuid, err := u.NewV4()
	if err != nil {
		log.Panicln(err)
	}
	return &Photo{ID: uuid.String(), UploadedBy: profileID}
}

func (p *PhotoManager) GetAllUploadeFiles(r *http.Request, fieldName string) ([]*FileUpload, error) {
	err := r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		return nil, err
	}
	if up := r.MultipartForm.File[fieldName]; len(up) > 0 {
		var rst []*FileUpload
		var ferr error
		for _, v := range up {
			f, ferr := v.Open()
			if ferr != nil {
				continue
			}
			file, ferr := p.getFileUpload(f)
			if ferr != nil {
				log.Println(ferr)
				continue
			}
			rst = append(rst, file)
		}
		if len(rst) > 0 {
			return rst, nil
		}
		if ferr != nil {
			return nil, ferr
		}
	}
	return nil, http.ErrMissingFile
}
func (p *PhotoManager) GetSingleFileUpload(r *http.Request, fieldName string) (*FileUpload, error) {
	file, _, err := r.FormFile(fieldName)
	if err != nil {
		return nil, err
	}
	return p.getFileUpload(file)
}
func (p *PhotoManager) getFileUpload(file multipart.File) (*FileUpload, error) {
	ext, err := p.getFileExt(file)
	if err != nil {
		return nil, err
	}
	return &FileUpload{&file, ext}, nil
}
func (p *PhotoManager) getFileExt(file multipart.File) (string, error) {
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	defer file.Seek(0, 0)
	if err != nil {
		return "", err
	}
	f := http.DetectContentType(buf)
	switch f {
	case "image/jpeg", "image/jpg":
		return "jpg", nil
	case "image/png":
		return "png", nil
	default:
		return "", fmt.Errorf("file %s not supported", f)
	}

}

func (p *PhotoManager) SaveMultiplePhotos(files []*FileUpload, profileID string) error {
	for _, v := range files {
		err := p.SaveSingleFile(v, profileID)
		if err != nil {
			return err
		}
	}
	return nil
}
func (p *PhotoManager) SaveSingleFile(file *FileUpload, profileID string) error {
	photo := p.NewPhoto(profileID)
	photo.Type = file.Ext
	data, err := p.encodePhoto(file)
	if err != nil {
		return err
	}
	photo.Size = len(data)
	photo.UploadedAt = time.Now()
	photo.UpdatedAt = time.Now()

	meta, err := json.Marshal(photo)
	if err != nil {
		return err
	}

	s := p.store.Create(p.MetaBucket, photo.ID, meta)
	if s.Error != nil {
		return s.Error
	}
	s = p.store.Create(p.DataBucket, photo.ID, data)
	if s.Error != nil {
		return s.Error
	}
	return nil
}

func (p *PhotoManager) encodePhoto(file *FileUpload) ([]byte, error) {
	ext := file.Ext
	switch ext {
	case "jpg", "jpeg":
		img, err := jpeg.Decode(*file.Body)
		if err != nil {
			return nil, err
		}
		opts := jpeg.Options{98}
		buf := new(bytes.Buffer)
		jpeg.Encode(buf, img, &opts)
		return buf.Bytes(), nil
	case "png", "PNG":
		img, err := png.Decode(*file.Body)
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		png.Encode(buf, img)
		return buf.Bytes(), nil
	}
	return nil, errors.New("mrs: file not supported")
}
