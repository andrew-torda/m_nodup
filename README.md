# Purpose
In the teaching platform moodle, we have duplicate questions in our question bank. Let us try to remove them.

# How

## How to run it

  m_nodup [-f] infile  [outfile]

Normally, we just take the question text and use it to see if questions are questions are identical. If we have the `-f` flag, It will also check the `<generalfeedback>` tag. Comparisons are made after removing whitespace.

If you give an `outfile`, it will be created for output. The `.xml` is not appended. Put it in the filename. If you do not give an `outfile`, the name `infile_nodup.xml` will be used for output.

** How it runs

The question bank should be exported as an XML file. This program eats the XML file and prints out a new XML file with duplicates removed.

### Identical questions, decisions

We say that two questions are identical if
 - their question text is identical (after removing white space) and
 - their generalfeedback field is identical and we have the -f flag

These criteria are not obvious. We could also consider just checking the whole question.

# Example

   go build m_nodup.go
   m_nodup -f testme.xml

Should remove 13 duplicates and produce a file called testme_nodup.xml.

# Bugs
Moodle xml does not include authors or dates. This information will be lost. There is not much that can be done there.

If no duplicates are found, we still write a file. This should probably be fixed.