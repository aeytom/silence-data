package silence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	apiUrl  = "https://api.connectivity.silence.eco/api/v1/"
	version = "2.1.0"

	debugHttp = false
)

type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Version  string `json:"version,omitempty"`
}

type RefreshTokenRequest struct {
	Token   string `json:"token,omitempty"`
	Version string `json:"version,omitempty"`
}

type LoginResponse struct {
	Kind         string `json:"kind,omitempty"`
	LocalId      string `json:"localId,omitempty"`
	Email        string `json:"email,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	IdToken      string `json:"idToken,omitempty"`
	Registered   bool   `json:"registered,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    string `json:"expiresIn,omitempty"`
}

type RefreshTokenResp struct {
	AccessToken  string `json:"access_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	UserId       string `json:"user_id,omitempty"`
	ProjectId    string `json:"project_id,omitempty"`
	IdToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    string `json:"expires_in,omitempty"`
}

type ProfileResponse struct {
	Id            string `json:"id,omitempty"`
	Nif           string `json:"nif,omitempty"`
	Name          string `json:"name,omitempty"`
	LastName      string `json:"lastName,omitempty"`
	Birthday      string `json:"birthday,omitempty"`
	Gender        string `json:"gender,omitempty"`
	Country       string `json:"country,omitempty"`
	City          string `json:"city,omitempty"`
	Address       string `json:"address,omitempty"`
	PostalCode    string `json:"postalCode,omitempty"`
	Telephone     string `json:"telephone,omitempty"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"emailVerified,omitempty"`
	StripeUid     string `json:"stripeUid,omitempty"`
}

type ScooterResp struct {
	Id              string `json:"id,omitempty"`
	Model           string `json:"model,omitempty"`
	Revision        string `json:"revision,omitempty"`
	Color           string `json:"color,omitempty"`
	Name            string `json:"name,omitempty"`
	SharedToMe      bool   `json:"shared_to_me,omitempty"`
	BatteryOut      bool   `json:"battery_out,omitempty"`
	AlarmActivated  bool   `json:"alarmActivated,omitempty"`
	Charging        bool   `json:"charging,omitempty"`
	FriendSharing   string `json:"friendSharing,omitempty"`
	Imei            string `json:"imei,omitempty"`
	BtMac           string `json:"btMac,omitempty"`
	FrameNo         string `json:"frame_no,omitempty"`
	Plate           string `json:"plate,omitempty"`
	ManufactureDate string `json:"manufacture_date,omitempty"`
	TrackingDevice  struct {
		FirmwareVersion string `json:"firmwareVersion,omitempty"`
		Model           string `json:"model,omitempty"`
		Timestamp       string `json:"timestamp,omitempty"`
	} `json:"trackingDevice,omitempty"`
	LastLocation struct {
		Latitude     float64 `json:"latitude,omitempty"`
		Longitude    float64 `json:"longitude,omitempty"`
		Altitude     int32   `json:"altitude,omitempty"`
		CurrentSpeed int32   `json:"currentSpeed,omitempty"`
		Time         string  `json:"time,omitempty"`
	} `json:"lastLocation"`
	BatteryId           int64  `json:"batteryId,omitempty"`
	BatterySoc          int16  `json:"batterySoc"`
	Odometer            int32  `json:"odometer"`
	BatteryTemperature  int16  `json:"batteryTemperature"`
	MotorTemperature    int16  `json:"motorTemperature"`
	InverterTemperature int16  `json:"inverterTemperature"`
	Range               int16  `json:"range"`
	Velocity            int16  `json:"velocity"`
	Status              int16  `json:"status,omitempty"`
	LastReportTime      string `json:"lastReport_time,omitempty"`
	LastConnection      string `json:"lastConnection,omitempty"`
}

type Trip struct {
	Id              string  `json:"id,omitempty"`
	StartDate       string  `json:"startDate,omitempty"`
	EndDate         string  `json:"endDate,omitempty"`
	StartBattery    int16   `json:"startBattery,omitempty"`
	EndBattery      int16   `json:"endBattery,omitempty"`
	Distance        int32   `json:"distance,omitempty"`
	SpeedMax        float32 `json:"speedMax,omitempty"`
	SpeedAvg        float32 `json:"speedAvg,omitempty"`
	Co2Savings      float32 `json:"co2Savings,omitempty"`
	FromDescription string  `json:"fromDescription,omitempty"`
	ToDescription   string  `json:"toDescription,omitempty"`
	Points          struct {
		Lat       float32 `json:"lat,omitempty"`
		Lon       float32 `json:"lon,omitempty"`
		Timestamp string  `json:"timestamp,omitempty"`
	} `json:"points,omitempty"`
}

type TripsListResponse struct {
	Limit  int32  `json:"limit,omitempty"`
	Left   int32  `json:"left,omitempty"`
	Offset string `json:"offset,omitempty"`
	Items  []Trip `json:"items,omitempty"`
}

type Silence struct {
	auth         LoginResponse
	expiresAfter time.Time
}

func (s *Silence) addReqHeaders(req *http.Request) {

	if err := s.refreshToken(); err != nil {
		log.Fatalln(err)
	}

	if s.auth.IdToken != "" {
		req.Header.Set("authorization", "Bearer "+s.auth.IdToken)
	}

	req.Header.Set("x-userrole", "Mgmt.Customer")
	req.Header.Set("x-useragent", "APP")
	req.Header.Set("user-agent", "okhttp/4.9.2")
}

func (s *Silence) Post(path string, pd any, res any) error {

	js, err := json.Marshal(pd)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, apiUrl+path, bytes.NewBuffer(js))
	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/json")
	if resp, err := s.doHttpRequest(req); err != nil {
		return err
	} else {
		buf := new(bytes.Buffer)
		defer resp.Body.Close()
		if _, err = io.Copy(buf, resp.Body); err != nil {
			return err
		}

		if err := json.Unmarshal(buf.Bytes(), res); err != nil {
			return err
		}
		return nil
	}
}

func (s *Silence) Get(path string, res any, values url.Values) error {

	var err error
	var rurl string
	if rurl, err = url.JoinPath(apiUrl, path); err != nil {
		return err
	}

	if values != nil {
		rurl += "?" + values.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, rurl, nil)
	if err != nil {
		return err
	}

	if resp, err := s.doHttpRequest(req); err != nil {
		return err
	} else {
		// Read the token json of the response body
		buf := new(bytes.Buffer)
		defer resp.Body.Close()
		if _, err = io.Copy(buf, resp.Body); err != nil {
			return err
		}

		if err := json.Unmarshal(buf.Bytes(), res); err != nil {
			return err
		}
		return nil
	}
}

func (s *Silence) Login(email string, password string) error {
	now := time.Now()

	if err := s.Post("login", LoginRequest{
		Email:    email,
		Password: password,
		Version:  version,
	}, &s.auth); err != nil {
		return err
	}

	expin, err := time.ParseDuration(s.auth.ExpiresIn + "s")
	if err != nil {
		return err
	}
	s.expiresAfter = now.Add(expin)

	return nil
}

// refreshToken renews authentication bearer token
func (s *Silence) refreshToken() error {

	if s.auth.RefreshToken == "" {
		return nil
	}

	if time.Now().After(s.expiresAfter) {
		now := time.Now()
		qrt := RefreshTokenRequest{
			Token:   s.auth.RefreshToken,
			Version: version,
		}
		resp := RefreshTokenResp{}
		s.auth.IdToken = ""
		s.auth.RefreshToken = ""
		if err := s.Post("refreshToken", qrt, &resp); err != nil {
			return err
		} else {
			s.auth.IdToken = resp.IdToken
			s.auth.RefreshToken = resp.RefreshToken
			s.auth.ExpiresIn = resp.ExpiresIn
			expin, err := time.ParseDuration(s.auth.ExpiresIn + "s")
			if err != nil {
				return err
			}
			s.expiresAfter = now.Add(expin)
		}

	}
	return nil
}

func (s *Silence) Me() (ProfileResponse, error) {
	var profile ProfileResponse
	if err := s.Get("me", &profile, nil); err != nil {
		return profile, err
	}
	return profile, nil
}

func (s *Silence) Details() ([]ScooterResp, error) {
	var scooters []ScooterResp
	args := url.Values{
		"details": {"true"},
		"dynamic": {"true"},
	}
	if err := s.Get("me/scooters", &scooters, args); err != nil {
		return nil, err
	}
	return scooters, nil
}

func (s *Silence) Avatar() error {
	var avatar map[string]string
	if err := s.Get("me/avatar", &avatar, nil); err != nil {
		return err
	}
	fmt.Printf("%#v", avatar)
	return nil
}

func (s *Silence) TripsList(sid string, limit int32) (TripsListResponse, error) {
	args := url.Values{
		"limit": {fmt.Sprint(limit)},
	}
	var trips TripsListResponse
	err := s.Get("scooters/"+sid+"/trips", &trips, args)
	return trips, err
}

func (s *Silence) Trip(id string, tid string) (Trip, error) {
	var trip Trip
	err := s.Get("scooters/"+id+"/trips/"+tid, &trip, nil)
	return trip, err
}

func (s *Silence) doHttpRequest(req *http.Request) (*http.Response, error) {
	s.addReqHeaders(req)
	debugHttpRequest(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	debugHttpResponse(resp)

	if resp.StatusCode == 401 {
		// retry authorisation
		if err := s.refreshToken(); err != nil {
			return resp, err
		}
		s.addReqHeaders(req)
		debugHttpRequest(req)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return resp, err
		}
		debugHttpResponse(resp)
	}
	return resp, err
}

func debugHttpRequest(req *http.Request) {
	if debugHttp {
		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("REQUEST:\n%s\n", string(reqDump))
	}
}

func debugHttpResponse(resp *http.Response) {
	if debugHttp {
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("RESPONSE:\n%s\n", string(respDump))
	}
}
