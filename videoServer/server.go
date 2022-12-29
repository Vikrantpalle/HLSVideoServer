package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Segment struct {
	Duration float32
	RelUrl string
}

func min(a,b int) int {
	if a<b {
		return a
	}
	return b
}

func cors (fs http.Handler) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request)  {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w,r)
	}
}

var segment_list = make([]Segment, 0)

func getSegments() {
	f,err := os.Open("C:/Users/vikra/go-video/video/tenet/file.m3u8")
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	line :=1



	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#EXTINF:") {
			var newSegment Segment
			dur,err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(scanner.Text(), "#EXTINF:"), ","), 32)
			if err!=nil {
				fmt.Println(err)
			}
			
			scanner.Scan()
			if err := scanner.Err();err!=nil {
				fmt.Println(err)
			}
			RelUrl := scanner.Text()

			newSegment.Duration = float32(dur)
			newSegment.RelUrl = RelUrl

			segment_list = append(segment_list, newSegment)
		}
		line++
	}

	if err = scanner.Err(); err!=nil {
		fmt.Println(err)
	}

}

func updatePlaylist(){
	
	media_seq := 0
	for media_seq < len(segment_list){
		f,err := os.Create("C:/Users/vikra/go-video/video/tenet/output.m3u8")
		if err !=nil {
			fmt.Println(err)
		}
		w := bufio.NewWriter(f)
	w.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:13\n#EXT-X-MEDIA-SEQUENCE:"+strconv.Itoa(media_seq)+"\n")	
	for i:= media_seq; i < min(media_seq+4, len(segment_list));i++ {
		w.WriteString("#EXTINF:"+strconv.FormatFloat(float64(segment_list[i].Duration),'f',-1,32)+",\n"+segment_list[i].RelUrl+"\n")
	}
	w.Flush()
	f.Close()
	c1 := time.NewTimer( time.Duration(segment_list[media_seq].Duration * float32(time.Second)))

	<-c1.C
	media_seq++
	}
}


func main() {
	getSegments()
	go updatePlaylist()

	videoFs := http.FileServer(http.Dir("/Users/vikra/go-video"))

	http.Handle("/video/", cors(videoFs))

	http.ListenAndServe(":9000", nil)
}