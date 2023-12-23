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

Can Contain: 
    - "Some javascript-ish values"
    - [ 5, 2.3, 1e-3, 0x20, "\n", "🐈", "\U0001f408" ]
      # supports c,go,python,etc. style escape codes

Related Projects:
  - "YAML"       # https://yaml.org/
  - "JSON"       # http://json.org/
  - "NestedText" # https://nestedtext.org/
```

Some differences from yaml:

* Documents hold a single value.
* String literals must be quoted ( as in json. )
* Multiline strings use a custom heredoc syntax.
* No flow style ( although there is an array syntax. )
* No anchors or references.
* Comments can be captured during decoding, and returned as part of the data.

It isn't intended to be a subset of yaml, but it tries to be close enough to leverage some syntax highlighting in markdown, editors, etc.

Status 
----

Version 0.7

The go implementation successfully reads and writes some well-formed documents.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/ionous/tell)](https://pkg.go.dev/github.com/ionous/tell)
![Go](https://github.com/ionous/tell/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ionous/tell)](https://goreportcard.com/report/github.com/ionous/tell)

### Missing features

* serialization of structs not supported ( only maps, slices, and primitive values. )
* arrays should (probably) support nested arrays; 
* arrays should (probably) support comments.
* error reporting needs improvement.

see also the [issues page](https://github.com/ionous/tell/issues).

Usage
-----

```go

// Read a tell document.
func ExampleUnmarshal() {
	var out any
	const msg = `- Hello: "\U0001F30F"`
	if e := tell.Unmarshal([]byte(msg), &out); e != nil {
		panic(e)
	} else {
		fmt.Printf("%#v", out)
	}
	// Output:
	// []interface {}{map[string]interface {}{"Hello:":"🌏"}}
}

