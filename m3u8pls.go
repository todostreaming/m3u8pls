package m3u8pls

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type M3U8pls struct {
	m3u8      string
	Targetdur float64
	Mediaseq  int64
	Segment   []string
	Duration  []float64
	Ok        bool // the .m3u8 playlist is reachable and has segments
	mu_pls    sync.Mutex
}

func M3U8playlist(m3u8 string) *M3U8pls {
	m3u := &M3U8pls{}
	m3u.mu_pls.Lock()
	defer m3u.mu_pls.Unlock()

	m3u.m3u8 = m3u8

	return m3u
}

func (m *M3U8pls) Parse() {
	m.mu_pls.Lock()
	m.Targetdur = 0.0
	m.Mediaseq = 0
	m.Segment = []string{}
	m.Duration = []float64{}
	m.Ok = false
	m.mu_pls.Unlock()

	m.analyzem3u8()
}

func (m *M3U8pls) analyzem3u8() {
	substr := ""
	issubstr := false
	m.mu_pls.Lock()
	m3u8 := m.m3u8
	m.mu_pls.Unlock()
	resp, err := http.Get(m3u8)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\n")
		if strings.Contains(line, ".m3u8") {
			substr = substream(m3u8, line)
			issubstr = true
			break
		}
		if strings.Contains(line, "#EXT-X-TARGETDURATION:") {
			var targetdur float64
			fmt.Sscanf(line, "#EXT-X-TARGETDURATION:%f", &targetdur)
			if targetdur > 12 {
				targetdur = 12.0
			}
			m.mu_pls.Lock()
			m.Targetdur = targetdur
			m.mu_pls.Unlock()
		}
		if strings.Contains(line, "#EXT-X-MEDIA-SEQUENCE:") {
			var mediaseq int64
			fmt.Sscanf(line, "#EXT-X-MEDIA-SEQUENCE:%d", &mediaseq)
			m.mu_pls.Lock()
			m.Mediaseq = mediaseq
			m.mu_pls.Unlock()
		}
		if strings.Contains(line, "#EXTINF:") {
			var extinf float64
			fmt.Sscanf(line, "#EXTINF:%f,", &extinf)
			m.mu_pls.Lock()
			m.Duration = append(m.Duration, extinf)
			m.mu_pls.Unlock()
		}
		if strings.Contains(line, ".ts") {
			m.mu_pls.Lock()
			m.Segment = append(m.Segment, substream(m3u8, line))
			m.Ok = true
			m.mu_pls.Unlock()
		}
		//fmt.Printf("1)=>[%s]<=\n",line)
	}
	resp.Body.Close()
	if issubstr {
		resp, err := http.Get(substr)
		if err != nil {
			return
		}
		if resp.StatusCode != 200 {
			return
		}
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimRight(line, "\n")
			if strings.Contains(line, "#EXT-X-TARGETDURATION:") {
				var targetdur float64
				fmt.Sscanf(line, "#EXT-X-TARGETDURATION:%f", &targetdur)
				if targetdur > 12 {
					targetdur = 12.0
				}
				m.mu_pls.Lock()
				m.Targetdur = targetdur
				m.mu_pls.Unlock()
			}
			if strings.Contains(line, "#EXT-X-MEDIA-SEQUENCE:") {
				var mediaseq int64
				fmt.Sscanf(line, "#EXT-X-MEDIA-SEQUENCE:%d", &mediaseq)
				m.mu_pls.Lock()
				m.Mediaseq = mediaseq
				m.mu_pls.Unlock()
			}
			if strings.Contains(line, "#EXTINF:") {
				var extinf float64
				fmt.Sscanf(line, "#EXTINF:%f,", &extinf)
				m.mu_pls.Lock()
				m.Duration = append(m.Duration, extinf)
				m.mu_pls.Unlock()
			}
			if strings.Contains(line, ".ts") {
				m.mu_pls.Lock()
				m.Segment = append(m.Segment, substream(substr, line))
				m.Ok = true
				m.mu_pls.Unlock()
			}
			//fmt.Printf("2)=>[%s]<=\n",line)
		}
		resp.Body.Close()
	}
}

func substream(m3u8, sub string) string {
	var substream string
	is_extra := false
	var extra string

	// extra = ?whatever after the base url (authentication, etc)
	if strings.Contains(m3u8, "?") {
		is_extra = true
		p := strings.Split(m3u8, "?")
		m3u8 = p[0]
		extra = p[1]
	}

	m3u8 = m3u8[7:] // quito http://
	substream = "http://"
	parts := strings.Split(m3u8, "/")
	for _, v := range parts {
		if strings.Contains(v, ".m3u8") {
			substream = substream + sub
			break
		} else {
			substream = substream + v + "/"
		}
	}

	if is_extra {
		substream = substream + "?" + extra
	}

	return substream
}
