package vrm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	_ "github.com/rs/zerolog"
)

const (
	baseURL string = "https://vrmapi.victronenergy.com/v2/"

	usersURL string = "{{ .baseURL }}admin/users"

	// Auth related URLs
	loginURL       string = "{{ .baseURL }}auth/login"
	logoutURL      string = "{{ .baseURL }}auth/logout"
	loginAsDemoURL string = "{{ .baseURL }}auth/loginAsDemo"

	// User related URLs
	installationsURL      string = "{{ .baseURL }}users/{{ .UserID }}/installations"
	accessTokensListURL   string = "{{ .baseURL }}users/{{ .UserID }}/accesstokens/list"
	accessTokensCreateURL string = "{{ .baseURL }}users/{{ .UserID }}/accesstokens/create"
	accessTokensRevokeURL string = "{{ .baseURL }}users/{{ .UserID }}/accesstokens/{{ .accessTokenID }}/revoke"

	// Site related URLs
	systemOverviewURL string = "{{ .baseURL }}installations/{{ .siteID }}/system-overview"
	diagnosticsURL    string = "{{ .baseURL }}installations/{{ .siteID }}/diagnostics"
	tagsURL           string = "{{ .baseURL }}installations/{{ .siteID }}/tags"
	downloadURL       string = "{{ .baseURL }}installations/{{ .siteID }}/data-download"
	gpsDownloadURL    string = "{{ .baseURL }}installations/{{ .siteID }}/gps-download?end=<end>&start=<start>"
	statsURL          string = "{{ .baseURL }}installations/{{ .siteID }}/stats"
	widgetsURL        string = "{{ .baseURL }}installations/{{ .siteID }}/widgets/{{ .widgetID }}"

	// Widgets
	WidgetGraph               string = "Graph"
	WidgetVeBusState          string = "VeBusState"
	WidgetMPPTState           string = "MPPTState"
	WidgetBatterySummary      string = "BatterySummary"
	WidgetSolarChargerSummary string = "SolarChargerSummary"
	WidgetBMSDiagnostics      string = "BMSDiagnostics"
	WidgetHistoricData        string = "HistoricData"
	WidgetIOExtenderInOut     string = "IOExtenderInOut"
	WidgetLithiumBMS          string = "LithiumBMS"
	WidgetMotorSummary        string = "MotorSummary"
	WidgetPVInverterStatus    string = "PVInverterStatus"
	WidgetStatus              string = "Status"
	WidgetAlarm               string = "Alarm"
	WidgetGPS                 string = "GPS"
	WidgetHoursOfAC           string = "HoursOfAC"

	DemoUserID int = 22
)

type InstallationsResponse struct {
	Success bool `json:"success"`
	Records []struct {
		Name            string `json:"name"`
		SiteID          int    `json:"idSite"`
		UserID          int    `json:"idUser"`
		PVMax           int    `json:"pvMax"`
		ReportsEnabled  bool   `json:"reports_enabled"`
		AccessLevel     int    `json:"accessLevel"`
		Timezone        string `json:"timezone"`
		Owner           bool   `json:"owner"`
		Geofence        string `json:"geofence"`
		GeofenceEnabled bool   `json:"geofenceEnabled"`
		DeviceIcon      string `json:"device_icon"`
		Alarm           bool   `json:"alarm,omitempty"`
		LastTimestamp   int    `json:"last_timestamp,omitempty"`
		Tags            []struct {
			TagID int    `json:"idTag"`
			Name  string `json:"name"`
		} `json:"tags,omitempty"`
		TimezoneOffset int `json:"timezone_offset,omitempty"`
		Extended       []struct {
			DataAttributeID int             `json:"idDataAttribute"`
			Code            string          `json:"code"`
			Description     string          `json:"description"`
			FormatWithUnit  string          `json:"formatWithUnit"`
			RawValue        json.RawMessage `json:"rawValue"`
			TextValue       string          `json:"textValue"`
			FormattedValue  string          `json:"formattedValue"`
		} `json:"extended,omitempty"`
		CurrentTime string `json:"current_time,omitempty"`
	} `json:"records"`
}

