package main

import (
	"flag"
	"fmt"
	"tktools/time_tool"
	"tktools/webdav_uploader"
)

var (
	T               bool
	tt              int64
	ts              string
	Webdav          string
	Webdav_sub_mode string
)

func init() {
	flag.BoolVar(&T, "T", false, "Use time tools")
	flag.Int64Var(&tt, "Tt", 0, "time sub: timestamp to covert to string")
	flag.StringVar(&ts, "Ts", "", "time sub: time in string format to covert to timestamp")

	flag.StringVar(&Webdav, "W", "", "Use webdav uploader to upload a folder")
	flag.StringVar(&Webdav_sub_mode, "Ws", "", "use specific sub config for uploading")
}

func doTimeActions() {
	anyActionDone := false
	if tt > 0 {
		time_tool.PrintTimestampToString(tt)
		anyActionDone = true
	}
	if len(ts) > 0 {
		time_tool.PrintStringToTimestamp(ts)
		anyActionDone = true
	}
	if !anyActionDone {
		time_tool.PrintDefault()
	}
}

func main() {
	flag.Parse()
	if T {
		doTimeActions()
		return
	}
	if len(Webdav) > 0 {
		webdav_uploader.Upload(Webdav, Webdav_sub_mode)
		return
	}
	webdav_uploader.Upload(Webdav, Webdav_sub_mode)
	fmt.Println(`tktools.exe [-TW] {sub commands}`)
	flag.PrintDefaults()
}
