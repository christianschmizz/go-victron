package vrm_test

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"victron/vrm"
)

func TestLoginAsDemo(t *testing.T) {
	session, err := vrm.LoginAsDemo()
	if assert.NoError(t, err) {
		users, err := session.Installations(session.UserID)
		if assert.NoError(t, err) {
			assert.True(t, users.Success)
		}
	}
}

func TestDemoInstallations(t *testing.T) {
	session, err := vrm.LoginAsDemo()
	if assert.NoError(t, err) {
		installs, err := session.Installations(vrm.DemoUserID)
		if assert.NoError(t, err) {
			assert.True(t, installs.Success)
			assert.Len(t, installs.Records, 36)
			for _, record := range installs.Records {
				siteID := strconv.Itoa(int(record.SiteID))

				overview, err := session.SystemOverview(siteID)
				if assert.NoError(t, err) {
					assert.True(t, overview.Success)
				}

				diag, err := session.Diagnostics(siteID, 1000)
				if assert.NoError(t, err) {
					assert.True(t, diag.Success)
				}
			}
		}
	}
}

func TestDiagnosticsData(t *testing.T) {
	x := []byte(`{
		"success": true,
		"records": [{
			"idSite": 1495,
			"timestamp": 1497056046,
			"Device": "Gateway",
			"instance": 0,
			"idDataAttribute": 1,
			"description": "gatewayID",
			"formatWithUnit": "%s",
			"dbusServiceType": null,
			"dbusPath": null,
			"formattedValue": "Venus",
			"dataAttributeEnumValues": [{
				"nameEnum": "VGR, VGR2 or VER",
				"valueEnum": 0
			},
			{
				"nameEnum": "Venus",
				"valueEnum": 1
			},
			{
				"nameEnum": "Venus",
				"valueEnum": 2
			}],
			"id": 1
		}],
		"num_records": 1
	}`)
	diag := vrm.DiagnosticsResponse{}
	err := json.NewDecoder(bytes.NewBuffer(x)).Decode(&diag)
	if assert.NoError(t, err) {
		assert.True(t, diag.Success)
		assert.Equal(t, uint(1), diag.NumRecords)
		assert.Equal(t, 1, len(diag.Records))
		assert.Equal(t, 3, len(diag.Records[0].DataAttributeEnumValues))
	}
}

func TestStatsData(t *testing.T) {
	x := []byte(`{
  "success": true,
  "records": {
    "Pc": [
      [
        1441066216000,
        12.927161
      ],
      [
        1441152616000,
        28.52883
      ],
      [
        1441239016000,
        17.722068
      ],
      [
        1441325416000,
        9.1537885
      ],
      [
        1441411816000,
        4.453626
      ],
      [
        1441498216000,
        4.5079915
      ],
      [
        1441584616000,
        16.8285763
      ],
      [
        1441671016000,
        12.1123506
      ],
      [
        1441757416000,
        29.2207336
      ],
      [
        1441843816000,
        29.7107766
      ],
      [
        1441930216000,
        13.5401983
      ],
      [
        1442016616000,
        5.872294
      ]
    ]
  },
  "totals": {
    "Pb": 2.5122129,
    "Pc": 184.5783944,
    "Gb": 9.3752899,
    "Gc": 252.0008088,
    "Pg": 64.7871119,
    "Bc": 0.0182044,
    "kwh": 513.2720223
  }
}
`)
	stats := vrm.StatsResponse{}
	err := json.NewDecoder(bytes.NewBuffer(x)).Decode(&stats)
	if assert.NoError(t, err) {
		assert.True(t, stats.Success)
		assert.Equal(t, 12, len(stats.Records.Pc))
		assert.Equal(t, float64(2.5122129), stats.Totals.Pb)
		assert.Equal(t, float64(513.2720223), stats.Totals.Kwh)
	}
}
