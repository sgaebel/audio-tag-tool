package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ssrathi/go-attr" // https://github.com/sentriz/go-taglib
	yaml "gopkg.in/yaml.v3"
)

var ProcessorHash string = ""

type Track struct {
	Artist       string `json:"artist" yaml:"artist"`
	Album        string `json:"album" yaml:"album"`
	CoverPath    string `json:"cover" yaml:"cover"`
	Date         string `json:"date" yaml:"date"`
	Disk         int    `json:"disk" yaml:"disk"`
	DiskTotal    int    `json:"disktotal" yaml:"disktotal"`
	Genre        string `json:"genre" yaml:"genre"`
	Instrumental bool   `json:"instrumental" yaml:"instrumental"`
	Language     string `json:"lang" yaml:"lang"`
	SourceURL    string `json:"url" yaml:"url"`

	FileName string `json:"path" yaml:"path"`
	// FileHashMD5 string `json:"md5" yaml:"md5"`
	Title       string `json:"title" yaml:"title"`
	TrackNumber int    `json:"track" yaml:"track"`
}

type Album struct {
	Artist       string `json:"artist" yaml:"artist"`
	Album        string `json:"album" yaml:"album"`
	CoverPath    string `json:"cover" yaml:"cover"`
	Date         string `json:"date" yaml:"date"`
	Disk         int    `json:"disk" yaml:"disk"`
	DiskTotal    int    `json:"disktotal" yaml:"disktotal"`
	Genre        string `json:"genre" yaml:"genre"`
	Instrumental bool   `json:"instrumental" yaml:"instrumental"`
	Language     string `json:"lang" yaml:"lang"`
	SourceURL    string `json:"url" yaml:"url"`

	Tracks []Track `json:"tracks" yaml:"tracks"`
}

func (t *Track) ApplyAlbumDefaults(album Album) {
	if t.Artist == "" {
		t.Artist = album.Artist
	}
	if t.Album == "" {
		t.Album = album.Album
	}
	if t.Date == "" {
		t.Date = album.Date
	}
	if t.Language == "" {
		t.Language = album.Language
	}
	if t.Disk == 0 {
		t.Disk = album.Disk
	}
	if t.DiskTotal == 0 {
		t.DiskTotal = album.DiskTotal
	}
	if t.Genre == "" {
		t.Genre = album.Genre
	}
	if t.CoverPath == "" {
		t.CoverPath = album.CoverPath
	}
	if t.SourceURL == "" {
		t.SourceURL = album.SourceURL
	}
	if !t.Instrumental && album.Instrumental {
		t.Instrumental = album.Instrumental
	}
}

func (t *Track) FromMetaFile(fileName string) error {
	var data []byte
	var err error
	if data, err = os.ReadFile(fileName); err != nil {
		return fmt.Errorf("Failed to load meta file %q: %v", fileName, err)
	}

	if strings.HasSuffix(strings.ToLower(fileName), ".json") {
		if err := json.Unmarshal(data, t); err != nil {
			return fmt.Errorf("Failed to unmarshal JSON: %v", err)
		}
	} else if strings.HasSuffix(strings.ToLower(fileName), ".yaml") {
		if err := yaml.Unmarshal(data, t); err != nil {
			return fmt.Errorf("Failed to unmarshal YAML: %v", err)
		}
	} else {
		return fmt.Errorf("Unknown file extension on %q", fileName)
	}
	return nil
}

func (a *Album) FromMetaFile(fileName string) error {
	var data []byte
	var err error
	if data, err = os.ReadFile(fileName); err != nil {
		return fmt.Errorf("Failed to load meta file %q: %v", fileName, err)
	}

	if strings.HasSuffix(strings.ToLower(fileName), ".json") {
		if err := json.Unmarshal(data, a); err != nil {
			return fmt.Errorf("Failed to unmarshal JSON: %v", err)
		}
	} else if strings.HasSuffix(strings.ToLower(fileName), ".yaml") {
		if err := yaml.Unmarshal(data, a); err != nil {
			return fmt.Errorf("Failed to unmarshal YAML: %v", err)
		}
	} else {
		return fmt.Errorf("Unknown file extension on %q", fileName)
	}
	return nil
}

