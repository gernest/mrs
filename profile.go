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

// Profile contains  some basic fields for a user profile
//
// TODO (gernest): add validation.
// TODO (gernest): add a faster serialization implementation
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

// Photo stores metadata of uploaded file. Photos are kept in two version, the
// metadata part and the actual data part. They both reside in the same database
// but in different buckets, the two versions shares the same ID. The ID field must
// be a uuid v4 string.
type Photo struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Size       int       `json:"size"`
	UploadedBy string    `json:"uploaded_by"`
	UploadedAt time.Time `json:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PhotoManager helps in photo management
type PhotoManager struct {
	store      nutz.Storage
	MetaBucket string
	DataBucket string
}

// FileUpload holds data about the uploaded file
type FileUpload struct {
	Body *multipart.File
	Ext  string
}

// NewProfile creates a new profile object, using a given userID as its ownID.
// the reason behind this is, every profile resides in its own database, where as
// the database name has a signtature of db/{userID}.db
//
// I think by doing this it waill make management of profiles easy, and by the way,
// the profile data is inside the userID bucket, meaning we can store other info that
// are related to the profile in the same database( which is what I'm trying to do).
func NewProfile(userID string) *Profile {
	p := new(Profile)
	p.store = nutz.NewStorage("db/"+userID+".db", 0600, nil)
	p.ID = userID

	// The db folder must exist, so that we can be able to create our database there
	// TODO (gernest): Move this elsewhere, but meanwhile I can't think of a beeter
	// place this should be.
	err := os.MkdirAll("db", 0700)
	if err != nil {
		log.Println(err)
	}
	return p
}

// Create stores the current profile object inside the user database. The database
// name is in the form of db/{userID}.db where ueserID is a uuid v4 string.
func (p *Profile) Create() error {
	p.CreatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	r := p.store.Create(p.ID, p.ID, data)
	return r.Error
}

// Get retrieves a given profile object from the database and Unmarshall it to the
// caller. The caller object must have the ID field set. Note that, its wise to call
// this method on new Profile objects created by NewProfile.
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

// Update stores the current state of the Profile object into the database. Note that
// the Profile.ID field must be present, also, the object must have been created
// prior to calling this method.
//
// If the  Profile.ID is not found in the the database, an error is returned.
func (p *Profile) Update() error {
	p.UpdatedAt = time.Now()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	r := p.store.Update(p.ID, p.ID, data)
	return r.Error
}

// Delete removes a given profile object from the database.
// TODO (gernest): Accept Profile.ID as argument instead of assuming the underlying
// caller  has the ID field set.
func (p *Profile) Deleta() error {
	d := p.store.Delete(p.ID, p.ID)
	return d.Error
}

// NewPhotomanager initializes a PhotoManager object. The meta and data string
// represent the buckets to store metadata, and actual data about the photos respectively.
// The db is the database name to be used.
func NewPhotoManager(db, meta, data string) *PhotoManager {
	return &PhotoManager{
		store:      nutz.NewStorage(db, 0600, nil),
		MetaBucket: meta,
		DataBucket: data,
	}
}

// NewPhoto returns a new Photo object, given a profileID. The returned Photo object
// has a unique uuid v4 and the Photo.ProfileID set to profileID.
func (p *PhotoManager) NewPhoto(profileID string) *Photo {
	uuid, err := u.NewV4()
	if err != nil {
		log.Panicln(err)
	}
	return &Photo{ID: uuid.String(), UploadedBy: profileID}
}

// GetUploadedFiles extracts uploaded files from a given request. The filedName argument
// is the name of the form field which has the given files. It rerurns a slice of
// FileUpload object.
//
// Errors are ignored when iterating over the uploaded files.But the last version of the
// error encountered is recorded. Only if the slice is empty( meaning we failed to open any
// uploade file) will the error be returned( error from iteration)
//
// TODO (gernest): proper error handling. It will be better if a correct error message
// containing details about the files that failed to be extracted..
func (p *PhotoManager) GetUploadFiles(r *http.Request, fieldName string) ([]*FileUpload, error) {
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
				continue
			}
			rst = append(rst, file)
		}
		if len(rst) > 0 {
			return rst, nil
		}

		// by now its obvious we had problems opening any of the the uploaded files.
		// but I hope there is a better way to address this, where all the errors,
		// are collected in order to get a clear idea on what really went wrong.
		if ferr != nil {
			return nil, ferr
		}
	}
	return nil, http.ErrMissingFile
}

// GetSingleFileUpload retrieves a single file from the request. The fieldName argument
// is the name of the form file field.
func (p *PhotoManager) GetSingleFileUpload(r *http.Request, fieldName string) (*FileUpload, error) {
	file, _, err := r.FormFile(fieldName)
	if err != nil {
		return nil, err
	}
	return p.getFileUpload(file)
}

// TODO (gernest): Add optional parameter for a filter fuction, which will be used
// to match the supported files.
func (p *PhotoManager) getFileUpload(file multipart.File) (*FileUpload, error) {
	ext, err := p.getFileExt(file)
	if err != nil {
		return nil, err
	}
	return &FileUpload{&file, ext}, nil
}

// properly etracting the type of the uploaded file, since I only want to waork
// with images,this method will only return extention for jpeg and png format. otherwise
// it returns an empty string and probably a meaningful error.
//
//TODO (gernest): Add credit, I borrowed this from avatar-go project but I can't
// remember the project
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

// SaveMultiple stores multiple uploaded files.
func (p *PhotoManager) SaveMultiple(files []*FileUpload, profileID string) error {
	for _, v := range files {
		err := p.SaveSingle(v, profileID)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveSingle stores a given file into the database. The file is broken ito two parts
// metadata, and actual data. Metadata is a Photo object holding iformation about the file
// the data is a byte slice of an encoded image.
//
// To make retrieving the two parts easy, they are both stored in the same database
// but dirrenet buckets. The buckets used are the ones specified in the PhotoManager
// instance, where meatadata will go into the MetadaBucket attribute, and the data
// will go into the  the DataBucket attribute.
//
// All the two parts shares the same Key, which is generated with the NewPhoto method.
func (p *PhotoManager) SaveSingle(file *FileUpload, profileID string) error {
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

// handles encoding of the uploaded files into a byte slice
func (p *PhotoManager) encodePhoto(file *FileUpload) ([]byte, error) {
	ext := file.Ext
	switch ext {
	case "jpg", "jpeg":
		img, err := jpeg.Decode(*file.Body)
		if err != nil {
			return nil, err
		}

		// this is supposed to increase the quality of the image. But I'm not sure
		// yet if it is necessary or we should just put nil, which will result into
		// using default values.
		opts := jpeg.Options{Quality: 98}

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
