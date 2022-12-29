package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Segment struct {
	Duration float32
	RelUrl   string
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func cors(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w, r)
	}
}

func getSegments(path string) ([]Segment, error) {

	var segment_list = make([]Segment, 0)
	f, err := os.Open(path)
	if err != nil {
		return segment_list, fmt.Errorf("getSegments: could not open file: %w", err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	line := 1

	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#EXTINF:") {
			var newSegment Segment
			dur, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(scanner.Text(), "#EXTINF:"), ","), 32)
			if err != nil {
				return segment_list, fmt.Errorf("getSegments: could not parse #EXTINF: %w", err)
			}

			scanner.Scan()
			if err := scanner.Err(); err != nil {
				return segment_list, fmt.Errorf("getSegments: %w", err)
			}
			RelUrl := scanner.Text()

			newSegment.Duration = float32(dur)
			newSegment.RelUrl = RelUrl

			segment_list = append(segment_list, newSegment)
		}
		line++
	}

	if err = scanner.Err(); err != nil {
		return segment_list, fmt.Errorf("getSegments: %w", err)
	}

	return segment_list, nil
}

func updatePlaylist(segment_list []Segment, path string) error {

	out_path := filepath.Dir(path) + "/output.m3u8"
	media_seq := 0
	for media_seq < len(segment_list) {
		f, err := os.Create(out_path)
		if err != nil {
			return fmt.Errorf("updatePlaylist: could not create file: %w", err)
		}
		w := bufio.NewWriter(f)
		w.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:13\n#EXT-X-MEDIA-SEQUENCE:" + strconv.Itoa(media_seq) + "\n")
		for i := media_seq; i < min(media_seq+4, len(segment_list)); i++ {
			w.WriteString("#EXTINF:" + strconv.FormatFloat(float64(segment_list[i].Duration), 'f', -1, 32) + ",\n" + segment_list[i].RelUrl + "\n")
		}
		w.Flush()
		f.Close()
		c1 := time.NewTimer(time.Duration(segment_list[media_seq].Duration * float32(time.Second)))

		<-c1.C
		media_seq++
	}

	return nil
}

func streamVideo(path string) {
	segment_list, err := getSegments(path)
	if err != nil {
		fmt.Println(err)
	}
	err = updatePlaylist(segment_list, path)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	go streamVideo("C:/Users/vikra/go-video/video/tenet/file.m3u8")

	videoFs := http.FileServer(http.Dir("/Users/vikra/go-video"))

	http.Handle("/video/", cors(videoFs))

	http.ListenAndServe(":9000", nil)
}
