package main

import (
	// built-in
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	// 3rd-parties
	"github.com/fatih/color"
	"github.com/rishavxyz/makaut-notice-reader/utils"
)

const url = "https://makaut1.ucanapply.com/smartexam/public/api/notice-data"

type MakautJSONData struct {
	Data []struct {
		FilePath    string `json:"file_path"`
		NoticeTitle string `json:"notice_title"`
		NoticeDate  string `json:"notice_date"`
		CreatedAt   string `json:"created_at"`
	} `json:"data"`
	Timestamp int64 `json:"timestamp"`
}

func PrintToConsole(data *MakautJSONData) {
	var v = data.Data

	var grey = color.New(color.FgHiBlack).Add(color.Bold)
	var bold = color.New().Add(color.Bold)
	var green = color.New(color.FgGreen).Add(color.Bold)

	for i := len(v) - 1; i >= 0; i-- {
		fmt.Println()
		if i < 2 {
			grey.Printf("%v. NEW\t", i+1)
		} else {
			grey.Printf("%v.\t", i+1)
		}
		bold.Printf("%v\n", v[i].NoticeTitle)
		grey.Printf("\tOn %v\n", utils.Fdate(v[i].NoticeDate))
		green.Print("\tPDF link: ")
		grey.Print(v[i].FilePath)
		fmt.Println()
	}
}

func GetPdfId() int {
	fmt.Print("\nWhich one I download? [0 - cancel, 1..58] ")
	var input string
	fmt.Scan(&input)

	num, err := strconv.Atoi(input)

	if err != nil {
		color.Yellow("Only accept numbers")
		os.Exit(1)
	}
	if num == 0 {
		color.Yellow("Nothing to do")
		os.Exit(0)
	}
	return num
}

func main() {
	if utils.IsOffline() {
		color.Red("### You are offline ###")
		return
	}
	res, err := http.Get(url)

	if err != nil {
		color.Red("Error while fetching data:", err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		color.Red("Error:", err)
		return
	}
	var jsonData MakautJSONData

	if err := json.Unmarshal(body, &jsonData); err != nil {
		color.Red("Error:", err)
		return
	}

	PrintToConsole(&jsonData)

	pdfId := GetPdfId()

	filePath := jsonData.Data[pdfId-1].FilePath
	homeDir, _ := os.UserHomeDir()
	fname := homeDir + "/makaut-notice-" + strconv.Itoa(pdfId) + ".pdf"

	out, err := os.Create(fname)

	if err != nil {
		color.Red("Error:", err)
		return
	}
	defer out.Close()

	res, httpErr := http.Get(filePath)

	if httpErr != nil {
		color.Red("Error:", err)
		return
	}
	defer res.Body.Close()

	if _, err := io.Copy(out, res.Body); err != nil {
		color.Red("Error while saving file:", err)
		return
	}
	color.Green("file saved in %v", fname)
}
