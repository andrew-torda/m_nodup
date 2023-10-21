// 7 Oct 2023

// moodle_dup reads an XML file of a moodle question bank.
// It looks for duplicates and removes them.
// There is very little interpretation of xml. The input is parsed into
// questions and their content stored and written verbatim.
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
	return a + postpend + suffix
}

type qstrct struct {
	XMLName xml.Name `xml:"question"`
	Attr    xml.Attr `xml:",any,attr"`
	Inner   string   `xml:",innerxml"` // you need the comma before innerxml
}

type quiz struct {
	XMLName xml.Name `xml:"quiz"`
	Attr    xml.Attr `xml:",any,attr"`
	Qs      []qstrct `xml:"question"`
}

// getToken reads a string a looks for the element matching tagType. We have a
// bit of text and expect that the <stuff> token is there. Before we go into
// the loop looking for the element, we can do a quick test seeing if "stuff" is there
// at all. If we find <stuff> we make the string start from there, so the loop
// does not go from the start of all elements. We still use the xml.functions
// because they handle things like attributes, closing of tags and all the hard
// work.

func getToken(s string, tagType string) (string, error) {
	type qtxt struct {
		Qtxt string `xml:"text"` // moodle items hide their info in
	}                            // <text> tags.
	if n := strings.Index(s, "<"+tagType); n == -1 { // quick bailout
		return "", nil // if there is no chance of the tag being present
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
			if t.Name.Local == tagType {
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

// whoToDelete is given a slice of questions. It returns a bool slice
// called delme. If delme[i] is true, then question [i] is a duplicate
// and should be removed. Check for identity by putting the questiontext
// into a map structure. We remove white space from the question text when
// doing the comparison.
func whoToDelete(qstns []qstrct, delme []bool, fdbackFlag bool) ([]bool, error) {
	var nothing struct{}
	qhash := make(map[string]struct{}, len(qstns))
	re := regexp.MustCompile("\\s+") // white space

	ndel := 0
	for i, q := range qstns {
		if key, err := getToken(q.Inner, "questiontext"); err != nil {
			return nil, err
		} else { // if key == "", don't worry. We don't mark it for deletion
			if fdbackFlag {
				if tt, e := getToken(q.Inner, "generalfeedback"); err != nil {
					return nil, e
				} else if tt != "" {
					key = key + tt
				}
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

// dedup takes input and output descriptors and flags. It reads the xml, looks
// for duplicates and writes questions back to the io.writer.
func dedup(fIn io.Reader, fOut io.Writer, fdbackFlag bool) error {
	var qz quiz
	var err error // Get a new decoder and do the work in one line:
	if err := xml.NewDecoder(fIn).Decode(&qz); err != nil {
		return err
	}
	qstns := qz.Qs
	delme := make([]bool, len(qstns)) // entries to be deleted
	if delme, err = whoToDelete(qstns, delme, fdbackFlag); err != nil {
		return err
	}

	for i := 0; i < len(qstns)-1; {
		if delme[i] {
			qstns = append(qstns[:i], qstns[i+1:]...)
			delme = append(delme[:i], delme[i+1:]...)
		} else {
			i++
		}
	}

	if delme[len(qstns)-1] == true { // special treatment of last
		qstns = qstns[:len(qstns)-1] // element
	}
	qz.Qs = qstns // The shortened, cleaned up slice

	fmt.Fprintf(fOut, "%s", xml.Header)
	enc := xml.NewEncoder(fOut)
	enc.Indent("", " ")
	return (enc.Encode(qz))
}

// mymain does the work, decoding, removing duplicates and encoding
func mymain() error {
	var fdbackFlag = flag.Bool("f", false, "add the feedback field to be checked for copies")

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
		flag.Usage()
		return fmt.Errorf("Too many (%d) command line args", flag.NArg())
	}

	if outputName == "" {
		outputName = makeName(inputName)
	}
	fmt.Println("inputName", inputName, " outputName", outputName)

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
	return dedup(fIn, fOut, *fdbackFlag)
}

// main
func main() {
	if e := mymain(); e != nil {
		fmt.Fprintln(os.Stderr, "Broke:", e)
		os.Exit(1)
	}
	os.Exit(0)
}
