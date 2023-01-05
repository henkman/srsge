package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

type SaveGame struct {
	Year    uint16
	Rubles  float32
	Dollars float32
}

const (
	DOLLARS_POS = 388
	RUBLES_POS  = 392
	YEAR_POS    = 412
)

func (sg *SaveGame) Read(in io.ReadSeeker) error {
	var buf [4]byte

	in.Seek(DOLLARS_POS, io.SeekStart)
	if _, err := in.Read(buf[:]); err != nil {
		return err
	}
	sg.Dollars = math.Float32frombits(binary.LittleEndian.Uint32(buf[:]))

	in.Seek(RUBLES_POS, io.SeekStart)
	if _, err := in.Read(buf[:]); err != nil {
		return err
	}
	sg.Rubles = math.Float32frombits(binary.LittleEndian.Uint32(buf[:]))

	in.Seek(YEAR_POS, io.SeekStart)
	if _, err := in.Read(buf[:2]); err != nil {
		return err
	}
	sg.Year = binary.LittleEndian.Uint16(buf[:2])

	return nil
}

func (sg *SaveGame) ReadFile(filepath string) error {
	fd, err := os.Open(filepath)
	if err != nil {
		return err
	}
	err = sg.Read(fd)
	fd.Close()
	return err
}

func (sg *SaveGame) WriteTo(out io.WriteSeeker) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(sg.Dollars))
	if _, err := out.Seek(DOLLARS_POS, io.SeekStart); err != nil {
		return err
	}
	out.Write(buf[:])

	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(sg.Rubles))
	if _, err := out.Seek(RUBLES_POS, io.SeekStart); err != nil {
		return err
	}
	out.Write(buf[:])

	binary.LittleEndian.PutUint16(buf[:], sg.Year)
	if _, err := out.Seek(YEAR_POS, io.SeekStart); err != nil {
		return err
	}
	out.Write(buf[:2])
	return nil
}

func (sg *SaveGame) WriteToFile(filepath string) error {
	fd, err := os.OpenFile(filepath, os.O_WRONLY, 0750)
	if err != nil {
		return err
	}
	err = sg.WriteTo(fd)
	fd.Close()
	return err
}

type SaveGameEditor struct {
	SaveFolder  string
	LoadButton  *walk.PushButton
	SaveButton  *walk.PushButton
	YearEdit    *walk.NumberEdit
	RublesEdit  *walk.NumberEdit
	DollarsEdit *walk.NumberEdit
}

