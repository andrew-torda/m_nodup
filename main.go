// 7 Oct 2023

// moodle_dup reads an XML file of a moodle question bank.
// It looks for duplicates and removes them.
//
// Usage:
//      moodle_dup input.xml [output.xml]

package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func usage() {
	const usageMessage = "moodle_dup [flags] infile.xml [outfile.xml]"
	fmt.Fprintln(os.Stderr, usageMessage)
	flag.Usage()
}

// makeName takes a.xml and returns a_nodup.xml, or takes a.foo and returns
// a_nodup.foo. Given "a", return a_nodup.
func makeName(s string) string {
	i := strings.LastIndex(s, ".")
	postpend := "_nodup"
	if i == -1 {
		return (s + postpend)
	}
	a := s[:i]
	suffix := s[i:]
	r := a + postpend + suffix
	return r
}

// nothing does nothing
func nothing(x interface{}) { return }

// firstfew gives us the first few characters of a byte slice
func firstfew(s []byte) string {
	ss := s[:10]
	return string(ss)
}

// Some things one may want from the question structure.
// Most important are questiontext and perhaps answer.
// You could also add
// Qname      string   `xml:"name>text"`
//
type qstrct struct {
	XMLName xml.Name `xml:"question"`
	Attr    xml.Attr `xml:"name,attr"`
	Ctgry   string   `xml:"category>text"`
	Qtxt    string   `xml:"questiontext>text"`
	Answr   string   `xml:"answer>text"`
	Inner   []byte   `xml:",innerxml"` // you need the comma before innerxml
}

type quiz struct {
	XMLName xml.Name `xml:"quiz"`
	Qs      []qstrct `xml:"question"`
}

func breaker() {}

// cleanupQstns is given a slice of questions and removes duplicates.
// Try to put each question in a hash. If it is already there, remove
// the question from the slice.
// If answerFlag is set, do not add the <answer> into the key
// Do not worry about the order, we will sort them anyway.
func cleanupQstns(qstns []qstrct, answrFlag bool) ([]qstrct, error) {
	qhash := make (map[string]struct{}, len(qstns))
	var nothing struct {}
	for _, q := range qstns{
		var key string
		if answrFlag {
			key = q.Qtxt
		} else {
			key = q.Qtxt + q.Answr
		}
		if _, ok := qhash[key]; ok {
			fmt.Println ("dup", q.Qtxt)
		} else {
			qhash[key] = nothing
		}
	}
	breaker()
	return qstns, nil
}

// dedup takes input and output descriptors and flags. Currently just
// the answrFlag which tells us not to worry about checking the answers.
func dedup(fIn *os.File, fOut io.Writer, answrFlag bool) error {
	var qz quiz
	d := xml.NewDecoder (fIn)
	if e := d.Decode(&qz); e != nil {
		return e
	}
	var err error
	qz.Qs, err = cleanupQstns(qz.Qs, answrFlag)
	return err
}

// mymain does the work and either returns an error, which might be nil
// if all is well.
func mymain() error {
	var answrFlag = flag.Bool("a", false, "do not check if Answers are the same")

	var inputName, outputName = "", ""
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Too few command line args")
		usage()
		return errors.New("Too few command line arguments")
	}
	inputName = flag.Arg(0)
	if flag.NArg() == 2 {
		outputName = flag.Arg(1)
	} else if flag.NArg() > 2 {
		fmt.Fprintln(os.Stderr, "Too many (", flag.NArg(), ") command line args")
		flag.Usage()
		//		return errors.New("Too many command line args")
	}

	{
		t := "off"
		if *answrFlag {
			t = "off"
		}
		fmt.Println("answrFlag is", t)
	}
	if outputName == "" {
		outputName = makeName(inputName)
	}
	fmt.Println("inputName", inputName, " outputName", outputName)
	var err error
	fIn, err := os.Open(inputName)
	if err != nil {
		return err
	}
	defer fIn.Close()
	fOut, err := os.Create(outputName)
	if err != nil {
		return err
	}
	defer fOut.Close()
	err = dedup(fIn, fOut, *answrFlag)
	return err
}

func main() {
	if e := mymain(); e != nil {
		fmt.Fprintln(os.Stderr, "Broke:", e)
		os.Exit(1)
	}
	os.Exit(0)
}
