package profile

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fractapp-server/controller"
	"fractapp-server/controller/middleware"
	"fractapp-server/db"
	"fractapp-server/types"
	"fractapp-server/utils"
	"fractapp-server/validators"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	UpdateProfileRoute   = "/updateProfile"
	UsernameRoute        = "/username"
	UploadAvatarRoute    = "/uploadAvatar"
	MyProfileRoute       = "/my"
	SearchRoute          = "/search"
	MyContactsRoute      = "/contacts"
	UploadContactsRoute  = "/uploadContacts"
	MyMatchContactsRoute = "/matchContacts"
	InfoRoute            = "/info"

	AvatarDir       = "/.avatars"
	MaxAvatarSize   = 1 << 20
	MaxUsersResult  = 10
	MinSearchLength = 4

	MaxContacts = 400
)

var (
	InvalidFileFormatErr = errors.New("invalid file format")
	InvalidFileSizeErr   = errors.New("invalid file size")
	UsernameIsExistErr   = errors.New("username is exist")
	UsernameNotFoundErr  = errors.New("username not found")
	InvalidPropertyErr   = errors.New("property has invalid symbols or length")
)

type Controller struct {
	db db.DB
}

func NewController(db db.DB) *Controller {
	return &Controller{
		db: db,
	}
}