func main() {
	var sge SaveGameEditor
	var mw *walk.MainWindow
	var appIcon, _ = walk.NewIconFromResourceId(2)
	if err := (MainWindow{
		Icon:     appIcon,
		AssignTo: &mw,
		Title:    "SRSGE",
		Size: Size{
			Width:  200,
			Height: 140,
		},
		Layout: VBox{
			MarginsZero: true,
		},
		Children: []Widget{
			Composite{
				Layout: Grid{
					Columns: 2,
					Margins: Margins{
						Left:   5,
						Top:    5,
						Right:  0,
						Bottom: 0,
					},
				},
				Children: []Widget{
					Label{
						Text:    "Year",
						MaxSize: Size{40, 20},
					},
					NumberEdit{
						Enabled:            false,
						AssignTo:           &sge.YearEdit,
						ToolTipText:        "Year",
						MinValue:           0.0,
						MaxValue:           9999.0,
						SpinButtonsVisible: true,
						MinSize:            Size{120, 20},
						MaxSize:            Size{120, 20},
					},
				},
			},
			Composite{
				Layout: Grid{
					Columns: 2,
					Margins: Margins{
						Left:   5,
						Top:    5,
						Right:  0,
						Bottom: 0,
					},
				},
				Children: []Widget{
					Label{
						Text:    "Rubles",
						MaxSize: Size{40, 20},
					},
					NumberEdit{
						Enabled:     false,
						Decimals:    4,
						AssignTo:    &sge.RublesEdit,
						ToolTipText: "Rubles",
						MinSize:     Size{120, 20},
						MaxSize:     Size{120, 20},
					},
				},
			},
			Composite{
				Layout: Grid{
					Columns: 2,
					Margins: Margins{
						Left:   5,
						Top:    5,
						Right:  0,
						Bottom: 0,
					},
				},
				Children: []Widget{
					Label{
						Text:    "Dollars",
						MaxSize: Size{40, 20},
					},
					NumberEdit{
						Enabled:     false,
						Decimals:    4,
						AssignTo:    &sge.DollarsEdit,
						ToolTipText: "Dollars",
						MinSize:     Size{120, 20},
						MaxSize:     Size{120, 20},
					},
				},
			},
			Composite{
				Layout: Grid{
					Columns: 2,
					Margins: Margins{
						Left:   60,
						Top:    5,
						Right:  0,
						Bottom: 5,
					},
				},
				Children: []Widget{
					PushButton{
						AssignTo: &sge.LoadButton,
						Text:     "Load",
						MaxSize:  Size{40, 20},
						OnClicked: func() {
							var fd walk.FileDialog
							fd.Title = "Select Save Directory"
							const DEFAULT_SAVE_FOLDER = `C:\Program Files (x86)\Steam\steamapps\common\SovietRepublic\media_soviet\save`
							if _, err := os.Stat(DEFAULT_SAVE_FOLDER); err == nil {
								fd.InitialDirPath = DEFAULT_SAVE_FOLDER
							}
							if accept, err := fd.ShowBrowseFolder(mw); !accept || err != nil {
								return
							}
							sge.SaveFolder = fd.FilePath
							var sg SaveGame
							if err := sg.ReadFile(filepath.Join(sge.SaveFolder, "header.bin")); err != nil {
								walk.MsgBox(mw, "Error reading",
									fmt.Sprint("Could not read save: ", err.Error()),
									walk.MsgBoxOK|walk.MsgBoxIconError)
								return
							}
							sge.YearEdit.SetValue(float64(sg.Year))
							sge.YearEdit.SetEnabled(true)
							sge.RublesEdit.SetValue(float64(sg.Rubles))
							sge.RublesEdit.SetEnabled(true)
							sge.DollarsEdit.SetValue(float64(sg.Dollars))
							sge.DollarsEdit.SetEnabled(true)
							sge.SaveButton.SetEnabled(true)
						},
					},
					PushButton{
						AssignTo: &sge.SaveButton,
						Enabled:  false,
						Text:     "Save",
						MaxSize:  Size{40, 20},
						OnClicked: func() {
							if sge.SaveFolder == "" {
								walk.MsgBox(mw, "Info", "Load a savegame first",
									walk.MsgBoxOK|walk.MsgBoxIconInformation)
								return
							}
							var sg SaveGame
							sg.Year = uint16(sge.YearEdit.Value())
							sg.Rubles = float32(sge.RublesEdit.Value())
							sg.Dollars = float32(sge.DollarsEdit.Value())
							if err := sg.WriteToFile(filepath.Join(sge.SaveFolder, "header.bin")); err != nil {
								walk.MsgBox(mw, "Error writing",
									fmt.Sprint("Could not write save: ", err.Error()),
									walk.MsgBoxOK|walk.MsgBoxIconError)
								return
							}
							sge.SaveFolder = ""
							sge.YearEdit.SetValue(0)
							sge.YearEdit.SetEnabled(false)
							sge.RublesEdit.SetValue(0)
							sge.RublesEdit.SetEnabled(false)
							sge.DollarsEdit.SetValue(0)
							sge.DollarsEdit.SetEnabled(false)
							sge.SaveButton.SetEnabled(false)
						},
					},
				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	r := mw.Bounds()
	scrWidth := int(win.GetSystemMetrics(win.SM_CXSCREEN))
	scrHeight := int(win.GetSystemMetrics(win.SM_CYSCREEN))
	mw.SetBounds(walk.Rectangle{
		X:      int((scrWidth - r.Width) / 2),
		Y:      int((scrHeight - r.Height) / 2),
		Width:  r.Width,
		Height: r.Height,
	})
	win.SetWindowLong(mw.Handle(), win.GWL_STYLE,
		win.GetWindowLong(mw.Handle(), win.GWL_STYLE) & ^win.WS_MINIMIZEBOX & ^win.WS_MAXIMIZEBOX)
	mw.Run()
}
