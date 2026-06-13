package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"go.senan.xyz/taglib"
)

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func EmbedImage(fileName, imageFileName string) error {
	MimeType := ""
	if strings.HasSuffix(strings.ToLower(imageFileName), ".jpeg") || strings.HasSuffix(strings.ToLower(imageFileName), ".jpg") {
		MimeType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(imageFileName), ".png") {
		MimeType = "image/png"
	} else {
		return fmt.Errorf("Image type not implemented: file name = %q", imageFileName)
	}
	// Read image data from file
	imageData, err := os.ReadFile(imageFileName)
	if err != nil {
		return fmt.Errorf("Error reading image file: %v", err)
	}

	if err := taglib.WriteImageOptions(
		fileName,
		imageData,
		0,               // index to replace; use higher index to append
		"Front Cover",   // picture type
		"Front artwork", // description
		MimeType,        // MIME type
	); err != nil {
		return fmt.Errorf("Error embedding image: %v", err)
	}
	return nil
}

func PrintMetadata(fileName string) error {
	properties, err := taglib.ReadProperties(fileName)
	if err != nil {
		log.Fatalf("Error reading properties: %v", err)
	}

	fmt.Printf("Length: %v\n", properties.Length)
	fmt.Printf("Bitrate: %d\n", properties.Bitrate)
	fmt.Printf("SampleRate: %d\n", properties.SampleRate)
	fmt.Printf("Channels: %d\n", properties.Channels)

	// Image metadata (without reading actual image data)
	for i, img := range properties.Images {
		fmt.Printf("Image %d - Type: %s, Description: %s, MIME type: %s\n",
			i, img.Type, img.Description, img.MIMEType)
	}
	tags, err := taglib.ReadTags("path/to/audiofile.mp3")
	if err != nil {
		log.Fatalf("Error reading metadata: %v", err)
		return err
	}

	fmt.Printf("tags: %v\n", tags) // map[string][]string

	fmt.Printf("AlbumArtist: %q\n", tags[taglib.AlbumArtist])
	fmt.Printf("Album: %q\n", tags[taglib.Album])
	fmt.Printf("TrackNumber: %q\n", tags[taglib.TrackNumber])
	return nil
}

func WriteMetadata(filename string, tags map[string][]string, clearOldTags bool) error {
	if clearOldTags {
		if err := taglib.WriteTags(filename, tags, taglib.Clear); err != nil {
			return fmt.Errorf("Failed to wipe metadata: %v", err)
		}
	} else {
		if err := taglib.WriteTags(filename, tags, 0); err != nil {
			return fmt.Errorf("Failed to write metadata: %v", err)
		}
	}
	return nil
}

func WipeMetadata(filename string) error {
	return taglib.WriteTags(filename, nil, taglib.Clear)
}
