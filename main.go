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
	"regexp"
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

// Some things one may want from the question structure.
// Most important are questiontext and perhaps answer.
// You could also add
// Qname      string   `xml:"name>text"`
// Ctgry   string   `xml:"category>text,omitempty"`
// Qtxt    string   `xml:"questiontext>text,omitempty"`
// Answr   string   `xml:"answer>text,omitempty"`
// Chars   string   `xml:",chardata"` - don't know what this is
//

type qstrct struct {
	XMLName xml.Name `xml:"question,omitempty"`
	Attr    xml.Attr `xml:",any,attr"`
	Inner   string   `xml:",innerxml"` // you need the comma before innerxml
}

type quiz struct {
	XMLName xml.Name `xml:"quiz"`
	Attr    xml.Attr `xml:",any,attr"`
	Qs      []qstrct `xml:"question"`
}

func breaker()           {}
func nofunc(interface{}) {}

// getToken reads a string a looks for the element matching tokType. We have a
// bit of text and expect that the <stuff> token is there. Before we go into
// the loop looking for the element, we can do a quick test seeing if "stuff" is there
// at all.

func getToken(s string, tokType string) (string, error) {
	type qtxt struct {
		Qtxt string `xml:"text"`
	}
	if n := strings.Index(s, "<"+tokType); n == -1 {
		return "", nil
	} else {
		s = s[n:]
	}
	dec := xml.NewDecoder(strings.NewReader(s))
	for {
		tok, err := dec.Token()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				return "", nil
			}
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == tokType {
				var qt qtxt
				if err := dec.DecodeElement(&qt, &t); err != nil {
					return "", err
				} else {
					return qt.Qtxt, nil
				}
			}
		}
	}
}

// whotToDelete is given a slice of questions. It returns a bool slice
// called delme. If delme[i] is true, then question [i] is a duplicate
// and should be removed. Check for identity by putting the questiontext
// into a map structure.
func whoToDelete(qstns []qstrct, delme []bool) ([]bool, error) {
	var nothing struct{}
	qhash := make(map[string]struct{}, len(qstns))
	re := regexp.MustCompile("\\s+")

	ndel := 0
	for i, q := range qstns {
		if key, err := getToken(q.Inner, "questiontext"); err != nil {
			return nil, err
		} else { // if key == "", don't worry. We don't mark it for deletion
			if strings.Contains(key, "Berechnung des GC Gehalts überprüft. Kopieren Sie Ihren Code d") {
				//				fmt.Println ("DBG", i, re.ReplaceAllString(key, ""))

			}
			key = re.ReplaceAllString(key, "")
			if _, ok := qhash[key]; ok {
				ndel++
				delme[i] = true
			} else {
				qhash[key] = nothing
			}
		}
	}
	fmt.Println(ndel, "duplicates from", len(qstns))
	return delme, nil
}

// dedup takes input and output descriptors and flags. Currently just
// the answrFlag which tells us not to worry about checking the answers.
func dedup(fIn io.Reader, fOut io.Writer, answrFlag bool) error {
	var qz quiz
	var err error
	if err := xml.NewDecoder(fIn).Decode(&qz); err != nil {
		return err
	}
	qstns := qz.Qs
	delme := make([]bool, len(qstns)) // entries to be deleted
	if delme, err = whoToDelete(qstns, delme); err != nil {
		return err
	}

	for i := 0; i < len(qstns) - 1; {
		if delme[i] {
			qstns = append(qstns[:i], qstns[i+1:]...)
			delme = append(delme[:i], delme[i+1:]...)
		} else {
			i++
		}
	}

	if delme[len(qstns)-1] == true { //
		qstns = qstns[:len(qstns)-1]
	}
	qz.Qs = qstns // The shortened, cleaned up slice

	fmt.Fprintf(fOut, "%s", xml.Header)
	enc := xml.NewEncoder(fOut)
	enc.Indent("", " ")
	enc.Encode(qz)
	return (enc.Flush()) // newer versions of go have enc.Close().
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
