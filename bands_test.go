package bandplan

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuiltinsValid(t *testing.T) {
	names := map[string]bool{}
	for _, bs := range Builtins() {
		if err := bs.Validate(); err != nil {
			t.Error(err)
		}
		if names[bs.Name] {
			t.Errorf("duplicate band set %q", bs.Name)
		}
		names[bs.Name] = true
	}
	for _, n := range []string{
		NameWiFi24, NameWiFi5, NameFPV58, NameZigbee, NameBluetooth,
		NameLoRaUS915, NameLoRaEU868, NameLPD433, NamePMR446,
		NameMURS, NameGMRS, NameMarineVHF, NameNOAAWeather, NameAmateurVHF,
	} {
		if Builtin(n) == nil {
			t.Errorf("no builtin %q", n)
		}
	}
}

func TestBuiltinsReturnsCopies(t *testing.T) {
	a := Builtins()
	a[0].Name = "changed"
	a[0].Bands[0].Channels[0].Center = 1
	b := Builtin(NameWiFi24)
	if b.Name != NameWiFi24 || b.Bands[0].Channels[0].Center != 2412*MHz {
		t.Fatal("modifying a returned set changed the builtins")
	}
}

func TestBuiltinValues(t *testing.T) {
	gmrs := Builtin(NameGMRS)
	if ch := gmrs.Bands[0].Channels[1]; ch.Name != "1" || ch.Center != 462_562_500 {
		t.Errorf("GMRS 1 = %+v", ch)
	}
	murs := Builtin(NameMURS)
	if ch := murs.Bands[0].Channels[0]; ch.Width != 11_250 {
		t.Errorf("MURS 1 width = %v", ch.Width)
	}
	if lo, hi := murs.Span(); lo != 151_814_375 || hi != 154_610_000 {
		t.Errorf("MURS span = %v..%v", lo, hi)
	}
	wifi5 := Builtin(NameWiFi5)
	if ch := wifi5.Bands[3].Channels[4]; ch.Name != "165" || ch.Center != 5825*MHz {
		t.Errorf("WiFi 165 = %+v", ch)
	}
	zigbee := Builtin(NameZigbee)
	chs := zigbee.Bands[0].Channels
	if chs[0].Name != "11" || chs[0].Center != 2405*MHz || chs[15].Name != "26" || chs[15].Center != 2480*MHz {
		t.Errorf("Zigbee = %+v .. %+v", chs[0], chs[15])
	}
	if len(Builtin(NameFPV58).Bands) != 10 {
		t.Errorf("FPV bands = %d", len(Builtin(NameFPV58).Bands))
	}
	if ch := wifi5.Bands[4].Channels[2]; ch.Name != "177" || ch.Center != 5885*MHz {
		t.Errorf("WiFi 177 = %+v", ch)
	}
}

// TestBuiltinChannels checks the first and last channel of each band built
// from a formula.
func TestBuiltinChannels(t *testing.T) {
	tests := []struct {
		set         string
		band        int
		count       int
		first, last Channel
	}{
		{NameZigbee, 0, 16, Channel{"11", 2405 * MHz, 2 * MHz}, Channel{"26", 2480 * MHz, 2 * MHz}},
		{NameBluetooth, 0, 3, Channel{"37", 2402 * MHz, 2 * MHz}, Channel{"39", 2480 * MHz, 2 * MHz}},
		{NameBluetooth, 1, 37, Channel{"0", 2404 * MHz, 2 * MHz}, Channel{"36", 2478 * MHz, 2 * MHz}},
		{NameBluetooth, 2, 79, Channel{"0", 2402 * MHz, MHz}, Channel{"78", 2480 * MHz, MHz}},
		{NameLoRaUS915, 0, 64, Channel{"0", 902_300 * KHz, 125 * KHz}, Channel{"63", 914_900 * KHz, 125 * KHz}},
		{NameLoRaUS915, 1, 8, Channel{"64", 903_000 * KHz, 500 * KHz}, Channel{"71", 914_200 * KHz, 500 * KHz}},
		{NameLoRaUS915, 2, 8, Channel{"0", 923_300 * KHz, 500 * KHz}, Channel{"7", 927_500 * KHz, 500 * KHz}},
		{NameLoRaEU868, 1, 5, Channel{"3", 867_100 * KHz, 125 * KHz}, Channel{"7", 867_900 * KHz, 125 * KHz}},
		{NameLPD433, 0, 69, Channel{"1", 433_075 * KHz, 25 * KHz}, Channel{"69", 434_775 * KHz, 25 * KHz}},
		{NamePMR446, 0, 16, Channel{"1", 446_006_250, 12_500}, Channel{"16", 446_193_750, 12_500}},
		{NameMarineVHF, 0, 57, Channel{"60", 156_025 * KHz, 16 * KHz}, Channel{"88", 157_425 * KHz, 16 * KHz}},
	}
	for _, tt := range tests {
		chs := Builtin(tt.set).Bands[tt.band].Channels
		if len(chs) != tt.count || chs[0] != tt.first || chs[len(chs)-1] != tt.last {
			t.Errorf("%s band %d: %d channels, %+v .. %+v", tt.set, tt.band, len(chs), chs[0], chs[len(chs)-1])
		}
	}
	marine := Builtin(NameMarineVHF).Bands[0].Channels
	for _, ch := range marine {
		if ch.Name == "16" && ch.Center != 156_800*KHz {
			t.Errorf("marine 16 = %+v", ch)
		}
	}
}

