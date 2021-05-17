package profile

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/controller"
	"fractapp-server/controller/middleware"
	"fractapp-server/db"
	"fractapp-server/types"
	"fractapp-server/utils"
	"fractapp-server/validators"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	UpdateProfileRoute    = "/updateProfile"
	UsernameRoute         = "/username"
	UploadAvatarRoute     = "/uploadAvatar"
	MyProfileRoute        = "/my"
	SearchRoute           = "/search"
	MyContactsRoute       = "/contacts"
	UploadContactsRoute   = "/uploadContacts"
	MyMatchContactsRoute  = "/matchContacts"
	UserInfoRoute         = "/userInfo"
	AvatarRoute           = "/avatar"
	TransactionsRoute     = "/transactions"
	SubstrateBalanceRoute = "/substrate/balance"

	AvatarDir       = "/.avatars"
	MaxAvatarSize   = 1 << 20
	MaxUsersResult  = 10
	MinSearchLength = 4

	MaxContacts = 400
)

var (
	InvalidFileFormatErr      = errors.New("invalid file format")
	InvalidFileSizeErr        = errors.New("invalid file size")
	UsernameIsExistErr        = errors.New("username is exist")
	UsernameNotFoundErr       = errors.New("username not found")
	InvalidPropertyErr        = errors.New("property has invalid symbols or length")
	AvatarNotFoundErr         = errors.New("avatar not found")
	InvalidConnectionTxApiErr = errors.New("invalid connection to transaction API")
)

type Controller struct {
	db        db.DB
	txApiHost string
}