func (t *Track) SetMetadata() error {
	// var tags map[string][]string
	tags := make(map[string][]string)
	for _, key := range Must(attr.Names(t)) {
		if key == "CoverPath" || key == "FileName" {
			continue
		}
		strValue := []string{""}
		switch Must(attr.GetKind(t, key)) {
		case "int":
			strValue[0] = fmt.Sprintf("%d", Must(attr.GetValue(t, key)).(int))
		case "string":
			strValue[0] = Must(attr.GetValue(t, key)).(string)
		case "bool":
			strValue[0] = strconv.FormatBool(Must(attr.GetValue(t, key)).(bool))
		}

		tags[key] = strValue
	}
	WriteMetadata(t.FileName, tags, true)
	if t.CoverPath != "" {
		return EmbedImage(t.FileName, t.CoverPath)
	}
	return nil
}

func (a *Album) GetTracks() []Track {
	for _, track := range a.Tracks {
		track.ApplyAlbumDefaults(*a)
	}
	return a.Tracks
}

func SetMetadataMain(parser *flag.FlagSet, argv []string) {
	metaPath := parser.String("meta", "", "Path to the metadata including file names of the tracks")
	artist := parser.String("artist", "", "Set artist")
	albumName := parser.String("album", "", "Album name")
	title := parser.String("title", "", "Song title")
	genre := parser.String("genre", "", "Music genre")
	sourceUrl := parser.String("url", "", "Source URL")
	isInstrumental := parser.Bool("instrumental", false, "Is instumental track")
	language := parser.String("lang", "", "Track language")
	coverPath := parser.String("cover", "", "Path of the cover image")
	date := parser.String("date", "", "Release date")
	parser.Parse(argv)

	args := parser.Args()
	if *metaPath == "" && len(args) != 1 {
		fmt.Fprintln(os.Stderr, "If no 'meta' is given, a positional fileName is required")
		os.Exit(1)
	}
	if *metaPath != "" && !FileExists(*metaPath) {
		fmt.Fprintf(os.Stderr, "Error: metadata file %q does not exist\n", *metaPath)
	}
	var album Album
	if err := album.FromMetaFile(*metaPath); err != nil {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: metadata file %q does not seem to be an album, and no fileName was given", *metaPath)
			os.Exit(1)
		}
	} else {
		if *artist != "" {
			album.Artist = *artist
		}
		if *albumName != "" {
			album.Album = *albumName
		}
		if *genre != "" {
			album.Genre = *genre
		}
		if *sourceUrl != "" {
			album.SourceURL = *sourceUrl
		}
		if *language != "" {
			album.Language = *language
		}
		if *date != "" {
			album.Date = *date
		}
		if *isInstrumental {
			album.Instrumental = true
		}
		if *coverPath != "" {
			album.CoverPath = *coverPath
		}
		for idx, track := range album.GetTracks() {
			if err := track.SetMetadata(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to set metadata for track %d: %v", idx+1, err)
				os.Exit(2)
			}
		}
		os.Exit(0)
	}
	var track Track
	if *metaPath != "" {
		if err := track.FromMetaFile(*metaPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to parse metadata file %q: %v\n", *metaPath, err)
			os.Exit(1)
		}
	} else {
		if !FileExists(args[0]) {
			fmt.Fprintf(os.Stderr, "File %q does not exist", args[0])
			os.Exit(1)
		}
		track.FileName = args[0]
	}
	if *artist != "" {
		track.Artist = *artist
	}
	if *albumName != "" {
		track.Album = *albumName
	}
	if *title != "" {
		track.Title = *title
	}
	if *genre != "" {
		track.Genre = *genre
	}
	if *sourceUrl != "" {
		track.SourceURL = *sourceUrl
	}
	if *language != "" {
		track.Language = *language
	}
	if *date != "" {
		track.Date = *date
	}
	if *isInstrumental {
		track.Instrumental = true
	}
	if *coverPath != "" {
		track.CoverPath = *coverPath
	}
	if err := track.SetMetadata(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to set metadata for track %q: %v", track.FileName, err)
		os.Exit(2)
	}
}