// Write a tell document.
func ExampleMarshal() {
	m := map[string]any{
		"Tell":           "A yaml-like text format.",
		"What It Is":     "A way of describing data...",
		"What It Is Not": "A subset of yaml.",
	}
	if out, e := tell.Marshal(m); e != nil {
		panic(e)
	} else {
		fmt.Println(string(out))
	}
	// Output:
	// Tell: "A yaml-like text format."
	// What It Is: "A way of describing data..."
	// What It Is Not: "A subset of yaml."
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

The individual elements of a sequence, and pairs of key-values in a mapping are called the "terms" of the collection.

### Documents
Documents are most often text files. UTF8, no byte order marks. 

"Structural whitespace" in documents is restricted to the ascii space and the ascii linefeed. Quoted strings can have horizontal tabs; single line strings, for perhaps obvious reasons, can't contain linefeeds. All other Unicode control codes are disallowed ( and, so cr/lf is considered an error. )

### Values
Any **scalar**, **array**, **sequence**, **mapping**, or **heredoc**.

### Scalars

* **bool**: `true`, or `false`.
* **raw string** ( backtick ): `` `backslashes are backslashes.` ``
* **interpreted string** ( double quotes ): `"backslashes indicate escaped characters."`
* **number**: 64-bit int or float numbers optionally starting with `+`/`-`; floats can have exponents `[e|E][|+/-]...`; hex values can be specified with `0x`notation. Like json, but unlike yaml: Inf and NaN are not supported. _( may expand to support https://go.dev/ref/spec#Integer_literals, etc. as needed. )_ 

A scalar value always appears on a single line. There is no null keyword, null is implicit where no explicit value was provided. Only heredocs support multi-line strings. _( Comments are defined as a hash followed by a space in order to maybe support css style hex colors, ie. `#ffffff`. Still thinking about this one. )_

**Escaping**: The individually escaped characters are: `a` ,`b` ,`f` ,`n` ,`r` ,`t` ,`v` ,`\` ,`"`. 
And, for describing explicit unicode points, `tell` uses the same rules as Go, namely: `\x` escapes for any unprintable ascii chars (bytes less than 128), `\u` for unprintable code points of less than 3 bytes, and `\U` for (four?) the rest.

### Arrays
Arrays use a syntax similar to javascript  (ex. `[1, 2, ,3]` ) except that a comma with no explicit value indicates a null value. Arrays cannot contain collections; heredocs in arrays are discouraged. _( fix: Currently, arrays cannot contain other arrays, nor can they contain comments. )_ 

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
2. **interpreted**, triple quotes: newlines act as word separators; backslashes are special; double newlines provide structure; single quotes don't need to be escaped ( but can be. )

Whitespace in both string types is influenced by the position of the closing heredoc marker. Therefore, any text to the left of the closing marker is an error. Both string types can define an custom tag to end the heredoc ( even if, unfortunately, that breaks `yaml` syntax highlighting. )

_(TBD: if documents should be trimmed of trailing whitespace: many editing programs are likely to do this by default. however, that would make intentional trailing whitespace in raw heredocs impossible.)_

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

Note that the interpreted heredoc is different from some more common implementations. The newline here exists for formatting the tell document, not the string.
```yaml
"""
hello
doc
"""
```
yields:
```hello doc```

while:
```yaml
"""
hello

line
"""
```
yields:

```
hello 
line
```

***Custom end tags***

I quite like the way some markdown implementations provide syntax coloring of triple quoted strings when there's a filetype after the quotes. ( for example: ` ```C++` ) Many of them, also nicely ignore any text after the filetype, and so lines like ` ```C++  something something`, even if maybe not technically *legal*, still provide good syntax highlighting.

With that in mind, Tell uses a triple less-than redirection marker (`<<<`) to define a custom end tag.  ( Triple to match the quotes. ) The redirection marker allows an author to have a filetype, or not. For example: ` ```C++ <<<END`, or if no filetype is desired:` ```<<<END`.

Maybe in some far off distant age, tell-aware syntax coloring could display the heredoc with fancy colors.

### Comments
Hate me forever, comments are preserved, are significant, and introduce their own indentation rules. 

**Rationale:** Comments are a good mechanism for communicating human intent. In [Tapestry](https://git.sr.ht/~ionous/tapestry), story files can be edited by hand, visually edited using blockly, or even extracted to present documentation; therefore, it's important to preserve an author's comments across different transformations. ( This was one of the motivations for creating tell. )

Similar to yaml, tell comments begin with the `#` hash, **followed by a space**, and continue to the end of a line. Comments cannot appear within a scalar _( **TBD**: comma separated arrays split across lines might be an exception. )_  

This implementation stores the comments for a collection in a string called a "comment block". Each collection has its own comment block stored in the zeroth element of its sequence, the blank key of its mappings, or the comment field of its document.

**When comments are preserved, collections are one-indexed.** On the bright side, this means that no special types are needed to store tell data: just native go maps and slices. 

The readme in package note gets into all the specifics.


Changes
-----

0.3 -> 0.4: 
	- adopt the golang (package stringconv) rules for escaping strings.
  - simplify the attribution of comments in the space between a key (or dash) and its value.
  - change the decoder api to support custom sequences, mirroring custom maps; package 'maps' is now more generically package 'collect'.
  - encoding/decoding heredocs for multiline strings
  - encoding/decoding of arrays; ( encoding will write empty collections as arrays; future: a heuristic to determine what should be encoded as an array, vs. sequence. )
  - the original idea for arrays was to use a bare comma full-stop format. switched to square brackets because they are easier to decode, they can support nesting, and are going to be more familiar to most users. ( plus, full stop (.) is tiny and easy to miss when looking at documents. )
 
 0.4 -> 0.5:
 	- simplify comment handling
 	
 0.5 -> 0.6:
 	- bug fixes, and re-encoding of comments

0.6 -> 0.7
	- replace comment raw string buffer usage with an opaque object ( to make any future changes more friendly )