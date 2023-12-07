Notes 
-----
The notes package parses tell document comments.


Comments begin with the `#` hash, **followed by a space**, and continue to the end of a line. Comments cannot appear within a scalar _( **TBD**: comma separated arrays split across lines might be an exception. )_

The following rules reflect how i think most people comment yaml-like documents.
This, of course, is based on absolutely no research whatsoever. Still, my hope is that no special knowledge is needed. 

```yaml
# header comments start at the indentation of the following collection.
# they can continue at the same level of indentation ( aka sub-headers. )
- "header example"

# for consistency with key comments ( described below )
  # nested indentation is allowed starting on the second line.
  # nested headers and sub-headers cannot co-exist.
- "nested header example"

- "inline example"  # inline trailing comments follow a scalar value.
                    # they continue on following lines
                    # if, and only if, they are all left aligned.

- "alternate trailing example"
    # alternatively, trailing comments comments can start on the next line 
    # still indented to the right of the value
    # and all left-aligned, with no nesting.
  
- # comments can live after the key
    # between a signature ( or dash ) and its value.
    # they support nesting to continue the comment.
  "entry example"
  
- # comments for nil values
  # are the same as any other collection entries.
  
# for entries that contain sub collections....

- # without nesting, the first *line* describes the entry.
  # subsequent lines act as a header for
  # the first element of the sub-collection.
  sub:collection:with: "first element"
  
- 
  # this is on the second line, so it describes the element.
  sub:collection:with: "one element"

- # for consistency, the entry
    # can use nesting here.
  # the header can also use nesting...
    # just like this line does.
  sub:collection:with: "one element"

# closing comments are allowed for a document.
# presumably matching the indentation at the start.
```

#### Comment storage:

Each comment block is a single string of text generated in the following manner:

* Individual comments are stored as encountered. Hash marks are kept as part of the string, as is all whitespace. The latter is necessary to support commenting out heredoc lines; the former makes it easier to split individual comments out of their comment blocks.

* A carriage return (`\r`) replaces each key (or dash) and value within a comment block.

* Line breaks between comments use line feed (`\n`), while nesting indentation gets normalized to a single horizontal tab (`\t`)<sup>\[1]</sup>. 

* Inline trailing comments start directly after the value's carriage return. Any continuing comment lines are treated as nested comments. If, instead of starting inline, the trailing comment starts on the line following the value ( a trailing block comment ) -- a nested comment should directly follow the carriage.

* To separate comments for different collection entries, the end of each entry is indicated with a form feed (`\f`). _( Putting it at the end of each, rather than the start of each, keeps the first header comment with the first entry. )_

* Fully blank lines are ignored.
  
The resulting block can then be trimmed of trailing whitespace. And, a program that wants to read ( or maintain ) comments can split or count by form feed to find the comments of particular entries. 

For example, the following document generates two comment blocks: one for the document (`# header\f# footer`), and one for the collection containing the scalar "value" (`\r\r# inline`).

```yaml
# header
- "value" # inline
# footer
```

The pattern for a single scalar value in a collection looks like this:

```
# header ( \n \t # nested header; exclusive with sub headers )
\n # sub headers
\r # key comments follow the key ( or dash ) ( \n \t # nested comment... )
\r # inline trailing comments directly follow the value ( \n \t # nested trailing... )
( \n \t # alternatively, trailing block comments start nesting directly after the carriage )
\f 
```

The tests have plenty of other examples.

\[1] _the two rules mean that every nested comment starts with a \n\t combo. while i would have preferred using a single vertical tab... unfortunately, json excludes \v from its set of escaped control codes. :/ translating tell comments to json ( like tapestry does ) would have resulted in the ugly looking long format `\u000b`. the alternative, using backspace ( the only escape not already used ), seems even more problematic._