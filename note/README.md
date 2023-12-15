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
The comment block lives in the zeroth term of its sequence, the blank key of its mapping, or the comment field of its document.

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
# header comments start at the first line.
# if the document contains a collection,
# the comment becomes part of that collection.
- "header example"

- "inline suffix example"  # a "suffix" can follow a scalar value.
                           # they can continue on following lines:
                           # left-aligned with no additional nesting.

- "trailing suffix example"
    # alternatively, a suffix can start on the next line 
    # slightly indented to the right of the value.
    # and again, left-aligned, with no additional nesting.
 
-
# comments aligned to the left edge of a collection
# act as a header for the next term, not a suffix for the previous.
# therefore this also generates an implicit nil for the preceding key.
 
- # a "prefix" comment can live between 
  # a key ( or dash ) and a scalar value.
  "key comment example"
  
- # however, if the key has a sub-collection as its value,
  # the prefix becomes a header for that collection's first term.
  - "term header example"

# footer comments are allowed for a document.
# if the document contains a collection
# the footer becomes part of that collection.
```

Here's another way of visualizing the different comment types:

```
+------------------+
|                  | a document header exists for document scalar values
|     [Header]     | ( otherwise, the first comment becomes the first key header. )
|                  |
+-----------+------+
| Key Name: |      |  prefix comments annotate scalar values;
|    _______|      |  if there is a a sub-collection:
|   |              |  the prefix becomes a header for 
|   |  [ Prefix ]  |  the first element of that collection.
|   |              |  
+---+--------+-----+
|   "Scalar" |     |  suffix comments follow scalar values.
|    ________+     |  in the generated comment block,
|   |              |  these are distinguished from other comments
|   |  [ Suffix ]  |  by the inclusion of a horizontal tab.
|   |              |
+---+--------------+
|                  |  a footer becomes the header for the next key
|     [Footer]     |  ( if there is any such key )
|                  |  a document footer exists for document scalar values.
+------------------+
```


Comment Block Generation
------------------------

Each comment block is a single string of text generated in the following manner:

* Individual comments are stored in the order encountered. Hash marks are kept as part of the string, as is all whitespace. The former makes it easier to split individual comments out of their comment blocks; the latter is necessary to support commenting out heredoc raw strings. 

* A carriage return (`\r`) replaces each key (or dash) and value within a comment block.

* Line breaks between comments use line feed (`\n`), all nested indentation gets normalized to a horizontal tab (`\t`).

* Inline trailing comments start directly after the value's carriage return. Any continuing comment lines are treated as nested comments. A trailing block comment follows the carriage return with a line feed and nests all comments.

* The end of each term in a collection is indicated with a form feed (`\f`).

* Fully blank lines are ignored.
  
* The resulting block can be trimmed of control characters. 

A program that wants to read ( or maintain ) comments can split or count by form feed to find the comments of particular entries. 

The pattern for a single scalar value in a collection looks like this:

```
# header ( \n \t # nested header; exclusive with sub headers )
\n # sub headers
\r # key comments follow the key ( or dash ) ( \n \t # nested comment... )
\r # inline trailing comments directly follow the value ( \n \t # nested trailing... )
( \n \t # alternatively, trailing block comments start nesting directly after the carriage )
\f 
```