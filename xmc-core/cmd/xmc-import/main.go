package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xmc-dev/xmc/xmc-core/importer"
	"github.com/xmc-dev/xmc/xmc-core/importer/xmc"
)

var format string
var update bool

var object string
var fp string

var log = importer.Log

func init() {
	flag.StringVar(&format, "format", "xmc", "the format of the objects to be imported")
	flag.BoolVar(&update, "update", true, "update if already exists")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: xmc-import object path...\n")
}

type importFunc func() importer.Spec

func importGrader() importer.Spec {
	gi := xmc.NewGraderImporter()
	gs, err := gi.ReadGrader(fp)
	if err != nil {
		log.Fatalf("Error while reading grader: %v", err)
	}
	return gs
}

func importDataset() importer.Spec {
	di := xmc.NewDatasetImporter()
	ds, err := di.ReadDataset(fp)
	if err != nil {
		log.Fatalf("Error while reading dataset: %v", err)
	}
	return ds
}

func importTask() importer.Spec {
	ti := xmc.NewTaskImporter()
	ts, err := ti.ReadTask(fp)
	if err != nil {
		log.Fatalf("Error while reading task: %v", err)
	}
	return ts
}

func importTaskList() importer.Spec {
	tli := xmc.NewTaskListImporter()
	tls, err := tli.ReadTaskList(fp)
	if err != nil {
		log.Fatalf("Error while reading task list: %v", err)
	}
	return tls
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 2 {
		usage()
		os.Exit(1)
	}
	object = flag.Args()[0]
	var f importFunc
	switch object {
	case "grader":
		f = importGrader
	case "dataset":
		f = importDataset
	case "task":
		f = importTask
	case "tasklist":
		f = importTaskList
	default:
		fmt.Fprintln(os.Stderr, "unsupported object type")
		os.Exit(1)
	}
	for _, afp := range flag.Args()[1:] {
		fp = afp
		spec := f()
		isNew, err := spec.IsNew()
		if err != nil {
			log.Fatalf("Error while checking if %s is new: %v", object, err)
		}
		needsUpdate, err := spec.NeedsUpdate()
		if err != nil {
			log.Fatalf("Error while checking if %s needs update: %v", object, err)
		}
		if isNew || (update && needsUpdate) {
			err = spec.Import()
			if err != nil {
				log.Fatalf("Error while importing %s: %v", object, err)
			}
		} else {
			log.Infof("Skipping %s %s", object, afp)
		}
	}
}
