// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package main

import (
    "citron/config"
    "citron/errors"
    "citron/merge"
    "citron/process"
    "citron/statistic"
    "citron/store"
    "citron/transport"
    "flag"
    "fmt"
    "github.com/xfali/goutils/log"
    "path/filepath"
    "strings"
)

var logLv = map[string]int{
    "DEBUG": log.DEBUG,
    "INFO":  log.INFO,
    "WARN":  log.WARN,
    "ERROR": log.ERROR,
}

func main() {
    sourceDir := flag.String("s", "", "source dir")
    destUri := flag.String("d", "", "dest uri")
    checksumType := flag.String("checksum", "", "checksum type: MD5 | SHA1")
    incremental := flag.String("incr", "true", "incremental backup")
    newRepo := flag.String("n", "true", "creating a new backup repository every time")
    mergeSrc := flag.String("merge-src", "", "path of src merge dir")
    mergeDest := flag.String("merge-dest", "", "path of dest merge dir")
    mergeSave := flag.String("merge-save", "", "dir save merge result")
    logPath := flag.String("log-path", "./citron.log", "log file path")
    logLevel := flag.String("log-lv", "INFO", "log level: DEBUG | INFO | WARN | ERROR")
    multiTask := flag.Int("multi-task", 1, "backup multi task number")

    flag.Parse()

    config.GConfig.SourceDir = *sourceDir
    config.GConfig.DestUri = *destUri
    config.GConfig.ChecksumType = *checksumType
    config.GConfig.Incremental = *incremental == "true"
    config.GConfig.NewRepo = *newRepo == "true"
    config.GConfig.MultiTaskNum = *multiTask

    fmt.Printf("config: %s\n", config.GConfig.String())
    logWriter := log.NewFileLogWriter(*logPath)
    if logWriter == nil {
        log.Fatal("log writer init failed")
    }
    defer logWriter.Close()
    log.Level = logLv[*logLevel]
    log.Writer = logWriter
    log.Log(log.Level, "config: %s\n", config.GConfig.String())

    if *mergeDest != "" || *mergeSrc != "" || *mergeSave != "" {
        if *mergeDest == "" || *mergeSrc == "" || *mergeSave == "" {
            log.Fatal("Merge param error, merge-src, merge-dest, merge-save must be not empty")
        }
        err := merge.Merge(*mergeSrc, *mergeDest, *mergeSave)
        if err != nil {
            log.Fatal(err.Error())
        }
    }

    checkResource()

    st := statistic.New()
    t, err := transport.Open(
        "file",
        config.GConfig.DestUri,
        config.GConfig.Incremental,
        config.GConfig.NewRepo,
        st)
    if err != nil {
        log.Fatal(err.Error())
    }
    defer t.Close()
    s, err := store.Open(
        "file",
        filepath.Join(filepath.Dir(config.GConfig.SourceDir), config.InfoDir, "root.meta"))
    if err != nil {
        log.Fatal(err.Error())
    }
    defer s.Close()

    errP := process.Process(config.GConfig.SourceDir, t, s, st)

    fmt.Printf(st.String())
    log.Log(log.Level, st.String())

    if errP != nil {
        log.Fatal(errP.Error())
    }
}

func checkResource() {
    if config.GConfig.SourceDir == "" {
        log.Fatal(errors.SourceDirNotExists.Error())
    }

    if config.GConfig.DestUri == "" {
        log.Fatal(errors.TargetUriEmpty.Error())
    }

    if config.GConfig.SourceDir == config.GConfig.DestUri {
        log.Fatal(errors.SourceAndTargetSame.Error())
    }

    if strings.Index(config.GConfig.DestUri, config.GConfig.SourceDir) != -1 {
        log.Fatal(errors.SourceOrTargetError.Error())
    }
}
