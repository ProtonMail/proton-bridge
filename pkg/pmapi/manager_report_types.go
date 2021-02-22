package pmapi

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
)

// ClientType is required by API.
const (
	EmailClientType = iota + 1
	VPNClientType
)

type reportAtt struct {
	name, filename string
	body           io.Reader
}

// ReportBugReq stores data for report.
type ReportBugReq struct {
	OS                string      `json:",omitempty"`
	OSVersion         string      `json:",omitempty"`
	Browser           string      `json:",omitempty"`
	BrowserVersion    string      `json:",omitempty"`
	BrowserExtensions string      `json:",omitempty"`
	Resolution        string      `json:",omitempty"`
	DisplayMode       string      `json:",omitempty"`
	Client            string      `json:",omitempty"`
	ClientVersion     string      `json:",omitempty"`
	ClientType        int         `json:",omitempty"`
	Title             string      `json:",omitempty"`
	Description       string      `json:",omitempty"`
	Username          string      `json:",omitempty"`
	Email             string      `json:",omitempty"`
	Country           string      `json:",omitempty"`
	ISP               string      `json:",omitempty"`
	Debug             string      `json:",omitempty"`
	Attachments       []reportAtt `json:",omitempty"`
}

// AddAttachment to report.
func (rep *ReportBugReq) AddAttachment(name, filename string, r io.Reader) {
	rep.Attachments = append(rep.Attachments, reportAtt{name: name, filename: filename, body: r})
}

func writeMultipartReport(w *multipart.Writer, rep *ReportBugReq) error { // nolint[funlen]
	fieldData := map[string]string{
		"OS":                rep.OS,
		"OSVersion":         rep.OSVersion,
		"Browser":           rep.Browser,
		"BrowserVersion":    rep.BrowserVersion,
		"BrowserExtensions": rep.BrowserExtensions,
		"Resolution":        rep.Resolution,
		"DisplayMode":       rep.DisplayMode,
		"Client":            rep.Client,
		"ClientVersion":     rep.ClientVersion,
		"ClientType":        "1",
		"Title":             rep.Title,
		"Description":       rep.Description,
		"Username":          rep.Username,
		"Email":             rep.Email,
		"Country":           rep.Country,
		"ISP":               rep.ISP,
		"Debug":             rep.Debug,
	}

	for field, data := range fieldData {
		if data == "" {
			continue
		}
		if err := w.WriteField(field, data); err != nil {
			return err
		}
	}

	quoteEscaper := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

	for _, att := range rep.Attachments {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				quoteEscaper.Replace(att.name), quoteEscaper.Replace(att.filename+".zip")))
		h.Set("Content-Type", "application/octet-stream")
		// h.Set("Content-Transfer-Encoding", "base64")
		attWr, err := w.CreatePart(h)
		if err != nil {
			return err
		}

		zipArch := zip.NewWriter(attWr)
		zipWr, err := zipArch.Create(att.filename)
		// b64 := base64.NewEncoder(base64.StdEncoding, zipWr)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipWr, att.body)
		if err != nil {
			return err
		}
		err = zipArch.Close()
		// err = b64.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
