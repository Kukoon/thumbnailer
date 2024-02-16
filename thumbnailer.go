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

// timeStamps returns four timeStamps at 20, 40, 60 and 80% of a given
// video file as a slice of time.Duration
func timeStamps(file string) []time.Duration {
	var stampAt = [4]float64{0.2, 0.4, 0.6, 0.8}
	var timeStamps []time.Duration

	args := []string{"-i", file, "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=duration", "-of", "default=noprint_wrappers=1:nokey=1"}
	out, err := exec.Command("ffprobe", args...).Output()
	if err != nil {
		log.Fatal(err)
	}

	s := string(out)
	s = strings.TrimRight(s, "\n")
	f, _ := strconv.ParseFloat(s, 32)

	for _, v := range stampAt {
		timeStamps = append(timeStamps, time.Duration(f*v)*time.Second)
	}
	return timeStamps
}

// generateSegments cuts a given file at the timeStamps
// and writes the results into the current directory
func generateSegments(file string) (result []string) {
	ts := timeStamps(file)
	segmentDuration := "2"
	ext := filepath.Ext(file)
	name := file[0 : len(file)-len(ext)]

	for i, t := range ts {
		filename := fmt.Sprintf("%s_segment_%d%s", name, i, ext)
		args := []string{"-y", "-i", file, "-ss", strconv.FormatFloat(t.Seconds(), 'f', 0, 64), "-t", segmentDuration, "-map", "0:v:0", "-vcodec", "copy", filename}
		cmd := exec.Command("ffmpeg", args...)
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, filename)
	}
	return
}

// makePreview generates the actual webp preview and writes it to current
// directory
func makePreview(file string) {
	files := generateSegments(file)
	var list string = "list.txt"

	// Create a temporary list that ffmppeg concat demuxer can consume
	f, err := os.Create(list)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Write files to list -- prefixed with "file"
	for _, file := range files {
		fmt.Fprintln(f, "file", file)
	}

	// Store file extension and file name
	ext := filepath.Ext(file)
	name := file[0 : len(file)-len(ext)]

	args := []string{"-y", "-safe", "0", "-f", "concat", "-i", filepath.Base(list), "-an", "-vcodec", "libwebp", "-loop", "0", "-preset", "picture", "-vf", "fps=6,scale=480:-1:flags=lanczos", "-qscale", "40", "-compression_level", "6", name + "_preview.webp"}
	cmd := exec.Command("ffmpeg", args...)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Clean up temporarily created files
	files = append(files, filepath.Base(list))
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	makePreview(file)
}
