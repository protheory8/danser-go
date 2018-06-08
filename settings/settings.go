package settings

import (
	"os"
	"strconv"
	"encoding/json"
)

const SETTINGSVERSION = "v1"

type general struct {
	OsuDir string //localappdata
}

type graphics struct {
	Width, Height int64
	WindowWidth, WindowHeight int64
	Fullscreen bool //true
	VSync bool //false
	FPSCap int64 //1000
	MSAA int32 //16
}

func (gr graphics) GetSize() (int64, int64) {
	if gr.Fullscreen {
		return gr.Width, gr.Height
	}
	return gr.WindowWidth, gr.WindowHeight
}

func (gr graphics) GetSizeF() (float64, float64) {
	if gr.Fullscreen {
		return float64(gr.Width), float64(gr.Height)
	}
	return float64(gr.WindowWidth), float64(gr.WindowHeight)
}

type audio struct {
	GeneralVolume float64 //0.5
	MusicVolume float64 //=0.5
	SampleVolume float64 //=0.5
	EnableBeatmapSampleVolume bool //= false
}

type cursor struct {
	EnableRainbow bool //true
	RainbowSpeed float64 //8, degrees per second
	Hue float64 //0..360, if EnableRainbow is disabled then this value will be used to calculate base color
	EnableCustomHueOffset bool //false, false means that every iteration has an offset of i*360/n
	HueOffset float64 //0, custom hue offset for mirror collages
	EnableCustomTrailGlowOffset bool //true, if enabled, value set below will be used, if not, HueOffset of previous iteration will be used (or offset of 180° for single cursor)
	TrailGlowOffset float64 //-36, offset of the cursor trail glow
	ScaleToCS bool //false, if enabled, cursor will scale to beatmap CS value
	CursorSize float64 //18, cursor radius in osu!pixels
	ScaleToMusicPower bool //true, cursor size is changing with music peak amplitude
	ShowCursorsOnBreaks bool //true
}

type objects struct {
	MandalaTexturesTrigger int64 //5, minimum value of cursors needed to use more translucent textures
	UseCursorColors bool //true, overrides lower color settings
	EnableRainbow bool //true
	RainbowSpeed float64 //..., degrees per second
	Hue float64 //0..360, if EnableRainbow is disabled then this value will be used to calculate base color
	EnableCustomHueOffset bool //false, false means that every iteration has an offset of i*360/n
	HueOffset float64 //0, custom hue offset for mirror collages
	ObjectsSize float64 //-1, objects radius in osu!pixels. If value is less than 0, beatmap's CS will be used
	ScaleToMusicPower bool //true, objects size is changing with music peak amplitude
}

type playfield struct {
	LeadInTime float64 //5, time to the beginning of music
	BackgroundInDim float64 //0, background dim at the start of app
	BackgroundDim float64 // 0.95, background dim at the beatmap start
	BackgroundDimBreaks float64 // 0.95, background dim at the breaks
	FlashToMusicPower bool //true, background dim varies accoriding to music power
	KiaiFactor float64 //1.2, scale and flash factor during Kiai
}

type fileformat struct {
	Version *string
	General *general
	Graphics *graphics
	Audio *audio
	Cursor *cursor
	Objects *objects
	Playfield *playfield
}

var Version string
var General *general
var Graphics *graphics
var Audio *audio
var Cursor *cursor
var Objects *objects
var Playfield *playfield

var fileStorage *fileformat
var fileName string
func initDefaults() {
	Version = SETTINGSVERSION
	General = &general{os.Getenv("localappdata") + string(os.PathSeparator) + "osu!" + string(os.PathSeparator) + "Songs" + string(os.PathSeparator)}
	Graphics = &graphics{1920, 1080, 1280, 720, true, false, 1000, 16}
	Audio = &audio{0.5, 0.5, 0.5, false}
	Cursor = &cursor{true, 8, 0, false, 0, true, -36.0, false, 18, true, true}
	Objects = &objects{5, true, true, 8, 0, false, 0, -1, true}
	Playfield = &playfield{5, 0, 0.95, 0.95, true, 1.1}
	fileStorage = &fileformat{&Version, General, Graphics, Audio, Cursor, Objects, Playfield}
}

func LoadSettings(version int) bool {
	initDefaults()
	fileName = "settings"

	if version > 0 {
		fileName += "-" + strconv.FormatInt(int64(version), 10)
	}
	fileName += ".json"

	file, err := os.Open(fileName)
	defer file.Close()
	if os.IsNotExist(err) {
		saveSettings(fileName)
		return true
	} else if err != nil {
		panic(err)
	} else {
		load(file)
		saveSettings(fileName) //this is done to save additions from the current format
	}

	return false
}

func load(file *os.File) {
	decoder := json.NewDecoder(file)
	decoder.Decode(fileStorage)
}

func Save() {
	saveSettings(fileName)
}

func saveSettings(path string) {
	file, err := os.Create(path)
	defer file.Close()

	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	Version = SETTINGSVERSION
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	encoder.Encode(fileStorage)
}

var DIVIDES = 10