func (c *Controller) MainRoute() string {
	return "/profile"
}
func (c *Controller) Handler(route string) (func(w http.ResponseWriter, r *http.Request) error, error) {
	switch route {
	case SearchRoute:
		return c.search, nil
	case UpdateProfileRoute:
		return c.updateProfile, nil
	case UsernameRoute:
		return c.findUsername, nil
	case UploadAvatarRoute:
		return c.uploadAvatar, nil
	case MyProfileRoute:
		return c.myProfile, nil
	case MyContactsRoute:
		return c.myContacts, nil
	case MyMatchContactsRoute:
		return c.myMatchContacts, nil
	case UploadContactsRoute:
		return c.uploadMyContact, nil
	case InfoRoute:
		return c.profile, nil
	}

	return nil, controller.InvalidRouteErr
}
func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
	switch err {
	case UsernameNotFoundErr:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

func (c *Controller) search(w http.ResponseWriter, r *http.Request) error {
	value := r.URL.Query().Get("value")
	value = strings.Trim(strings.ToLower(value), " ")

	searchType := r.URL.Query().Get("type")

	var users []UserProfileShort
	if len(value) < MinSearchLength {
		b, err := json.Marshal(&users)
		if err != nil {
			return err
		}
		w.Write(b)

		return nil
	}

	var profiles []db.Profile
	var err error
	if searchType == "email" {
		profile, err := c.db.SearchUsersByEmail(value)
		if err != nil {
			return err
		}

		profiles = append(profiles, *profile)
	} else {
		profiles, err = c.db.SearchUsersByUsername(value, MaxUsersResult)
		if err != nil {
			return err
		}
	}

	for _, v := range profiles {
		addresses, err := c.db.AddressesById(v.Id)
		if err != nil {
			continue
		}

		user := UserProfileShort{
			Id:         v.Id,
			Name:       v.Name,
			Username:   v.Username,
			AvatarExt:  v.AvatarExt,
			LastUpdate: v.LastUpdate,
			Addresses:  make(map[types.Currency]string),
		}

		for _, v := range addresses {
			user.Addresses[v.Network.Currency()] = v.Address
		}

		users = append(users, user)
	}

	b, err := json.Marshal(&users)
	if err != nil {
		return err
	}

	w.Write(b)
	return nil
}
func (c *Controller) profile(w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Query().Get("id")
	id = strings.Trim(id, " ")

	address := r.URL.Query().Get("address")
	address = strings.Trim(address, " ")

	var p *db.Profile
	var err error
	if id != "" {
		p, err = c.db.ProfileById(id)
	} else if address != "" {
		p, err = c.db.ProfileByAddress(address)
	} else {
		return errors.New("invalid params")
	}
	if err != nil {
		return err
	}

	addresses, err := c.db.AddressesById(p.Id)
	if err != nil {
		return err
	}

	user := UserProfileShort{
		Id:         p.Id,
		Name:       p.Name,
		Username:   p.Username,
		AvatarExt:  p.AvatarExt,
		LastUpdate: p.LastUpdate,
		Addresses:  make(map[types.Currency]string),
	}

	for _, v := range addresses {
		user.Addresses[v.Network.Currency()] = v.Address
	}

	b, err := json.Marshal(&user)
	if err != nil {
		return err
	}

	w.Write(b)
	return nil
}
func (c *Controller) myProfile(w http.ResponseWriter, r *http.Request) error {
	id := middleware.AuthId(r)

	profile, err := c.db.ProfileById(id)
	if err != nil {
		return err
	}

	myProfile := &MyProfile{
		Id:          profile.Id,
		Name:        profile.Name,
		Username:    profile.Username,
		PhoneNumber: profile.PhoneNumber,
		Email:       profile.Email,
		IsMigratory: profile.IsMigratory,
		AvatarExt:   profile.AvatarExt,
		LastUpdate:  profile.LastUpdate,
	}
	rsByte, err := json.Marshal(myProfile)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) updateProfile(w http.ResponseWriter, r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	id := middleware.AuthId(r)

	profile, err := c.db.ProfileById(id)
	if err != nil {
		return err
	}

	rq := UpdateProfileRq{}
	err = json.Unmarshal(b, &rq)
	if err != nil {
		return err
	}

	now := time.Now()
	sec := now.Unix()
	if profile.Username != strings.ToLower(rq.Username) {
		isExist, err := c.usernameIsExist(rq.Username)
		if err != nil {
			return err
		}

		if isExist {
			return UsernameIsExistErr
		}

		profile.Username = rq.Username
		profile.LastUpdate = sec
	}

	if profile.Name != rq.Name {
		if !validators.IsValidName(rq.Name) {
			return InvalidPropertyErr
		}

		profile.Name = rq.Name
		profile.LastUpdate = sec
	}

	err = c.db.UpdateByPK(profile)
	if err != nil {
		return err
	}

	return nil
}
func (c *Controller) uploadAvatar(w http.ResponseWriter, r *http.Request) error {
	base64File := r.FormValue("avatar")

	decoded, err := base64.StdEncoding.DecodeString(base64File)
	if err != nil {
		return err
	}

	extension := r.FormValue("format")

	if extension != "image/jpeg" && extension != "image/jpg" && extension != "image/png" {
		return InvalidFileFormatErr
	}

	size := len(decoded)
	if size > MaxAvatarSize {
		return InvalidFileSizeErr
	}

	id := middleware.AuthId(r)
	log.Printf("Id: %s \n", id)
	log.Printf("File Size: %+v\n", size)
	log.Printf("MIME Header: %+v\n", extension)

	ex := strings.Split(extension, "/")
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	fileName := path + AvatarDir + "/" + id + "." + ex[1]
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	err = file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = file.Write(decoded)
	if err != nil {
		return err
	}

	now := time.Now()
	sec := now.Unix()

	profile, err := c.db.ProfileById(id)
	if err != nil {
		return err
	}

	profile.AvatarExt = ex[1]
	profile.LastUpdate = sec
	err = c.db.UpdateByPK(profile)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) myContacts(w http.ResponseWriter, r *http.Request) error {
	id := middleware.AuthId(r)

	existContacts, err := c.db.AllContacts(id)
	if err != nil {
		return err
	}

	var contacts []string
	for _, v := range existContacts {
		contacts = append(contacts, v.PhoneNumber)
	}
	rsByte, err := json.Marshal(contacts)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}

	return nil
}
func (c *Controller) myMatchContacts(w http.ResponseWriter, r *http.Request) error {
	id := middleware.AuthId(r)
	p, err := c.db.ProfileById(id)
	if err != nil {
		return err
	}

	matchContacts, err := c.db.AllMatchContacts(p.Id, p.PhoneNumber)
	if err != nil {
		return err
	}

	var users []UserProfileShort
	for _, v := range matchContacts {
		addresses, err := c.db.AddressesById(v.Id)
		if err != nil {
			continue
		}

		user := UserProfileShort{
			Id:         v.Id,
			Name:       v.Name,
			Username:   v.Username,
			AvatarExt:  v.AvatarExt,
			LastUpdate: v.LastUpdate,
			Addresses:  make(map[types.Currency]string),
		}

		for _, v := range addresses {
			user.Addresses[v.Network.Currency()] = v.Address
		}

		users = append(users, user)
	}

	rsByte, err := json.Marshal(&users)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}

	return nil
}
func (c *Controller) uploadMyContact(w http.ResponseWriter, r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	id := middleware.AuthId(r)

	var contacts []string
	err = json.Unmarshal(b, &contacts)
	if err != nil {
		return err
	}

	for len(contacts) > MaxContacts {
		contacts = contacts[0:MaxContacts]
	}

	existContacts, err := c.db.AllContacts(id)
	if err != nil && err != db.ErrNoRows {
		return err
	}

	existContactsMap := make(map[string]bool)
	for _, v := range existContacts {
		existContactsMap[v.PhoneNumber] = true
	}

	var myContacts []db.Contact
	for _, v := range contacts {
		if !utils.ValidatePhoneNumber(v) {
			continue
		}
		if _, ok := existContactsMap[v]; ok {
			continue
		}

		myContacts = append(myContacts, db.Contact{
			Id:          id,
			PhoneNumber: v,
		})
	}

	if len(myContacts) > 0 {
		err = c.db.Insert(&myContacts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) findUsername(w http.ResponseWriter, r *http.Request) error {
	exist, err := c.usernameIsExist(strings.ToLower(r.URL.Query()["username"][0]))
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	return UsernameNotFoundErr
}
func (c *Controller) usernameIsExist(username string) (bool, error) {
	if !validators.IsValidUsername(username) {
		return false, InvalidPropertyErr
	}

	isExist, err := c.db.UsernameIsExist(username)
	if err != nil {
		return false, err
	}

	return isExist, nil
}