func TestChannelEdges(t *testing.T) {
	ch := Channel{Center: 151_820_000, Width: 11_250}
	if ch.Low() != 151_814_375 || ch.High() != 151_825_625 || ch.High()-ch.Low() != ch.Width {
		t.Fatalf("edges %v..%v", ch.Low(), ch.High())
	}
	if !ch.Contains(151_814_375) || !ch.Contains(151_825_625) || ch.Contains(151_825_626) {
		t.Fatal("Contains is not inclusive of exactly the edges")
	}
}

func TestSpanEmpty(t *testing.T) {
	if lo, hi := (&BandSet{}).Span(); lo != 0 || hi != 0 {
		t.Fatalf("empty span %v..%v", lo, hi)
	}
}

func TestValidate(t *testing.T) {
	ok := Channel{Name: "1", Center: MHz, Width: KHz}
	tests := []struct {
		name string
		bs   BandSet
		want string
	}{
		{"no set name", BandSet{}, "no name"},
		{"no band name", BandSet{Name: "s", Bands: []Band{{}}}, "band has no name"},
		{"no channel name", BandSet{Name: "s", Bands: []Band{{Name: "b", Channels: []Channel{{Center: 1, Width: 1}}}}}, "channel has no name"},
		{"zero width", BandSet{Name: "s", Bands: []Band{{Name: "b", Channels: []Channel{{Name: "1", Center: 1}}}}}, "width"},
		{"zero center", BandSet{Name: "s", Bands: []Band{{Name: "b", Channels: []Channel{{Name: "1", Width: 1}}}}}, "center"},
		{"valid", BandSet{Name: "s", Bands: []Band{{Name: "b", Channels: []Channel{ok}}}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bs.Validate()
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestLoadSaveFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "bands.json")

	sets, err := LoadFile(path)
	if err != nil || sets != nil {
		t.Fatalf("missing file: %v, %v", sets, err)
	}

	want := []*BandSet{Builtin(NameMURS)}
	if err := SaveFile(path, want); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"widthHz": 11250`) || !strings.HasPrefix(string(data), "{\n  \"version\": 1,") {
		t.Errorf("unexpected file:\n%s", data)
	}
	got, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	gj, _ := json.Marshal(got)
	wj, _ := json.Marshal(want)
	if string(gj) != string(wj) {
		t.Fatalf("round trip\n got %s\nwant %s", gj, wj)
	}

}

func TestDecodeErrors(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{`{"version":1,"sets":[{"name":"x","bands":[{"name":"b","channels":[{"name":"1","centerHz":1000000,"widthHz":0}]}]}]}`, "width"},
		{`{"version":1,"sets":[{"name":"x","bands":[{"name":"b","channels":[{"name":"1","centerHz":9007199254740992,"widthHz":1}]}]}]}`, "exceeds"},
		{`{"version":1,"sets":[{"name":"x","bands":[{"name":"b","channels":[{"name":"1","centerHz":1.5,"widthHz":1}]}]}]}`, "parsing"},
		{`{"version":1,"sets":[null]}`, "null"},
		{`{"version":2,"sets":[]}`, "version 2"},
		{`{"sets":[]}`, "version 0"},
		{`{"version":1,"sets":[],"extra":1}`, "unknown field"},
		{`[]`, "parsing"},
	} {
		if _, err := Decode([]byte(tt.in)); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("Decode(%s) err = %v, want containing %q", tt.in, err, tt.want)
		}
	}
}

func TestGeneratedFile(t *testing.T) {
	want, err := Encode(Builtins())
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile("bandplan.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("bandplan.json is out of date; run go generate")
	}
	sets, err := Decode(got)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sets, Builtins()) {
		t.Fatal("bandplan.json does not decode to Builtins")
	}
}

func TestSchemaVersion(t *testing.T) {
	data, err := os.ReadFile("bandplan.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties struct {
			Version struct {
				Const int `json:"const"`
			} `json:"version"`
		} `json:"properties"`
		Defs struct {
			Hertz struct {
				Maximum int64 `json:"maximum"`
			} `json:"hertz"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if schema.Properties.Version.Const != FileVersion {
		t.Errorf("schema version %d, want %d", schema.Properties.Version.Const, FileVersion)
	}
	if schema.Defs.Hertz.Maximum != int64(MaxFrequency) {
		t.Errorf("schema hertz maximum %d, want %d", schema.Defs.Hertz.Maximum, int64(MaxFrequency))
	}
}

func TestMergeSets(t *testing.T) {
	custom := []*BandSet{{Name: NameMURS}, {Name: "Mine"}}
	merged := MergeSets(Builtins(), custom)
	if Find(merged, NameMURS) != custom[0] || Find(merged, "Mine") != custom[1] || Find(merged, NameGMRS) == nil {
		t.Fatal("merge did not replace or keep sets")
	}
	if len(merged) != len(builtins)+1 {
		t.Fatalf("merged %d sets", len(merged))
	}
}

func TestLookup(t *testing.T) {
	gmrs := Builtin(NameGMRS)
	// 462.570 MHz lies where the 20 kHz GMRS 1 and 16 overlap.
	refs := gmrs.Lookup(462_570_000)
	var names []string
	for _, r := range refs {
		names = append(names, gmrs.Bands[r.Band].Channels[r.Channel].Name)
	}
	if strings.Join(names, ",") != "1,16" {
		t.Errorf("Lookup = %v", names)
	}
	if refs := gmrs.Lookup(467_712_500); len(refs) != 1 || refs[0] != (ChannelRef{1, 6}) {
		t.Errorf("Lookup GMRS 14 = %v", refs)
	}
	if refs := gmrs.Lookup(100 * MHz); refs != nil {
		t.Errorf("Lookup outside = %v", refs)
	}
}
