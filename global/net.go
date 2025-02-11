package global

import (
	"bufio"
	"fmt"
	"github.com/Mrs4s/go-cqhttp/internal/base"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"

	"github.com/pkg/errors"

	"github.com/tidwall/gjson"

	"github.com/Mrs4s/go-cqhttp/internal/download"
)

var (
	client = &http.Client{
		Transport: &http.Transport{
			Proxy: func(request *http.Request) (u *url.URL, e error) {
				if base.Proxy == "" {
					return http.ProxyFromEnvironment(request)
				}
				return url.Parse(base.Proxy)
			},
			ForceAttemptHTTP2:   true,
			MaxConnsPerHost:     0,
			MaxIdleConns:        0,
			MaxIdleConnsPerHost: 999,
		},
	}

	// ErrOverSize 响应主体过大时返回此错误
	ErrOverSize = errors.New("oversize")

	// UserAgent HTTP请求时使用的UA
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66"
)

// DownloadFile 将给定URL对应的文件下载至给定Path
func DownloadFile(url, path string, limit int64, headers map[string]string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}
	defer file.Close()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if limit > 0 && resp.ContentLength > limit {
		return ErrOverSize
	}
	_, err = file.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// DownloadFileMultiThreading 使用threadCount个线程将给定URL对应的文件下载至给定Path
func DownloadFileMultiThreading(url, path string, limit int64, threadCount int, headers map[string]string) error {
	if threadCount < 2 {
		return DownloadFile(url, path, limit, headers)
	}
	type BlockMetaData struct {
		BeginOffset    int64
		EndOffset      int64
		DownloadedSize int64
	}
	var blocks []*BlockMetaData
	var contentLength int64
	errUnsupportedMultiThreading := errors.New("unsupported multi-threading")
	// 初始化分块或直接下载
	initOrDownload := func() error {
		copyStream := func(s io.ReadCloser) error {
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o666)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err = file.ReadFrom(s); err != nil {
				return err
			}
			return errUnsupportedMultiThreading
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}
		if _, ok := headers["User-Agent"]; !ok {
			req.Header["User-Agent"] = []string{UserAgent}
		}
		req.Header.Set("range", "bytes=0-")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errors.New("response status unsuccessful: " + strconv.FormatInt(int64(resp.StatusCode), 10))
		}
		if resp.StatusCode == 200 {
			if limit > 0 && resp.ContentLength > limit {
				return ErrOverSize
			}
			return copyStream(resp.Body)
		}
		if resp.StatusCode == 206 {
			contentLength = resp.ContentLength
			if limit > 0 && resp.ContentLength > limit {
				return ErrOverSize
			}
			blockSize := func() int64 {
				if contentLength > 1024*1024 {
					return (contentLength / int64(threadCount)) - 10
				}
				return contentLength
			}()
			if blockSize == contentLength {
				return copyStream(resp.Body)
			}
			var tmp int64
			for tmp+blockSize < contentLength {
				blocks = append(blocks, &BlockMetaData{
					BeginOffset: tmp,
					EndOffset:   tmp + blockSize - 1,
				})
				tmp += blockSize
			}
			blocks = append(blocks, &BlockMetaData{
				BeginOffset: tmp,
				EndOffset:   contentLength - 1,
			})
			return nil
		}
		return errors.New("unknown status code")
	}
	// 下载分块
	downloadBlock := func(block *BlockMetaData) error {
		req, _ := http.NewRequest("GET", url, nil)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o666)
		if err != nil {
			return err
		}
		defer file.Close()
		_, _ = file.Seek(block.BeginOffset, io.SeekStart)
		writer := bufio.NewWriter(file)
		defer writer.Flush()

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		if _, ok := headers["User-Agent"]; ok {
			req.Header["User-Agent"] = []string{UserAgent}
		}
		req.Header.Set("range", "bytes="+strconv.FormatInt(block.BeginOffset, 10)+"-"+strconv.FormatInt(block.EndOffset, 10))
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errors.New("response status unsuccessful: " + strconv.FormatInt(int64(resp.StatusCode), 10))
		}
		buffer := make([]byte, 1024)
		i, err := resp.Body.Read(buffer)
		for {
			if err != nil && err != io.EOF {
				return err
			}
			i64 := int64(len(buffer[:i]))
			needSize := block.EndOffset + 1 - block.BeginOffset
			if i64 > needSize {
				i64 = needSize
				err = io.EOF
			}
			_, e := writer.Write(buffer[:i64])
			if e != nil {
				return e
			}
			block.BeginOffset += i64
			block.DownloadedSize += i64
			if err == io.EOF || block.BeginOffset > block.EndOffset {
				break
			}
			i, err = resp.Body.Read(buffer)
		}
		return nil
	}

	if err := initOrDownload(); err != nil {
		if err == errUnsupportedMultiThreading {
			return nil
		}
		return err
	}
	wg := sync.WaitGroup{}
	wg.Add(len(blocks))
	var lastErr error
	for i := range blocks {
		go func(b *BlockMetaData) {
			defer wg.Done()
			if err := downloadBlock(b); err != nil {
				lastErr = err
			}
		}(blocks[i])
	}
	wg.Wait()
	return lastErr
}

// QQMusicSongInfo 通过给定id在QQ音乐上查找曲目信息
func QQMusicSongInfo(id string) (gjson.Result, error) {
	d, err := download.Request{URL: `https://u.y.qq.com/cgi-bin/musicu.fcg?format=json&inCharset=utf8&outCharset=utf-8&notice=0&platform=yqq.json&needNewCode=0&data={%22comm%22:{%22ct%22:24,%22cv%22:0},%22songinfo%22:{%22method%22:%22get_song_detail_yqq%22,%22param%22:{%22song_type%22:0,%22song_mid%22:%22%22,%22song_id%22:` + id + `},%22module%22:%22music.pf_song_detail_svr%22}}`}.JSON()
	if err != nil {
		return gjson.Result{}, err
	}
	return d.Get("songinfo.data"), nil
}

// NeteaseMusicSongInfo 通过给定id在wdd音乐上查找曲目信息
func NeteaseMusicSongInfo(id string) (gjson.Result, error) {
	d, err := download.Request{URL: fmt.Sprintf("http://music.163.com/api/song/detail/?id=%s&ids=%%5B%s%%5D", id, id)}.JSON()
	if err != nil {
		return gjson.Result{}, err
	}
	return d.Get("songs.0"), nil
}
