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

type qstrct struct {
	XMLName xml.Name `xml:"question"`
	Attr    xml.Attr `xml:"name,attr"`
	Type    string   `xml:"type"`
	Text    string   `xml:",text"`
	Comment string   `xml:",comment"`
	Name    string   `xml:",name"`
	Qtext   string   `xml:",questiontext"`
	CharDat string   `xml:",chardata"`
	Inner   []byte   `xml:",innerxml"` // you need the comma before innerxml
}

type quiz struct {
	XMLName xml.Name `xml:"quiz"`
	QQtext string    `xml:",questiontext"`
	Attr xml.Attr `xml:"name,attr"`
	Comment string `xml:",comment"`
	Qs []qstrct `xml:"question"`
}


// tokenInfo gets a token and prints out some things. This is for debugging
// or looking at the structures
// StartElement, EndElement, CharData, Comment, ProcInst, or Directive.
func tokenInfo(tok xml.Token) string {
	var s string
	switch t := tok.(type) {
	case xml.StartElement:
		s = fmt.Sprintf("startelement name: %v attributes ", t.Name)
		for _, a := range t.Attr {
			s = s + fmt.Sprintf("name >>%s<< Value >>%s<<", a.Name.Local, a.Value)
		}
	case xml.EndElement:
		s = fmt.Sprintln("endelement space Local|", t.Name.Space, "|", t.Name.Local)
	case xml.CharData:
		s = fmt.Sprintln("chardata|", string(t))
	case xml.Comment:
		s = fmt.Sprintln("comment |", string(t))
	case xml.ProcInst:
		s = fmt.Sprintln("ProcInst Target Inst |", t.Target, "|", string(t.Inst))
	case xml.Directive:
		s = fmt.Sprintln("Directive |", string(t))
	default:
		s = fmt.Sprintln("unknown token", tok)
	}
	return strings.TrimSpace(s)
}
func breaker () {}
// dedup takes input and output descriptors and flags. Currently just
// the answrFlag which tells us not to worry about checking the answers.
func dedup(fIn io.Reader, fOut io.Writer, answrFlag bool) error {
	d := xml.NewDecoder(fIn)

	//	var actions []Executer // this seems to be a list of things that were done

	// Finding the first Root tag
	// for {
	// 	tok, err := d.Token()
	// 	if err != nil {
	// 		return nil
	// 	}
	// 	fmt.Println("before start", tokenInfo(tok))
	// 	if _, ok := tok.(xml.StartElement); ok {
	// 		break
	// 	}
	// }
	fmt.Println("------ starting loop ---------")
	// Looping through the rest of the tokens
	// finding the start of each.

	for {
		v, err := d.Token()
		if err != nil {
			return nil
		}
		var qz quiz
		switch t := v.(type) {

		case xml.StartElement:
			fmt.Println(tokenInfo(v))
			breaker()
			if err := d.DecodeElement(&qz, &t); err != nil {
				return err
			}
			nothing(qz)
		case xml.EndElement:
			fmt.Println("endelement", tokenInfo(v))
			return nil
		default:
			fmt.Println(tokenInfo(v))
		}
	}
	return nil
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
