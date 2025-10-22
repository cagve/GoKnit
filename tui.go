package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rivo/tview"
)


var session Session

type Session struct{
	Id   			int `json:"id"`
	Name   			string `json:"name"`
	Row 			int `json:"row"`
	LastModify		time.Time `json:"lastModify"`
}

func readSession(file string)  Session {
	jsonFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()
	
	byteValue, _ := io.ReadAll(jsonFile)
	var session Session

	json.Unmarshal(byteValue, &session)
	return session
}

func writeSession(session Session, file string){
	now := time.Now()
	session.LastModify = now

	jsonFile, err := json.Marshal(session)
	if err != nil {
		fmt.Println(err)
	}

	d1 := []byte(jsonFile)
	err = os.WriteFile(file, d1, 0644)
	if err != nil {
		fmt.Println(err)
	}

}


func listSessions() []string{
	sessionPath := "lib/"
	entries, err := os.ReadDir(sessionPath)
	if err != nil {
		fmt.Println(err)
	}
	var files[] string
	for _, f := range entries{
		if strings.HasSuffix(f.Name(), ".json"){
			files = append(files, f.Name())
		}
	}
	return files
}


func newSession(filename string) Session{
	sessions := listSessions()

	file := strings.ToLower(filename)
	file = strings.ReplaceAll(file, " ", "_")
	re := regexp.MustCompile(`[^a-z0-9_]`)
	file = re.ReplaceAllString(file, "")

	filepath := "lib/"+file+".json"

	s := Session{
		Id: len(sessions),
		Name: filename,
		Row: 0,
	}
	writeSession(s, filepath)
	return s
}



func app() {
    app := tview.NewApplication()

	// -------------- Session Info 
	idForm := tview.NewTextView().SetLabel("Id: ").SetText("")
	nameForm := tview.NewTextView().SetLabel("Name: ").SetText("")
	rowForm := tview.NewTextView().SetLabel("Row: ").SetText("")

	// -------------- Declaring widgets
	sessionInfoBox := tview.NewFlex()
	sessionInfoBox.SetDirection(tview.FlexRow).
		SetBorder(true).
		SetTitle("Knitting Session")

	actionList := tview.NewList()
	actionList.ShowSecondaryText(false).SetHighlightFullLine(true)

	patternBox := tview.NewBox().SetTitle("Pattern").SetBorder(true)

	left := tview.NewFlex().SetDirection(tview.FlexRow)
	placeHolder := tview.NewFlex().SetDirection(tview.FlexRow)

	// -------------- Updating info function
	updatePlaceHolder := func(widget tview.Primitive){
		placeHolder.Clear()
		placeHolder.AddItem(widget, 0, 1, true)
		app.SetFocus(widget)
	}

	updateInfo := func() {
		idForm.SetText(strconv.Itoa(session.Id))
		nameForm.SetText(session.Name)
		rowForm.SetText(strconv.Itoa(session.Row))
		updatePlaceHolder(sessionInfoBox)
		app.SetFocus(actionList)
    }
	updateInfo()


	// ------------- Adding commponents
	sessionInfoBox.AddItem(idForm, 1, 0, false).
		AddItem(nameForm, 1, 0, false).
		AddItem(rowForm, 1, 0, false)

	actionList.AddItem("open", "", 'o', func() {
			sessionList := tview.NewList().ShowSecondaryText(false)
			sessions := listSessions()

			for _, s := range sessions {
				char := string(s[0])
				runeArray := []rune(char)
				sessionList.InsertItem(-1, s, "", runeArray[0], func() {
					session = readSession("lib/"+s)
					updateInfo()
				})
			}
			sessionList.SetBorder(true)
			updatePlaceHolder(sessionList)
		}).
		AddItem("new", "", 'n', func(){
			form := tview.NewForm()
			form.AddInputField("Session name", "", 20, nil, nil).
				AddButton("Save", func(){
					name := form.GetFormItem(0).(*tview.InputField).GetText()
					session = newSession(name)
					updateInfo()
				}).
				AddButton("Quit", func() {
					app.Stop()
				})
			form.SetBorder(true).SetTitle("Creating file").SetTitleAlign(tview.AlignLeft)
			updatePlaceHolder(form)
		}).
		AddItem("save", "", 's', nil).
		AddItem("add", "", 'a', nil).
		AddItem("remove", "", 'r', nil).
		AddItem("close", "", 'q', func() {
			app.Stop()
		})
	

	
	left.AddItem(actionList, 0, 5, true).
		AddItem(placeHolder, 0, 5, false)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(left, 0, 3, true).
		AddItem(patternBox, 0, 7,  false)

		
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}


