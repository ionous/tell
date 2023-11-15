Tell
--------
A yaml-like text format with json-ish scalars. 


```yaml
Tell: "A yaml-like text format."

What It Is: """
   A way of describing data containing string, number, and boolean values, 
   and collections of those values. As in yaml, collections can be 
   both key-value mappings, and sequences of values.
   """

Related Projects:
  - "YAML" # Official YAML Website: https://yaml.org/
  - "JSON" # Official JSON Website: http://json.org/
```

Some major differences:

* Comments are important.
* String literals must be quoted ( as in json. )
* Multiline strings use a custom heredoc syntax.
* Except for string literals and comments, tabs are always invalid whitespace.
* No flow style ( although there is an array syntax. )
* The order of maps matters.
* No anchors or references.
* Documents hold a single value.

It isn't intended to be a subset of yaml, but it tries to be close enough to leverage syntax highlighting in editors, etc.


Status 
----

Version 0.

The go implementation successfully reads (some?) well-formed documents; it doesn't attempt to write documents.

### Missing features

* heredocs are defined but not yet supported.
* arrays would be nice, but aren't implemented.
* error reporting needs improvement.
* no serialization of go maps and slices to tell.

see also the [issues page](https://github.com/ionous/tell/issues).

Usage
-----

```go
// from tellExample_test.go
func ExampleString() {
	str := `true` // some tell document
	// tell/maps/imap contains a slice based ordered map implementation.
	// tell/maps/stdmap generates standard (unordered) go maps.
	// tell/maps/orderedmap uses Ian Coleman's ordered map implementation.
	doc := tell.NewDocument(imap.Builder, tell.KeepComments)
	// ReadDoc takes a string reader
	if res, e := doc.ReadDoc(strings.NewReader(str)); e != nil {
		panic(e)
	} else {
		// the results contains document level comments
		// and the content that was read.
		fmt.Println(res.Content)
	}

	// Output: true
}

```

Description
-----

Tell consists of collections of values, along with optional comments. These types are described below.

### Collections
* **Document**: a collection containing a single **value**.
* **Sequences**: aka lists: an ordered series of one or more **values**.
* **Mappings**: aka ordered dictionaries: relates **keys** to **values**. 

### Documents
Documents are most often text files. UTF8, no byte order marks. 

Whitespace in documents is restricted to the ascii space and the ascii linefeed; r/lf is considered an error; tabs are disallowed everywhere except for in string values and comments. ( This differs from `yaml` where, for example, tabs can appear after single space indentation. )

_( BUG: the implementation currently errors on tabs in comments. )_

### Values
Any **scalar**, **array**, **sequence**, **mapping**, or **heredoc**.

### Scalars

* **bool**: `true`, or `false`.
* **raw string** ( backtick ): `` `backslashes are backslashes.` ``
* **interpreted string** ( double quotes ): `"backslashes indicate escaped characters."`<sup>\[1]</sup>
* **number**: 64-bit int or float numbers optionally starting with `+`/`-`; floats can have exponents `[e|E][|+/-]...`; hex values can be specified with `0x`notation. _( may expand to support https://go.dev/ref/spec#Integer_literals, etc. as needed. )_  _( **TBD**: the implementation currently produces floats, and only floats. that's to match json, but what's best? )_ 

A scalar value always appears on a single line. There is no null keyword, null is implicit where no explicit value was provided.

_( It is sad that hex colors can't live as `#ffffff`. Maybe it would have been cool to use lua style comments ( -- ) instead of yaml hashes. For now, comments are defined as a hash followed by a space while i keep thinking about it. )_

\[1]: _the set of escaped character is: `a` ,`b` ,`f` ,`n` ,`r` ,`t` ,`v` ,`\` ,`"` ._

### Arrays
An array is a list of comma separated scalars, ending with an optional fullstop: `1, 2, 3.` 
_( **TBD**: all on one line?  )_  The fullstop is necessary when indicating an empty array. Nested arrays are not a thing; use sequences.

#### Sequences
Sequences define an ordered list of values. 
Entries in a sequence start with a dash, whitespace separates the value.
Additional entries in the same sequence start on the next line with the same indentation as the previous entry.
```
  - true
  - false
```

As in `yaml`, whitespace after the dash can include newlines. And, the lack of differentiation between newline and space implies that nested sequences can be declared on one line. For example, `- - 5` is equivalent to the json `[[5]]`.

Unlike `yaml`, if a value is specified on a line following its dash, the value must start on a column two spaces to the right of the dash. ( ie. while newlines and spaces are both whitespace, indentation still matters. ) This rule keeps values aligned.

```
  - "here"
  - 
    "there"
```

#### Mappings
Mappings relate keys to values in an ordered fashion.

Keys are defined with **signatures**: a series of one or more words, separated by colons, ending with a colon and whitespace. For example: `Hello:there: `. The first character of each word must be a (unicode) letter; subsequent characters can include letters, digits, and underscores _( **TBD**: this is somewhat arbitrary; what does yaml do? )_

For the same reason that nested sequences can appear inline, mappings can. However, `yaml` doesn't allow this and it's probably bad style. For example: `Key: Nested: "some value"` is equivalent to the json `{"Key:": {"Nested:": "some value" }`. Like sequences, if the value of a mapping appears on a following line, two spaces of indentation are required.

_**Note**: [Tapestry](git.sr.ht/~ionous/tapestry) wants those colons. In this implementation the interpretation of `key:` is therefore `"key:"` not `"key"`. This feels like an implementation detail, and could be an exposed as an option._

#### Heredocs

Heredocs provide multi-line strings wherever a scalar string is permitted ( but not in an array, dear god. )

There are two types, one for each string type:

1. **raw**, triple backticks: newlines are structure; backslashes are backslashes.
2. **interpreted**, triple quotes: newlines are presentation; backslashes are special; double newlines provide structure.

Whitespace in both string types is influenced by the position of the closing heredoc marker. Therefore, any text to the left of the closing marker is an error. Both string types can define an custom tag to end the heredoc ( even if, unfortunately, that breaks `yaml` syntax highlighting. )


```
  - """
    i am a heredoc interpreted string.
    these lines are run together
    each separated by a single space.
     this sentence has an extra space in front.

    a blank line ^ becomes a single newline.
    trailing spaces in that line, or any line, are eaten.
    """

  - """
    this interpreted string starts with
1234 spaces. ( due to the position of the closing triple-quotes. )
"""

  - ```<<<END
    i am a heredoc raw string using a custom closing tag.
     this line has a single leading space.

    a blank line ^ is a blank line
    because raw strings preserve any and all whitespace, except:
    the starting and ending markers don't introduce newlines.
    ( so this line doesn't end with a newline. )
    END
```

_i'm quite taken with the way some markdown tools provide syntax coloring of triple quoted string blocks if the author specifies a filetype after the quotes. ( for example: ` ```C++ ... ` ) since github's version ignores text after the filetype, something like ` ```C++  END ... ` can still display correct coloring in some cases. however, since that wouldn't play well with end markers, tell requires redirection markers so it can support both filetypes and custom heredoc tags. ie. ` ```C++  <<<END` that way, maybe in some far off distant age, tell-aware syntax coloring could display the heredoc with fancy colors._

### Comments
Hate me forever, comments are preserved, are significant, and introduce their own indentation rules. 

**Rationale:** Comments are a good mechanism for communicating human intent. In [Tapestry](https://git.sr.ht/~ionous/tapestry), story files can be edited by hand, visually edited using blockly, or even extracted to present documentation, therefore it's important to preserve an author's comments across different transformations. ( This was one of the motivations for creating tell. )

Tell's rules reflect how i think most people comment yaml-like documents.
This, of course, is based on absolutely no research whatsoever. Still, my hope is that no special knowledge is needed. 

Comments begin with the `#` hash, **followed by a space**, and continue to the end of a line. Comments cannot appear within a scalar _( **TBD**: comma separated arrays split across lines might be an exception. )_  The position of a comment determines which document element a comment describes.

```yaml
# header comments start at the indentation of the following collection.
# header comments can continue at the same level of indentation.
- "header example"

# for consistency with the comments for collection entries ( described below )
  # nested indentation is allowed starting on the second line.
  # is that good? i don't know.
- "nested header example"

- "inline example"  # inline comments follow a value.
                    # they continue on following lines
                    # if their comment hash marks are aligned.

- "footer example"
  # footer comments continue at a consistent indentation
  # right at, or to the right of, the entry's indentation.
  # ( but no nesting. )
  
- "inline and footer"   # an inline comment...
  # can be extended with normal
  # footer comments. ( and still no nesting. )

- # comments describing a specific collection entry 
    # start immediately after the entry's dash ( or signature. )
    # they support nesting to continue the comment.
  "entry example"
  
- # comments for nil values
  # are the same as any other collection entries.
  
# for entries that contain sub collections....

- # without nesting, the first *line* describes the entry.
  # subsequent lines act as a "header" for
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

This implementation stores the comments for each collection separately in its own "comment block". Each comment block gets stored in the zeroth index of its sequence, the blank key of its mappings, or the comment field of its document. **This means all collections are one-indexed.** On the bright side, this means that no special types are needed to store tell data: just native go maps and slices. _(TODO: arrays should probably be one-indexed for consistency's sake, and to allow space for comments in future expansion.)_ 

Each comment block is a single string of text generated in the following manner:

* Individual comments are stored as encountered. Each line gets trimmed of trailing spaces, hash marks are kept as part of the string. _( Keeping the hash makes it more obvious how internal leading spaces are handled, and makes it easier to split comments out of their stored block of text. )_
* A carriage return (`\r`) separates the comments before an entry's marker (its dash or signature) from the comments after; effectively it replaces the value of an entry.
* Line breaks between comments use line feed (`\n`), all nesting indentation is normalized to a single horizontal tab (`\t`)<sup>\[1]</sup>. For these purposes, left-aligned inline comments are considered nested; footer comments are not.
* To separate comments for different collection entries, the end of each entry is indicated with a form feed (`\f`). _( Putting it at the end of each, rather than the start of each, keeps the first header comment with the first entry. )_
* Fully blank lines are ignored.
  
The resulting block can then be trimmed of trailing whitespace. And, a program that wants to read ( or maintain ) comments can split or count by form feed to find the comments of particular entries. 

For example, the comment block `"# header\r# inline\n# footer"` represents:

```yaml
# header 
- "value" # inline
# footer
```

The tests have plenty of other examples.

\[1] _for nesting, i would have preferred a single vertical tab rather than the newline htab combo. unfortunately, json excludes vertical tab from its set of escaped control codes even though javascript supports it. :/ translating tell comments to json ( like tapestry does ) would have resulted in this ugliness: `\u000b`. the alternative, using backspace ( the only escape not already used ), feels very wrong._