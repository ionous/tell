Tell
--------
A yaml-like text format with json-ish values. 


```yaml
Tell: "A yaml-like text format."

# Does this look suspiciously like the yaml overview?
# I have no idea how that could have happened.
What It Is: """
   A way of describing data containing string, number, and boolean values, 
   and collections of those values. As in yaml, collections can be 
   both key-value mappings, and sequences of values.
   """

What It Is Not: "A subset of yaml."
	
Related Projects:
  - "YAML"       # https://yaml.org/
  - "JSON"       # http://json.org/
  - "NestedText" # https://nestedtext.org/
```

Some differences from yaml:

* String literals must be quoted ( as in json. )
* Multiline strings use a custom heredoc syntax.
* Except for string literals and comments, tabs are always invalid whitespace.
* No flow style ( although there is an array syntax. )
* The order of maps matters.
* No anchors or references.
* Documents hold a single value.
* Comments can be captured during decoding, and returned as part of the data.

It isn't intended to be a subset of yaml, but it tries to be close enough to leverage some syntax highlighting in markdown, editors, etc.

Status 
----

Version 0.3

The go implementation successfully reads and writes some well-formed documents.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/ionous/tell)](https://pkg.go.dev/github.com/ionous/tell)
![Go](https://github.com/ionous/tell/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ionous/tell)](https://goreportcard.com/report/github.com/ionous/tell)

### Missing features

* heredocs are defined but not yet supported.
* arrays would be nice, but aren't implemented.
* error reporting needs improvement.
* no serialization of structs ( only maps, slices, and primitives. )

see also the [issues page](https://github.com/ionous/tell/issues).

Usage
-----

```go
func ExampleUnmarshal() {
	var b bool
	if e := tell.Unmarshal([]byte(`true`), &b); e != nil {
		panic(e)
	} else {
		fmt.Println(b)
	}
	// Output: true
}

func ExampleMarshal() {
	b := true
	if out, e := tell.Marshal(b); e != nil {
		panic(e)
	} else {
		fmt.Println(string(out))
	}
	// Output: true
}

// slightly lower level usage:
func ExampleDocument() {
	str := `true` // some tell document
	// maps/imap contains a slice based ordered map implementation.
	// maps/stdmap generates standard (unordered) go maps.
	// maps/orderedmap uses Ian Coleman's ordered map implementation.
	doc := decode.NewDocument(imap.Make, notes.DiscardComments())
	// ReadDoc takes a string reader
	if res, e := doc.ReadDoc(strings.NewReader(str)); e != nil {
		panic(e)
	} else {
		fmt.Println(res)
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

"Structural whitespace" in documents is restricted to the ascii space and the ascii linefeed. Quoted strings can have horizontal tabs; single line strings, for perhaps obvious reasons, can't contain linefeeds. All other Unicode control codes are disallowed ( and, so cr/lf is considered an error. )

_( BUG: the implementation currently errors on tabs in comments. )_

### Values
Any **scalar**, **array**, **sequence**, **mapping**, or **heredoc**.

### Scalars

* **bool**: `true`, or `false`.
* **raw string** ( backtick ): `` `backslashes are backslashes.` ``
* **interpreted string** ( double quotes ): `"backslashes indicate escaped characters."`<sup>\[1]</sup>
* **number**: 64-bit int or float numbers optionally starting with `+`/`-`; floats can have exponents `[e|E][|+/-]...`; hex values can be specified with `0x`notation. Like json, but unlike yaml: Inf and NaN are not supported. _( may expand to support https://go.dev/ref/spec#Integer_literals, etc. as needed. )_  _( **TBD**: the implementation currently produces floats, and only floats. that's to match json, but what's best? )_ 

A scalar value must always appears on a single line. There is no null keyword, null is implicit where no explicit value was provided. ( Heredocs support multiline strings. )

_( It is sad that hex colors can't live as `#ffffff`. Maybe it would have been cool to use lua style comments ( -- ) instead of yaml hashes. For now, comments are defined as a hash followed by a space while i keep thinking about it. )_

\[1]: _the set of escaped characters includes: `a` ,`b` ,`f` ,`n` ,`r` ,`t` ,`v` ,`\` ,`"`.
rather than try to invent robust unicode handling, tell uses the same rules as go: `\x` escapes for any unprintable ascii chars (bytes less than 128), `\u` for unprintable code points of less than 3 bytes, and `\U` for (four?) the rest._

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

Keys for mappings are defined using **signatures**: a series of one or more words, separated by colons, ending with a colon and whitespace. For example: `Hello:there: `. The first character of each word must be a (unicode) letter; subsequent characters can include letters, digits, and underscores _( **TBD**: this is somewhat arbitrary; what does yaml do? )_

For the same reason that nested sequences can appear inline, mappings can. However, `yaml` doesn't allow this and it's probably bad style. For example: `Key: Nested: "some value"` is equivalent to the json `{"Key:": {"Nested:": "some value" }`. Like sequences, if the value of a mapping appears on a following line, two spaces of indentation are required.

_**Note**: [Tapestry](git.sr.ht/~ionous/tapestry) wants those colons. In this implementation the interpretation of `key:` is therefore `"key:"` not `"key"`. This feels like an implementation detail, and could be an exposed as an option._

#### Heredocs

Heredocs provide multi-line strings wherever a scalar string is permitted ( but not in an array, dear god. )

There are two types, one for each string type:

1. **raw**, triple backticks: newlines are structure; backslashes are backslashes.
2. **interpreted**, triple quotes: newlines are presentation; backslashes are special; double newlines provide structure.

Whitespace in both string types is influenced by the position of the closing heredoc marker. Therefore, any text to the left of the closing marker is an error. Both string types can define an custom tag to end the heredoc ( even if, unfortunately, that breaks `yaml` syntax highlighting. )

```yaml
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

**Rationale:** Comments are a good mechanism for communicating human intent. In [Tapestry](https://git.sr.ht/~ionous/tapestry), story files can be edited by hand, visually edited using blockly, or even extracted to present documentation; therefore, it's important to preserve an author's comments across different transformations. ( This was one of the motivations for creating tell. )

As in yaml, tell comments begin with the `#` hash, **followed by a space**, and continue to the end of a line. Comments cannot appear within a scalar _( **TBD**: comma separated arrays split across lines might be an exception. )_  

This implementation stores the comments for a collection in a string called a "comment block". Each collection has its own comment block stored in the zeroth element of its sequence, the blank key of its mappings, or the comment field of its document.

**This means all collections are one-indexed.** On the bright side, this means that no special types are needed to store tell data: just native go maps and slices. _( **TBD**: arrays will probably need to be one-indexed for consistency's sake, and to allow space for comments in future expansion.)_

The readme in package notes gets into all the specifics.


Changes
-----

0.3 - 0.4: 
	- adopted the golang (package stringconv) rules for escaping strings.
  - simplified the attribution of comments in the space between a key (or dash) and its value.
  - changes the decoder api to support custom sequences, mirroring custom maps; package 'maps' is now more generically package 'collect'.
  - encoder writes empty sequences as empty arrays
  - encoder writes heredocs for multiline strings