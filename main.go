package main

import (
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prologic/bitcask"
	"github.com/urfave/cli"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var dataDir string
var dbDir string
var ipRestriction string

var DB *bitcask.Bitcask

func upload(c echo.Context) error {
	// Read form fields
	name := c.FormValue("name")
	email := c.FormValue("email")
	ip := c.FormValue("IP")

	log.Println(name, ip)

	candi, err := getDB(ip)
	if ipRestriction == "true" {
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

func InitFlags() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "dbDir",
			Usage:       "db dir path",
			Destination: &dbDir,
			EnvVar:      "DB_DIR",
			Value:       "db/",
		},
		cli.StringFlag{
			Name:        "dataDir",
			Usage:       "data dir path",
			Destination: &dataDir,
			EnvVar:      "DATA_DIR",
			Value:       "data/",
		},
		cli.StringFlag{
			Name:        "ipRestriction",
			Usage:       "enable ip restriction",
			Destination: &ipRestriction,
			EnvVar:      "IP_REST",
			Value:       "true",
		},
	}
	app.Action = func(c *cli.Context) error {
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
	}
}

func getshah(in string) string {
	h := sha1.New()
	h.Write([]byte(in))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
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

	// Create a csv file
	f, err := os.Create(dataDir + "/people.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	// Write Unmarshaled json data to CSV file
	w := csv.NewWriter(f)

	for key := range DB.Keys() {
		log.Println(key)
		data, _ := DB.Get(key)
		candi := candidate{}
		json.Unmarshal(data, &candi)

		var record []string
		record = append(record, candi.Name)
		record = append(record, candi.Email)
		record = append(record, candi.IPAddr)
		record = append(record, candi.Filename)
		w.Write(record)

		log.Println(candi)
	}

	w.Flush()
}

type candidate struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	IPAddr   string `json:"ip_addr"`
	Filename string `json:"filename"`
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {

	InitFlags()
	//dbFile = "db/db"

	log.Println("data dir:", dataDir)
	log.Println("db dir:", dbDir)
	log.Println("ip restriction:", ipRestriction)

	var err error
	DB, err = bitcask.Open(dbDir)

	if err != nil {
		log.Println(err)
	}

	log.Println(len(os.Args))
	if len(os.Args) == 2 {
		arg := os.Args[1]
		if arg == "csv" {
			readAll()
			return
		}

	}

	//readAll()

	defer DB.Close()

	//dataDir = "data/"

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