func NewController(db db.DB, txApiHost string) *Controller {
	return &Controller{
		db:        db,
		txApiHost: txApiHost,
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
		return c.uploadMyContacts, nil
	case UserInfoRoute:
		return c.userInfo, nil
	case AvatarRoute:
		return c.avatar, nil
	case TransactionsRoute:
		return c.transactions, nil
	case SubstrateBalanceRoute:
		return c.substrateBalance, nil
	}

	return nil, controller.InvalidRouteErr
}
func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
	switch err {
	case db.ErrNoRows:
		http.Error(w, "", http.StatusNotFound)
	case InvalidFileFormatErr:
		fallthrough
	case InvalidFileSizeErr:
		fallthrough
	case UsernameIsExistErr:
		fallthrough
	case InvalidPropertyErr:
		http.Error(w, err.Error(), http.StatusBadRequest)
	case UsernameNotFoundErr:
		http.Error(w, err.Error(), http.StatusNotFound)
	case AvatarNotFoundErr:
		path, err := os.Getwd()
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		b, err := ioutil.ReadFile(path + "/assets/default-avatar.png")
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Write(b)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

// search godoc
// @Summary Search user
// @Description search user by email or username
// @ID search
// @Tags Profile
// @Accept  json
// @Produce json
// @Param value query string true "username or email value"
// @Param type query string false "email/username"
// @Success 200 {object} []ShortUserProfile
// @Failure 400 {string} string
// @Failure 404
// @Router /profile/search [get]
func (c *Controller) search(w http.ResponseWriter, r *http.Request) error {
	value := strings.Trim(strings.ToLower(r.URL.Query().Get("value")), " ")

	users := make([]ShortUserProfile, 0)
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

	profile, err := c.db.SearchUsersByEmail(value)
	if err != nil && err != db.ErrNoRows {
		return err
	}

	if profile != nil {
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

		user := ShortUserProfile{
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

// userInfo godoc
// @Summary Get user
// @Description get user by id
// @ID profileInfo
// @Tags Profile
// @Accept  json
// @Produce json
// @Param id query string true "get user profile by user id"
// @Success 200 {object} ShortUserProfile
// @Failure 400 {string} string
// @Router /profile/userInfo [get]
func (c *Controller) userInfo(w http.ResponseWriter, r *http.Request) error {
	id := strings.Trim(r.URL.Query().Get("id"), " ")

	var p *db.Profile
	var err error
	if id != "" {
		p, err = c.db.ProfileById(id)
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

	user := ShortUserProfile{
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

// myProfile godoc
// @Summary Get my profile
// @Security AuthWithJWT
// @ID myProfile
// @Tags Profile
// @Accept  json
// @Produce json
// @Success 200 {object} MyProfile
// @Failure 400
// @Router /profile/my [get]
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

// updateProfile godoc
// @Summary Update my profile
// @Security AuthWithJWT
// @ID updateProfile
// @Tags Profile
// @Accept  json
// @Produce json
// @Param rq body UpdateProfileRq true "update profile model"
// @Success 200
// @Failure 400 {string} string
// @Router /profile/updateProfile [post]
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
	}

	if profile.Name != rq.Name {
		if !validators.IsValidName(rq.Name) {
			return InvalidPropertyErr
		}
		profile.Name = rq.Name
	}

	profile.LastUpdate = sec

	err = c.db.UpdateByPK(profile)
	if err != nil {
		return err
	}

	return nil
}

// avatar godoc
// @Summary Get user avatar
// @ID avatar
// @Tags Profile
// @Accept  json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200
// @Failure 400 {string} string
// @Router /profile/avatar/{userId} [get]
func (c *Controller) avatar(w http.ResponseWriter, r *http.Request) error {
	u, _ := url.Parse(r.URL.Path)
	userId := path.Base(u.Path)

	var p, err = c.db.ProfileById(userId)
	if err != nil {
		return AvatarNotFoundErr
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(path + AvatarDir + "/" + userId + "." + p.AvatarExt)
	if err != nil {
		return AvatarNotFoundErr
	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// uploadAvatar godoc
// @Summary Update avatar
// @Security AuthWithJWT
// @ID uploadAvatar
// @Tags Profile
// @Accept x-www-form-urlencoded
// @Produce json
// @Param format formData string true "image/jpeg or image/jpg or image/png"
// @Param avatar formData string true "avatar in base64 (https://onlinepngtools.com/convert-png-to-base64)"
// @Success 200
// @Failure 400 {string} string
// @Router /profile/uploadAvatar [post]
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

	err = utils.WriteAvatar(fileName, decoded)
	if err != nil {
		return err
	}

	now := time.Now()

	profile, err := c.db.ProfileById(id)
	if err != nil {
		return err
	}

	profile.AvatarExt = ex[1]
	profile.LastUpdate = now.Unix()
	err = c.db.UpdateByPK(profile)
	if err != nil {
		return err
	}

	return nil
}

// myContacts godoc
// @Summary Get my contacts
// @Security AuthWithJWT
// @ID myContacts
// @Tags Profile
// @Accept  json
// @Produce json
// @Success 200 {object} []string
// @Failure 400 {string} string
// @Router /profile/contacts [get]
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

// myMatchContacts godoc
// @Summary Get my matched contacts
// @Description Only those who are in your contacts can see your profile by phone number. Your number should also be in their contacts.
// @Security AuthWithJWT
// @ID myMatchContacts
// @Tags Profile
// @Accept  json
// @Produce json
// @Success 200 {object} []string
// @Failure 400 {string} string
// @Router /profile/matchContacts [get]
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

	var users []ShortUserProfile
	for _, v := range matchContacts {
		addresses, err := c.db.AddressesById(v.Id)
		if err != nil {
			continue
		}

		user := ShortUserProfile{
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

// uploadMyContacts godoc
// @Summary Upload my phone numbers of contacts
// @Security AuthWithJWT
// @ID uploadMyContacts
// @Tags Profile
// @Accept  json
// @Produce json
// @Param rq body []string true "phone numbers of contacts"
// @Success 200
// @Failure 400 {string} string
// @Router /profile/uploadContacts [post]
func (c *Controller) uploadMyContacts(w http.ResponseWriter, r *http.Request) error {
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
		if !validators.IsValidatePhoneNumber(v) {
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

// username godoc
// @Summary Is username exist?
// @ID username
// @Tags Profile
// @Accept  json
// @Produce json
// @Param username query string true "username min length 4"
// @Success 200
// @Failure 404 {string} string
// @Failure 400 {string} string
// @Router /profile/username [get]
func (c *Controller) findUsername(w http.ResponseWriter, r *http.Request) error {
	exist, err := c.usernameIsExist(strings.ToLower(r.URL.Query().Get("username")))
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

// username godoc
// @Summary Get transactions by address
// @ID getTransactions
// @Tags Profile
// @Accept  json
// @Produce json
// @Param address query string true "address"
// @Param currency query int true "currency"
// @Success 200 {object} Transaction
// @Failure 400 {string} string
// @Router /profile/transactions [get]
func (c *Controller) transactions(w http.ResponseWriter, r *http.Request) error {
	address := r.URL.Query().Get("address")
	currencyInt, err := strconv.ParseInt(r.URL.Query().Get("currency"), 10, 32)
	if err != nil {
		return err
	}
	currency := types.Currency(currencyInt)

	resp, err := http.Get(fmt.Sprintf("%s/transactions/%s?currency=%s", c.txApiHost, address, currency.String()))
	if err != nil {
		return InvalidConnectionTxApiErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return InvalidConnectionTxApiErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	txs := make([]types.Transaction, 0)
	err = json.Unmarshal(body, &txs)
	if err != nil {
		return err
	}

	responseTxs := make([]Transaction, 0)
	for _, v := range txs {
		txTime := time.Unix(v.Timestamp/1000, 0)
		prices, err := c.db.Prices(currency.String(), txTime.
			Add(-15*time.Minute).Unix()*1000, txTime.
			Add(15*time.Minute).Unix()*1000)

		if err != nil {
			return err
		}

		price := float32(0)
		// search for a value with a minimum difference from the transaction time
		if len(prices) > 0 {
			price = prices[0].Price
			diff := v.Timestamp - prices[0].Timestamp
			for _, p := range prices {
				newDiff := v.Timestamp - p.Timestamp
				if newDiff < 0 {
					newDiff *= -1
				}

				if newDiff < diff {
					diff = newDiff
					price = p.Price
				}
			}
		}

		value, _ := new(big.Int).SetString(v.Value, 10)
		fee, _ := new(big.Int).SetString(v.Fee, 10)

		floatValue := currency.ConvertFromPlanck(value)
		floatFee := currency.ConvertFromPlanck(fee)

		a := currency.Accuracy()
		aBig := new(big.Float).SetInt(big.NewInt(a))

		fv, _ := new(big.Float).Mul(floatValue, aBig).Float32()
		fv /= float32(a)
		usdValue := price * fv

		ff, _ := new(big.Float).Mul(floatFee, aBig).Float32()
		ff /= float32(a)
		usdFee := price * ff

		userFrom := ""
		p, err := c.db.ProfileByAddress(v.From)
		if err != nil && err != db.ErrNoRows {
			return err
		}
		if p != nil {
			userFrom = p.Id
		}

		userTo := ""
		p, err = c.db.ProfileByAddress(v.To)
		if err != nil && err != db.ErrNoRows {
			return err
		}
		if p != nil {
			userTo = p.Id
		}

		responseTxs = append(responseTxs, Transaction{
			ID:   v.ID,
			Hash: v.Hash,

			Currency: v.Currency,

			From:     v.From,
			UserFrom: userFrom,

			To:     v.To,
			UserTo: userTo,

			Value:      v.Value,
			UsdValue:   usdValue,
			FloatValue: floatValue.String(),

			Fee:       v.Fee,
			UsdFee:    usdFee,
			FloatFee:  floatFee.String(),
			Timestamp: v.Timestamp,
			Status:    v.Status,
		})
	}

	rsByte, err := json.Marshal(responseTxs)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}
	return nil
}

// username godoc
// @Summary Get substrateBalance by address
// @ID getBalance
// @Tags Profile
// @Accept  json
// @Produce json
// @Param address query string true "address"
// @Param currency query int true "currency"
// @Success 200 {object} Balance
// @Failure 400 {string} string
// @Router /profile/substrateBalance [get]
func (c *Controller) substrateBalance(w http.ResponseWriter, r *http.Request) error {
	address := r.URL.Query().Get("address")
	currencyInt, err := strconv.ParseInt(r.URL.Query().Get("currency"), 10, 32)
	if err != nil {
		return err
	}
	currency := types.Currency(currencyInt)

	resp, err := http.Get(fmt.Sprintf("%s/substrate/balance/%s?currency=%s", c.txApiHost, address, currency.String()))
	if err != nil {
		return InvalidConnectionTxApiErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return InvalidConnectionTxApiErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	balance := new(Balance)
	err = json.Unmarshal(body, &balance)
	if err != nil {
		return err
	}

	rsByte, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}
	return nil
}
