Notes 
-----
The `notes` package parses tell comments.

Comments begin with the `#` hash, **followed by a space**, and continue to the end of a line. Comments cannot appear within a scalar _( **TBD**: comma separated arrays split across lines might be an exception. )_

For instance:

```yaml
"i am a value" # i am a comment.
```

The decoder can be configured to discard those comments, or to keep them. 
When comment are kept, they are stored in their most closely related collection as a string called a "comment block".
The comment block lives in the zeroth element of its sequence, the blank key of its mapping, or the comment field of its document.

For example, the following document has two comment blocks: one for the document, and one for the sequence.

```yaml
# header
- "value" # inline
# footer
```

If, after decoding, the document and its comments were written out in json, this would be the result:


```json
{
    "comment": "# header\f# footer",
    "content": [
        "\r\r# inline",
        "value"
    ]
}
```

Although this method means every sequence is one indexed, and every mapping has a blank key: it provides a simple way to read, write, and manipulate comments for user code.

Associating comments with collections
------------------

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
                    # and ideally are left aligned.

- "alternate trailing example"
    # alternatively, trailing comments comments can start on the next line 
    # still indented to the right of the value
    # and, again, ideally left-aligned, with no additional nesting.
  
- # comments can live after the key
    # between a signature ( or dash ) and its value.
    # they support nesting to continue the comment.
  "entry example"
   
-
# placing a comment aligned with the left edge of the 
# collection keys generates an implicit nil for the preceding key
# ( this comment is therefore treated like a header for the next element )

- # without nesting, the entire block 
  # gets stored with the parent container
  # as a comment describing this key.
  sub:collection:with: "first element"
  
- 
  # introducing a fully blank line forces
  # a blank key comment, and starts a header 
  # for the first element of this next sub collection.
  sub:collection:with: "one element"

- # when any comment nesting is used
    # the leading comments go to the parent container
    # as a key comment.
  # while the final comment group 
    # acts a a header for the sub collection.
  sub:collection:with: "one element"

# closing comments are allowed for a document.
# presumably matching the indentation at the start.
```

Comment Block Generation
------------------------

Each comment block is a single string of text generated in the following manner:

* Individual comments are stored as encountered. Hash marks are kept as part of the string, as is all whitespace. The latter is necessary to support commenting out heredoc lines; the former makes it easier to split individual comments out of their comment blocks.

* A carriage return (`\r`) replaces each key (or dash) and value within a comment block.

* Line breaks between comments use line feed (`\n`), while nesting indentation gets normalized to a single horizontal tab (`\t`)<sup>\[1]</sup>. 

* Inline trailing comments start directly after the value's carriage return. Any continuing comment lines are treated as nested comments. If, instead of starting inline, the trailing comment starts on the line following the value ( a trailing block comment ) -- a nested comment should directly follow the carriage.

* To separate comments for different collection entries, the end of each entry is indicated with a form feed (`\f`). _( Putting it at the end of each, rather than the start of each, keeps the first header comment with the first entry. )_

* Fully blank lines are ignored.
  
The resulting block can then be trimmed of trailing whitespace. And, a program that wants to read ( or maintain ) comments can split or count by form feed to find the comments of particular entries. 

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