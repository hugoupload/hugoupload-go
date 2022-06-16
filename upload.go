package hugoPartUpload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hyahm/golog"
)

func (pc *PartClient) checkFiled() error {
	if pc.Domain == "" {
		pc.Domain = "http://admin.hugocut.com"
	}

	if pc.Filename == "" {
		return errors.New("filename not be empty")
	}
	if pc.User == "" {
		return errors.New("user not be empty")
	}

	if pc.Identifier == "" {
		return errors.New("identifier not be empty")
	}

	if pc.Token == "" {
		return errors.New("token not be empty")
	}

	if pc.Title == "" {
		pc.Title = pc.Rule
	}

	if pc.Rule == "" {
		return errors.New("rule not be empty")
	}

	if pc.Cat == "" {
		return errors.New("cat not be empty")
	}
	if pc.Domain[len(pc.Domain)-1:] == "/" {
		pc.Domain = pc.Domain[:len(pc.Domain)-1]
	}
	if pc.NewFilename == "" {
		i := strings.LastIndex(pc.Filename, ".")
		pc.NewFilename = pc.Identifier + pc.Filename[i:]
	}
	return nil
}

func (pc *PartClient) PartUpload() error {
	err := pc.checkFiled()
	if err != nil {
		golog.Error(err)
		return err
	}
	err = pc.initfunc()
	if err != nil {
		golog.Error(err)
		return err
	}
	return pc.dataForm(nil)
}

func (pc *PartClient) Upload() error {
	err := pc.checkFiled()
	if err != nil {
		return err
	}
	return pc.upload()
}

var PARTSIZE int = 30 << 20 // 10M

func (pc *PartClient) initfunc() error {
	x := `
	{
		"fileName": "%s",
		"totalParts": %d,
		"totalSize": %d,
		"user": "%s"
	}
	`
	f, err := os.Open(pc.Filename)
	if err != nil {
		golog.Error(err)
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		golog.Error(err)
		return err
	}
	pc.tp = int(fi.Size() / int64(PARTSIZE))
	if fi.Size()%int64(PARTSIZE) != 0 {
		pc.tp++
	}
	x = fmt.Sprintf(x, pc.NewFilename, pc.tp, fi.Size(), pc.User)
	golog.Info(x)
	cli := &http.Client{}

	r, err := http.NewRequest("POST", pc.Domain+"/init", strings.NewReader(x))
	if err != nil {
		golog.Error(err)
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Token", pc.Token)
	resp, err := cli.Do(r)
	if err != nil {
		golog.Error(err)
		return err
	}

	defer resp.Body.Close()
	init := &initData{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		golog.Error(err)
		return err
	}
	golog.Info(string(b))
	err = json.Unmarshal(b, init)
	if err != nil {
		golog.Error(err)
		return err
	}
	if init.Code != 0 {
		return errors.New(init.Message)
	}
	pc.UploadId = init.Data.UploadId
	return nil
}

func (pc *PartClient) dataForm(miss []int) error {
	// 切片上传
	f, err := os.Open(pc.Filename)
	// b, err := ioutil.ReadFile(pc.Filename)
	if err != nil {
		golog.Error(err)
		return err
	}
	defer f.Close()
	wg := &sync.WaitGroup{}
	for i := 0; i < pc.tp; i++ {
		buf := &bytes.Buffer{}
		w := multipart.NewWriter(buf)

		part, err := w.CreateFormFile("file", fmt.Sprintf("%d%s", i, filepath.Ext(pc.Filename)))
		if err != nil {
			golog.Error(err)
			return err
		}

		_, err = f.Seek(int64(i*PARTSIZE), 0)
		if err != nil {
			golog.Error(err)
			return err
		}
		b := make([]byte, PARTSIZE)
		n, err := f.Read(b)
		if err != nil {
			if err != io.EOF {
				golog.Error(err)
				return err
			} else {
				break
			}
		}
		_, err = part.Write(b[:n])
		if err != nil {
			golog.Info(err)
			return err
		}
		w.WriteField("partNumber", fmt.Sprintf("%d", i+1))
		w.WriteField("uploadId", fmt.Sprintf("%d", pc.UploadId))
		w.WriteField("user", pc.User)
		w.Close()

		wg.Add(1)
		go pc.cut(w.FormDataContentType(), buf, wg)
	}
	wg.Wait()
	return pc.complate()
}

func (pc *PartClient) cut(typ string, buf *bytes.Buffer, wg *sync.WaitGroup) {
	defer wg.Done()
	req, err := http.NewRequest("POST", pc.Domain+"/upload", buf)
	if err != nil {
		golog.Error(err)
		return
	}
	req.Header.Set("Content-Type", typ)
	req.Header.Set("Token", pc.Token)
	req.Header.Set("upload_id", fmt.Sprintf("%d", pc.UploadId))
	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		golog.Error(err)
		return
	}
	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		golog.Error(err)
		return
	}

	golog.Info(string(rb))
}

