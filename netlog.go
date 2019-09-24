// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

//NetLogWriter -- This log writer sends output to a log server
type NetLogWriter struct {
	rec chan *LogRecord
	rot chan bool

	lk *sync.Mutex

	// The logging format
	format string

	//send interval(in second)
	sendinterval int

	//max lines(local cache)
	maxcaches int

	//app name (server will create diffrent folder acording to this)
	appname string

	//log server
	serverurl string

	//caches
	caches []string
}

//LogWrite This is the NetLogWriter's output method
func (w *NetLogWriter) LogWrite(rec *LogRecord) {
	w.lk.Lock()
	w.caches = append(w.caches, FormatLogRecord(w.format, rec))
	w.lk.Unlock()
}

//SendToServer --
func (w *NetLogWriter) SendToServer() {

	var lines []string
	w.lk.Lock()
	if len(w.caches) > 0 {
		lines = w.caches
		w.caches = []string{}
	}
	w.lk.Unlock()
	if len(lines) > 0 {
		v := url.Values{}
		v.Set("app", w.appname)

		buf, _ := json.Marshal(&lines)
		v.Set("logs", string(buf))

		encValues := v.Encode()

		body := ioutil.NopCloser(strings.NewReader(encValues)) //把form数据编下码
		client := &http.Client{}
		req, _ := http.NewRequest("POST", w.serverurl, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		ioutil.ReadAll(resp.Body)
	}
}

//Close --
func (w *NetLogWriter) Close() {

}

// NewNetLogWriter creates a new LogWriter which writes to the given file and
// has rotation enabled if rotate is true.
//
// If rotate is true, any time a new log file is opened, the old one is renamed
// with a .### extension to preserve it.  The various Set* methods can be used
// to configure log rotation based on lines, size, and daily.
//
// The standard log-line format is:
//   [%D %T] [%L] (%S) %M
func NewNetLogWriter(app, sendurl string) *NetLogWriter {
	w := &NetLogWriter{
		rec:          make(chan *LogRecord, LogBufferLength),
		rot:          make(chan bool),
		appname:      app,
		format:       "[%D %T] [%L] (%S) %M",
		sendinterval: 3,
		maxcaches:    300,
		serverurl:    sendurl,
		lk:           &sync.Mutex{},
		caches:       []string{},
	}

	go func() {
		for {
			time.Sleep(time.Duration(w.sendinterval) * time.Second)
			w.SendToServer()
		}
	}()

	return w
}

// SetFormat Set the logging format (chainable).  Must be called before the first log
// message is written.
func (w *NetLogWriter) SetFormat(format string) *NetLogWriter {
	w.format = format
	return w
}

// SetMsgSendInterval Set the log message send interval (chainable).  Must be called before the first log
// message is written.
func (w *NetLogWriter) SetMsgSendInterval(format string) *NetLogWriter {
	w.format = format
	return w
}

// SetMaxCacheLines Set the max local message cache (chainable). if the max messages reaches send event triggered. Must be called before the first log
// message is written.
func (w *NetLogWriter) SetMaxCacheLines(format string) *NetLogWriter {
	w.format = format
	return w
}

// SetAppName Set the app name (chainable).  Must be called before the first log
// message is written.
func (w *NetLogWriter) SetAppName(format string) *NetLogWriter {
	w.format = format
	return w
}
