package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prologic/bitcask"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
)

var dataDir string
var dbFile string

var DB *bitcask.Bitcask

func upload(c echo.Context) error {
	// Read form fields
	name := c.FormValue("name")
	email := c.FormValue("email")
	ip := c.FormValue("IP")

	fmt.Println(name, ip)

	candi, err := getDB(ip)

	if err == nil {
		//return c.HTML(http.StatusOK, fmt.Sprintf("<p>File already submitted with fields </br>name=%s </br>email=%s.</p>", name, email))

		return c.Render(http.StatusOK, "upload.html", struct {
			Candidate candidate
			Status    string
		}{
			candi,
			"File already submitted ",
		})
	}

	//-----------
	// Read file
	//-----------

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	hash := getshah(name + email)

	name_parts := strings.Split(file.Filename, ".")
	ext := "." + name_parts[len(name_parts)-1]
	if len(name_parts) == 1 {
		ext = ""
	}
	// Destination
	dst_path := dataDir + "/" + hash + ext
	dst, err := os.Create(dst_path)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	candi = candidate{
		Name:     name,
		Email:    email,
		IPAddr:   ip,
		Filename: hash + ext,
	}

	saveDB(candi)

	return c.Render(http.StatusOK, "upload.html", struct {
		Candidate candidate
		Status    string
	}{
		candi,
		"File uploaded successfully",
	})

	//return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s uploaded successfully with fields </br>name=%s </br>email=%s.</p>", file.Filename, name, email))
}

func getshah(in string) string {
	h := sha1.New()
	h.Write([]byte(in))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {

	dbFile = "db/db"

	var err error
	DB, err = bitcask.Open(dbFile)

	if err != nil {
		fmt.Println(err)
	}

	readAll()

	defer DB.Close()

	dataDir = "data/"

	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e := echo.New()
	e.Renderer = t
	e.Use(middleware.BodyLimit("1M"))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/upload", upload)
	e.Logger.Fatal(e.Start(":1323"))

}

func saveDB(in candidate) {
	out, _ := json.Marshal(in)
	DB.Put(in.IPAddr, out)
}

func getDB(ip string) (candidate, error) {
	data, err := DB.Get(ip)
	if err != nil {
		fmt.Println(err)
		return candidate{}, err
	}
	out := candidate{}
	json.Unmarshal(data, &out)
	return out, nil
}

func readAll() {
	for key := range DB.Keys() {
		fmt.Println(key)
		data, _ := DB.Get(key)
		candi := candidate{}
		json.Unmarshal(data, &candi)
		fmt.Println(candi)
	}
}

type candidate struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	IPAddr   string `json:"ip_addr"`
	Filename string `json:"filename"`
}