func (pc *PartClient) complate() error {
	cli := http.Client{}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	if pc.Cover != "" {
		imageb, err := ioutil.ReadFile(pc.Cover)
		if err != nil {
			golog.Error(err)
			return err
		}
		image, err := w.CreateFormFile("image", fmt.Sprintf("%s.jpg", pc.Identifier))
		if err != nil {
			golog.Error(err)
			return err
		}
		_, err = io.Copy(image, bytes.NewReader(imageb))
		if err != nil {
			return err
		}
	}
	golog.Info(pc.User)
	w.WriteField("uploadId", fmt.Sprintf("%d", pc.UploadId))
	w.WriteField("user", pc.User)
	w.WriteField("identifier", pc.Identifier)
	w.WriteField("title", pc.Title)
	w.WriteField("rule", pc.Rule)
	w.WriteField("cat", pc.Cat)
	w.WriteField("unaudit", fmt.Sprintf("%d", pc.UnAudit))
	w.WriteField("ftp_user_id", fmt.Sprintf("%d", pc.FtpUserId))
	w.WriteField("subcat", strings.Join(pc.Subcat, ","))
	w.WriteField("actor", pc.Actor)
	w.Close()
	req, err := http.NewRequest("POST", pc.Domain+"/complete", buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Token", pc.Token)
	req.Header.Set("upload_id", fmt.Sprintf("%d", pc.UploadId))
	resp, err := cli.Do(req)
	if err != nil {
		return err

	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err

	}
	golog.Info(string(b))
	cd := &completeData{}
	err = json.Unmarshal(b, cd)
	if err != nil {
		return err

	}
	if cd.Code == 2 {
		pc.dataForm(cd.Data)
		return pc.complate()
	}
	golog.Info(string(b))
	return nil
}

func (pc *PartClient) upload() error {
	if pc.Token == "" {
		return errors.New("token not be empty")
	}
	if pc.Audio == "" {
		return errors.New("audio not be empty")
	}
	cli := http.Client{}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)

	videob, err := ioutil.ReadFile(pc.Audio)
	if err != nil {
		return err
	}
	audio, err := w.CreateFormFile("audio", pc.Filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(audio, bytes.NewReader(videob))
	if err != nil {
		return err
	}
	imageb, err := ioutil.ReadFile(pc.Cover)
	if err != nil {
		return err
	}
	image, err := w.CreateFormFile("image", fmt.Sprintf("%s.jpg", pc.Identifier))
	if err != nil {
		return err
	}

	_, err = io.Copy(image, bytes.NewReader(imageb))
	if err != nil {
		return err
	}
	w.WriteField("uploadId", fmt.Sprintf("%d", pc.UploadId))
	w.WriteField("user", pc.User)
	w.WriteField("identifier", pc.Identifier)
	w.WriteField("title", pc.Title)
	w.WriteField("rule", pc.Rule)
	w.WriteField("cat", pc.Cat)
	w.WriteField("subcat", strings.Join(pc.Subcat, ","))
	w.WriteField("actor", pc.Actor)
	w.WriteField("filename", pc.Filename)

	req, err := http.NewRequest("POST", pc.Domain+"/upload", buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Token", pc.Token)
	resp, err := cli.Do(req)
	if err != nil {
		return err

	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err

	}
	fmt.Println(string(b))
	return nil
}
