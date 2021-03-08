package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var file string = os.Args[1]

// segmentTimes returns the start and end times with 3 seconds in between
// at 20, 40, 60 and 80% of a given video duration in the form of an array of type string
func segmentTimes(file string) []time.Duration {
	var cutAt = [4]float64{0.2, 0.4, 0.6, 0.8}
	var splitSecs []time.Duration

	args := []string{"-i", file, "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=duration", "-of", "default=noprint_wrappers=1:nokey=1"}
	out, err := exec.Command("ffprobe", args...).Output()
	if err != nil {
		log.Fatal(err)
	}

	s := string(out)
	s = strings.TrimRight(s, "\n")
	f, _ := strconv.ParseFloat(s, 32)

	for _, v := range cutAt {
		splitSecs = append(splitSecs, time.Duration(f*v)*time.Second)
	}
	return splitSecs
}

// generateSegments cuts a given file at the times returned by segmentTimes and
// writes the results into the current directory
func generateSegments(file string) (result []string) {
	times := segmentTimes(file)
	longD := "30"
	shortD := "2"
	ext := filepath.Ext(file)
	name := file[0 : len(file)-len(ext)]

	for i, t := range times {
		filenameLong := fmt.Sprintf("%s_long_%d.mp4", name, i)
		args := []string{"-y", "-i", file, "-ss", strconv.FormatFloat(t.Seconds(), 'f', 0, 64), "-t", longD, "-map", "0:v:0", "-vcodec", "copy", filenameLong}
		cmd := exec.Command("ffmpeg", args...)
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		filenameShort := fmt.Sprintf("%s_short_%d.ts", name, i)
		args = []string{"-y", "-i", filenameLong, "-t", shortD, "-c", "copy", "-bsf:v", "h264_mp4toannexb", "-f", "mpegts", filenameShort}
		cmd = exec.Command("ffmpeg", args...)
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, filenameShort)

		if err := os.Remove(filenameLong); err != nil {
			log.Fatal(err)
		}
	}
	return
}

// makePreview generates the actual webp preview and writes it to current
// directory
func makePreview(file string) {
	files := generateSegments(file)
	list := strings.Join(files, "|")

	ext := filepath.Ext(file)
	name := file[0 : len(file)-len(ext)]

	args := []string{"-y", "-i", "concat:" + list, "-an", "-vcodec", "libwebp", "-loop", "0", "-preset", "picture", "-vf", "fps=6,scale=480:-1:flags=lanczos", "-qscale", "40", "-compression_level", "6", name + "_preview.webp"}
	cmd := exec.Command("ffmpeg", args...)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	makePreview(file)
}