func EmbedImageMain(parser *flag.FlagSet, argv []string) {
	parser.Parse(argv)
	args := parser.Args()
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Error: EmbedImage requires exactly 2 positional arguments, 'fileName' and 'imageName'")
		os.Exit(1)
	}
	audioFileName := args[0]
	if !FileExists(audioFileName) {
		fmt.Fprintf(os.Stderr, "Error: audio file %q does not exist\n", audioFileName)
		os.Exit(1)
	}
	imageFileName := args[1]
	if !FileExists(imageFileName) {
		fmt.Fprintf(os.Stderr, "Error: image file %q does not exist\n", imageFileName)
		os.Exit(1)
	}
	if err := EmbedImage(audioFileName, imageFileName); err != nil {
		fmt.Fprintf(os.Stderr, "Error while embedding image %q into %q: %v", imageFileName, audioFileName, err)
		os.Exit(2)
	}
}

func WipeMetadataMain(parser *flag.FlagSet, argv []string) {
	parser.Parse(argv)
	args := parser.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Error: expected exactly positional argument for 'wipeMetadata', the fileName.")
		os.Exit(1)
	}
	fileName := args[0]
	if !FileExists(fileName) {
		fmt.Fprintf(os.Stderr, "Error: input file %q does not exist\n", fileName)
		os.Exit(1)
	}
	if err := WipeMetadata(fileName); err != nil {
		fmt.Fprintf(os.Stderr, "Error while wiping metadata: %v", err)
		os.Exit(2)
	}
}

func PrintMetadataMain(parser *flag.FlagSet, argv []string) {
	parser.Parse(argv)
	args := parser.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Error: expected exactly positional argument for 'printMetadata', the fileName.")
		os.Exit(1)
	}
	fileName := args[0]
	if !FileExists(fileName) {
		fmt.Fprintf(os.Stderr, "Error: input file %q does not exist\n", fileName)
		os.Exit(1)
	}
	if err := PrintMetadata(fileName); err != nil {
		fmt.Fprintf(os.Stderr, "Error while printing metadata and tags: %v", err)
		os.Exit(2)
	}
}

func main() {
	// Define subcommands
	wipeMetadataCmd := flag.NewFlagSet("wipeMetadata", flag.ExitOnError)
	embedImageCmd := flag.NewFlagSet("embedImage", flag.ExitOnError)
	setMetadataCmd := flag.NewFlagSet("setMetadata", flag.ExitOnError)
	printMetadataCmd := flag.NewFlagSet("printMetadata", flag.ExitOnError)

	// Parse the subcommand
	if len(os.Args) < 2 {
		fmt.Println("Expected 'wipeMetadata', 'embedImage', 'setMetadata', or 'printMetadata' subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "wipeMetadata":
		fmt.Println("Executing wipeMetadata subcommand")
		WipeMetadataMain(wipeMetadataCmd, os.Args[1:])
	case "embedImage":
		fmt.Println("Executing embedImage subcommand")
		EmbedImageMain(embedImageCmd, os.Args[1:])
	case "setMetadata":
		fmt.Println("Executing setMetadata subcommand")
		SetMetadataMain(setMetadataCmd, os.Args[1:])
	case "printMetadata":
		fmt.Println("Executing printMetadata subcommand")
		PrintMetadataMain(printMetadataCmd, os.Args[1:])
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}