func (s *vrmSession) Installations(userID int) (*InstallationsResponse, error) {
	url, err := formatURL(installationsURL, URLParams{
		"UserID": strconv.Itoa(userID),
	}, struct {
		Extended uint8 `url:"extended"`
	}{1})
	if err != nil {
		return nil, err
	}

	data := InstallationsResponse{}
	if err := s.getAndLoad(url, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

type SystemOverviewResponse struct {
	Success bool `json:"success"`
	Records struct {
		Devices []struct {
			Name                 string `json:"name"`
			ProductCode          string `json:"productCode"`
			ProductName          string `json:"productName"`
			FirmwareVersion      string `json:"firmwareVersion,omitempty"`
			LastConnection       uint   `json:"lastConnection"`
			Class                string `json:"class"`
			LoggingInterval      uint   `json:"loggingInterval"`
			LastPowerUpOrRestart uint   `json:"lastPowerUpOrRestart,omitempty"`
			Instance             uint   `json:"instance,omitempty"`
		} `json:"devices"`
		UnconfiguredDevices bool `json:"unconfigured_devices"`
	} `json:"records"`
}

func (s *vrmSession) SystemOverview(siteID int) (*SystemOverviewResponse, error) {
	url, err := formatURL(systemOverviewURL, URLParams{
		"siteID": strconv.Itoa(siteID),
	}, nil)
	if err != nil {
		return nil, err
	}

	overview := SystemOverviewResponse{}
	if err := s.getAndLoad(url, &overview); err != nil {
		return nil, err
	}

	return &overview, nil
}

type DiagnosticsResponse struct {
	Success bool `json:"success"`
	Records []struct {
		ID                      uint            `json:"id"`
		Device                  string          `json:"Device"`
		Instance                uint            `json:"instance"`
		Description             string          `json:"description"`
		SiteID                  uint            `json:"idSite"`
		Timestamp               uint            `json:"timestamp"`
		FormatWithUnit          string          `json:"formatWithUnit"`
		DBusServiceType         json.RawMessage `json:"dbusServiceType"`
		DBusPath                json.RawMessage `json:"dbusPath"`
		FormattedValue          json.RawMessage `json:"formattedValue"`
		DataAttributeID         uint            `json:"idDataAttribute"`
		DataAttributeEnumValues []struct {
			Name  string          `json:"nameEnum"`
			Value json.RawMessage `json:"valueEnum"`
		} `json:"dataAttributeEnumValues"`
	} `json:"records"`
	NumRecords uint `json:"num_records"`
}

// Retrieve all most recent logged data for a given installation
func (s *vrmSession) Diagnostics(siteID int, count uint16) (*DiagnosticsResponse, error) {
	url, err := formatURL(diagnosticsURL, URLParams{
		"siteID": strconv.Itoa(siteID),
	}, struct {
		Count uint16 `url:"count"`
	}{count})
	if err != nil {
		return nil, err
	}

	diagnostics := DiagnosticsResponse{}
	if err := s.getAndLoad(url, &diagnostics); err != nil {
		return nil, err
	}

	return &diagnostics, nil
}

type StatsResponse struct {
	Success bool `json:"success"`
	Records struct {
		Pc [][]json.RawMessage `json:"Pc"`
	} `json:"records"`
	Totals struct {
		Pb  float64 `json:"Pb"`
		Pc  float64 `json:"Pc"`
		Gb  float64 `json:"Gb"`
		Gc  float64 `json:"Gc"`
		Pg  float64 `json:"Pg"`
		Bc  float64 `json:"Bc"`
		Kwh float64 `json:"kwh"`
	} `json:"totals"`
}

// Request the so-called energy readings for a given installation/site for a given period and interval.
// @todo Add other params
func (s *vrmSession) Stats(siteID int) (*StatsResponse, error) {
	url, err := formatURL(statsURL, URLParams{
		"siteID": strconv.Itoa(siteID),
	}, struct {
		StartTime    int64  `url:"start"`
		EndTime      int64  `url:"end"`
		Interval     string `url:"interval"`
		Type         string `url:"type"`
		ShowInstance bool   `json:"show_instance,omitempty"`
	}{time.Now().AddDate(0, -12, 0).Unix(), time.Now().Unix(), "15mins", "kwh", false})
	if err != nil {
		return nil, err
	}

	stats := StatsResponse{}
	if err := s.getAndLoad(url, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// Retrieve base64 encoded exports of installation data
// @todo Add other params
func (s *vrmSession) DownloadData(siteID int) ([]byte, error) {
	url, err := formatURL(downloadURL, URLParams{
		"siteID": strconv.Itoa(siteID),
	}, struct {
		Start    int64  `url:"start"`
		End      int64  `url:"end"`
		Format   string `url:"format"`
		Datatype string `url:"datatype"`
	}{time.Now().AddDate(0, -12, 0).Unix(), time.Now().Unix(), "csv", "kwh"})
	if err != nil {
		return nil, err
	}

	res, err := s.request(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for downloading data failed with status code %d: %w", res.StatusCode, err)
	}

	var data bytes.Buffer
	_, err = io.Copy(&data, res.Body)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

type UsersResponse struct {
	Success bool `json:"success"`
}

func (s *vrmSession) Users() (*UsersResponse, error) {
	url, err := formatURL(usersURL, URLParams{}, nil)
	if err != nil {
		return nil, err
	}

	data := UsersResponse{}
	if err := s.getAndLoad(url, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
