package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var file string = os.Args[1]

// segmentTimes returns the start and end times with 3 seconds in between
// at 20, 40, 60 and 80% of a given videos duration in the form of an array of type string
func segmentTimes(file string) []string {
	var cutAt = [4]float64{0.2, 0.4, 0.6, 0.8}
	var splitSecs [4]int
	var result []string

	args := []string{"-i", file, "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=duration", "-of", "default=noprint_wrappers=1:nokey=1"}
	out, err := exec.Command("ffprobe", args...).Output()
	if err != nil {
		log.Fatal(err)
	}

	s := string(out)
	s = strings.TrimRight(s, "\n")
	f, _ := strconv.ParseFloat(s, 32)

	for i, v := range cutAt {
		splitSecs[i] = int(f * v)
	}

	for _, v := range splitSecs {
		result = append(result, strconv.Itoa(v))
		result = append(result, strconv.Itoa(v+3))
	}
	return result
}

// generateSegments cuts a given file at the times returned by segmentTimes and
// writes the results into the current directory
func generateSegments(file string) {
	var t []string = segmentTimes(file)

	args := []string{"-i", file, "-map", "0", "-c", "copy", "-f", "segment", "-segment_times", t[0] + "," + t[1] + "," + t[2] + "," + t[3] + "," + t[4] + "," + t[5] + "," + t[6] + "," + t[7], "-reset_timestamps", "1", "-y", "segment_%3d.mp4"}

	cmd := exec.Command("ffmpeg", args...)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Wait()
}

// removeEven removes even segments, as we only want the odd ones
func removeEven() {
	var exp string = "segment_00[0|2|3|4|6|8].mp4"
	root := "."
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			r, err := regexp.MatchString(exp, info.Name())
			if err == nil && r {
				e := os.Remove(path)
				if e != nil {
					log.Fatal(e)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

// writeToFile writes either odd or all segments to a temporary
// 'segment_files.txt' that is accepted by ffmpeg concat demuxer
func writeToFile(num string) {
	var files []string
	var exp string
	if num == "odd" {
		exp = "segment_00[1|3|5|7|9].mp4"
	} else {
		exp = "segment_00[0-9].mp4"
	}
	root := "."
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			r, err := regexp.MatchString(exp, info.Name())
			if err == nil && r {
				files = append(files, "file")
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Create("segment_files.txt")
	if err != nil {
		log.Fatal(err)
	} else {
		for i := 0; i < len(files); i += 2 {
			file.WriteString(files[i])
			file.WriteString(" '" + files[i+1] + "'\n")
		}
	}
	file.Close()
}

// concatSegments concatenates segments specified in 'segment_files.txt'
// saves result as 'concat.mp4'
// cleans up after itself and removes all segments
func concatSegments() {
	args := []string{"-y", "-f", "concat", "-safe", "0", "-i", "./segment_files.txt", "-force_key_frames", "00:00:00.000", "-x264-params", "keyint=1:scenecut=0", "-c", "copy", "concat.mp4"}
	cmd := exec.Command("ffmpeg", args...)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Wait()

	// Remove Segments
	var exp string = "segment_00[0-9].mp4"
	root := "."
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			r, err := regexp.MatchString(exp, info.Name())
			if err == nil && r {
				e := os.Remove(path)
				if e != nil {
					log.Fatal(e)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

// quarterSegments iterates over all segments in current dir
// shortens segments to 2s and saves them as quarter_segments
// cleans up the old segments
func quarterSegments() {
	var files []string
	var exp string = "segment_00[0-9].mp4"
	root := "."
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			r, err := regexp.MatchString(exp, info.Name())
			if err == nil && r {
				files = append(files, "file")
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	for i, file := range files {
		args := []string{"-y", "-i", file, "-ss", "00:00:00.000", "-t", "2", "half_segment_00" + strconv.Itoa(i) + ".mp4"}
		cmd := exec.Command("ffmpeg", args...)
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		cmd.Wait()
	}

	// Remove segments but __not__ half_segments
	exp = "^segment_00[0-9].mp4"
	root = "."
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			r, err := regexp.MatchString(exp, info.Name())
			if err == nil && r {
				e := os.Remove(path)
				if e != nil {
					log.Fatal(e)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

// makePreview generates the actual webp preview and writes it to current
// directory
func makePreview(file string) {
	args := []string{"-y", "-i", file, "-vcodec", "libwebp", "-filter:v", "fps=fps=20", "-loop", "0", "-preset", "default", "-an", "-vsync", "0", "-vf", "scale=480:-1", "preview.webp"}
	cmd := exec.Command("ffmpeg", args...)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Wait()
}

// cleanUp cleans the working directory off of the tmp files that have been
// created
func cleanUp() {
	err := os.Remove("concat.mp4")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove("segment_files.txt")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	generateSegments(file)
	removeEven()
	writeToFile("odd")
	concatSegments()

	generateSegments("concat.mp4")
	quarterSegments()
	writeToFile("all")
	concatSegments()

	makePreview("concat.mp4")
	cleanUp()
}
