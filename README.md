Tell
--------
A yaml-like text format with json-ish values. 


```yaml
Tell: "A yaml-like text format."

# Does this look suspiciously like the yaml overview?
# I have no idea how that could have happened.
What It Is: "A way of describing data containing string, number, and boolean values, 
   as well as collections of those values. As in yaml, collections can be 
   both key-value mappings, and sequences."

What It Is Not: "A subset of yaml."

Can Contain: 
    - "Some javascript-ish values"
    - # with c, go, python, etc. style escape codes.
      [ 5, 2.3, 1e-3, 0x20, "\n", "🐈", "\U0001f408" ]

Related Projects:
  - "YAML"       # https://yaml.org/
  - "JSON"       # http://json.org/
  - "NestedText" # https://nestedtext.org/
```

Some differences from yaml:

* Supports a single unified UTF-8 document.
* String literals must be quoted ( as in json. )
* Boolean values are only `true` or `false` ( as in json. )
* Multiline string blocks use a custom heredoc syntax.
* No flow style ( although inline arrays are supported. )
* No anchors or references.
* Comments can be captured during decoding, and returned as part of the data.

It isn't intended to be a subset of yaml, but it tries to be close enough<sup>tm</sup> to leverage existing yaml syntax highlighting and validation.

Status 
----

The go implementation successfully reads and writes well-formed documents.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/ionous/tell)](https://pkg.go.dev/github.com/ionous/tell)
![Go](https://github.com/ionous/tell/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ionous/tell)](https://goreportcard.com/report/github.com/ionous/tell)

### Missing features

* serialization of structs not supported.
* arrays should (probably) support nested arrays.
* arrays should (probably) handle trailing comments.
* error reporting could use improvement.

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
	// maps/orderedmap uses Ian Coleman's ordered map.
	// ( https://github.com/iancoleman/orderedmap ) 
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

The individual elements of a sequence, and the pairs of key-values in a mapping, are called the "terms" of the collection.

### Documents
Documents are most often text files. UTF8, no byte order marks. 

Whitespace is restricted to the ascii space ( 0x20 ) and the ascii linefeed ( 0xa ). The exception is quoted strings which additionally allow horizontal tabs ( ascii 0x9. ) All other control codes are disallowed ( and, so cr/lf is considered an error. )

**TBD:** should comments allow horizontal tabs?

### Values
Any **scalar**, **array**, **sequence**, **mapping**, or **heredoc**.

### Scalars

* **bool**: `true`, or `false`.
* **raw string** ( backtick ): `` `Preserves *all* whitespace. Backslashes are backslashes.` ``
* **trimmed string** ( single quotes ): `'Treats newlines as semantic: folding lines together by injecting a single space. Eats all indentation while still preserving trailing whitespace. Backslashes are backslashes.`
* **interpreted string** ( double quotes ): `"Treats newlines as semantic: folding lines together by injecting a single space. Eats all indentation while still preserving trailing whitespace. Backslashes indicate escaped characters."`
* **number**: 64-bit int or float numbers optionally starting with `+`/`-`; floats can have exponents `[e|E][|+/-]...`; hex values can be specified with `0x`notation. As per `json`, Inf and NaN are not supported. _( **TBD**: may expand to support https://go.dev/ref/spec#Integer_literals )_ 
* **null**: There is no null keyword. instead, null is implicit where no explicit value was provided. 

**Scalar strings** act like their yaml counterparts. They can span lines, and the "trimmed" and "interpreted" strings use _semantic newlines._  This means linefeeds in the text are treated as a single space. Only a fully blank line is treated as having a newline. As per `yaml`: all indentation at the start of line is ignored, and ( although i do not like it ) all trailing space is kept. Also like `yaml`, a single backslash at the end of a line eliminates any space, joining the following line seamlessly. 

The "raw string" type does not exist in `yaml`. It acts like the `Go` raw string. It preserves all whitespace in between the opening tick and the closing tick exactly as is.

**Escaping**: Backslashes in interpreted strings can preceded certain characters to provide special values: `a` (alert - 0x7) ,`b` (backspace - 0x8), `f` (formfeed - 0xc), `n` (linefeed - 0xa), `r` (return - 0xd ), `t` (htab - 0x9), `v` (vtab - 0xb), `\` (backslash - 0x5c), `"` (doublequote - 0x22), and linefeeds (for joining lines.) For describing explicit unicode points, `tell` uses the same rules as `Go`, namely: `\x` escapes any unprintable ascii chars (bytes less than 128), `\u` any unprintable code points of less than 3 bytes, and `\U` for (four?) the rest.

**TBD:** `tell` could support css hex colors ( ex. `#ffffff` ) because comments are defined as "hash followed by a space". still thinking about this one....

### Arrays
Arrays use a syntax similar to javascript  (ex. `[1, 2, ,3]` ) except that a comma with no explicit value indicates a null value. Arrays cannot contain collections; heredocs in arrays are discouraged. _( **TODO**: arrays cannot currently contain other arrays, nor can they contain comments. )_ 

#### Sequences
Sequences define an ordered list of values. 
Entries in a sequence start with a dash and whitespace separates the value.
Additional entries in the same sequence start on the next line with the same indentation as the previous entry.
```
  - true
  - false
```

As in `yaml`, whitespace after a dash can include newlines. And that rule means nested sequences can start inline. For example, `- - 5` is equivalent to the json `[[5]]`.

Unlike `yaml`, if a value is specified on a line following its dash, the value **must** start on a column at least two spaces to the right of the dash. ( ie. while newlines and spaces are both whitespace, indentation still matters. ) This rule keeps values aligned.

```
  - "here"
  - 
    "there"
```

#### Mappings
Mappings relate keys to values in an ordered fashion.

Keys for mappings are defined using **signatures**: a series of one or more words, separated by colons, ending with a colon and whitespace. For example: `Hello:there: `. The first character of each word must be a (unicode) letter; subsequent characters can include letters, digits, and underscores _( **TBD**: this is somewhat arbitrary; what does yaml do? )_

For the same reason that nested sequences can appear inline, mappings can. However, `yaml` doesn't allow this and it's probably bad style. For example: `Key: Nested: "some value"` is equivalent to the json `{"Key:": {"Nested:": "some value" }`. Like sequences, if the value of a mapping appears on a following line, two spaces of indentation are required.

_**Note**: [Tapestry](git.sr.ht/~ionous/tapestry) wants those trailing colons. In this implementation the interpretation of `key:` is therefore `"key:"` not `"key"`. This feels like an implementation detail, and could be an exposed as an option._

#### Heredocs

Heredocs exist both to capture newlines, and to control the leading indentation of strings. They can appear anywhere a scalar string can, except not within inline arrays. Unlike the scalar strings: newlines are interpreted as actual newlines. Indentation is controlled by the indentation of the closing quotes ( or closing tag. )

There are three heredoc types, one for each scalar string type:

1. **raw** (` ``` `) using triple backticks. Backslashes are backslashes; the final newline of the heredoc is preserved.
2. **trimmed** (`'''`) using triple single quotes. Backslashes are backslashes; the final newline gets eaten.
2. **interpreted** (`"""`) using triple double quotes. Backslashes follow the same rules as interpreted scalar strings. Quotes don't need to be escaped in heredocs (`\"`) but can be. The final newline is preserved by default, but a backslash on the end of the final line can eat it.

The position of the closing heredoc tag controls the overall indentation. Any text to the left of the closing tag is an error. All three kinds can define an custom tag.

```
  - """
        i am an interpreted heredoc.
          this line has two extra spaces in front.
        lines are not automatically folded together.
        but this line ends with a backslash, \
        so it folds seamless into this line.
        the newline following this line is preserved.
        """

  - ```<<<END
    i am a raw heredoc with a custom closing tag.
    all three heredoc types support custom closing tags.
    raw strings preserve whitespace, including the newline after this line.
    END
    
  - '''
    this here is a trimmed doc. 
    backslashes \ are backslashes.
    trimmed heredocs eat the final newline.
    '''
```

***Custom end tags***

I like the way markdown allows syntax coloring for block quotes if there's a filetype after the quotes. ( for example: ` ```go` ) Many implementations also nicely ignore any text after the filetype, and so any opening like ` ```go  something something`, even if not technically legal, still works okay.

With that in mind, `tell` uses a redirection marker (`<<<`) to define a custom end tag. ( Triple to match the quotes. ) The redirection allows an author to still include a filetype, or not. For example: ` ```go <<<END`. Or, if a filetype isn't desired, just: ` ```<<<END`.

Maybe in some far off distant age, tell-aware syntax coloring could display the heredoc with fancy colors.

***Yaml compatibility***

Because `tell` relies on existing yaml syntax validation ( and color schemes ), there is one additional heredoc type provided for compatibility. It opens with the [yaml pipe](https://yaml.org/spec/1.2.2/#812-literal-style) (`|`), but ends with one of the tell triple quotes.

```yaml
  - |
        i am a heredoc starting with a pipe (|) for compatibility.
        if i end with double quotes, then backslashes are interpreted \
        otherwise, they are not. raw and trimmed strings are also still the same.
        and, indentation is controlled by the position of the closing quotes.
        ( just like all heredocs. )
        """
```

It looks a bit odd, but it allows `yaml` validation to succeed. ( Until `tell` conquers the world or something, and has validators and syntax highlighting in all the best editors. Then the pipe syntax can be deprecated. )

None of the ["chomping" indicators](https://yaml.org/spec/1.2.2/#8112-block-chomping-indicator) are allowed here ( the triple quote styles subsume that functionality. ) And, neither is the ["folded style"](https://yaml.org/spec/1.2.2/#813-folded-style) ( since that's equivalent to the scalar string functionality. ) Custom tags aren't allowed either, unfortunately, because nothing is allowed to follow the pipe.

  _Why pipe? In yaml, the pipe preserves newlines, and its default chomping also preserves the final newline. That's enough context enough so that -- if this was yaml -- the resulting string could still be evaluated to produce a tell-like result._

### Comments
Hate me forever, comments are significant, they must follow the indentation rules of the document, and -- in this implementation -- they can be accessed directly as part of the data.

Similar to yaml, tell comments begin with the `#` hash,  **but must be followed by a space**. They continue to the end of their line. Comments cannot appear within a scalar.

**When comments are preserved, collections are one-indexed.** This means no special types are needed to store tell data: only native go maps and slices. Different implementations could handle this in other ways. The basic point is that comments are both well-defined and easily accessible.

**Rationale:** Comments are a good mechanism for communicating human intent. In [Tapestry](https://git.sr.ht/~ionous/tapestry), story files can be edited by hand, visually edited using blockly, or even extracted for documentation. Therefore, it's important to preserve an author's comments across different transformations. ( This was one of the motivations for creating tell. )

The [readme](https://github.com/ionous/tell/blob/main/note/README.md)  in `package note` gets into all the specifics.

Version History
-----

0.8.1 -> 0.9.0: changes string folding; adds new string types.

  - scalar strings can now span lines. they follow the same rules as yaml's strings.
  - heredocs no longer fold lines since that's the role of the scalar strings.
  - for yaml compatibility, heredocs can optionally start with a pipe (|).
  - bug fix for the charm utility function `ParseEof()` ( affected tapestry, but not tell. )
  - fixes for various `staticcheck` warnings

0.8.0 -> 0.8.1:

  - catch tabs in whitespace
  - bug fix: report better errors when unable to decode a mistyped boolean literal (ex. `truex` )

0.7.0 -> 0.8.0:

  - Changes the encoder's interface to support customizing the comment style of mappings and sequences independently.
  - bug fix: when specifying map values: allow sequences to start at the same indentation as the key and allow a new map term to start after the sequence ends. ( previously, it generated an error, and an additional indentation was required. ) For example:
```yaml
  - First:  # the value of First is a sequence containing "yes"
    - "yes" 
    Second: # Second is an additional entry in the same map as First
    - "okay" 
``` 
  -  bug fix: for all other values, an indentation greater than the key is required. For example: 
```yaml
  First:
  "this is an error."
``` 

0.6 -> 0.7.0:

  - replace comment raw string buffer usage with an opaque object ( to make any future changes more friendly )

0.5 -> 0.6:

  - bug fixes, and re-encoding of comments

0.4 -> 0.5:

  - simplify comment handling

0.3 -> 0.4: 

  - adopt the golang (package stringconv) rules for escaping strings.
  - simplify the attribution of comments in the space between a key (or dash) and its value.
  - change the decoder api to support custom sequences, mirroring custom maps; package 'maps' is now more generically package 'collect'.
  - encoding/decoding heredocs for multiline strings
  - encoding/decoding of arrays; ( encoding will write empty collections as arrays; future: a heuristic to determine what should be encoded as an array, vs. sequence. )
  - the original idea for arrays was to use a bare comma full-stop format. switched to square brackets because they are easier to decode, they can support nesting, and are going to be more familiar to most users. ( plus, full stop (.) is tiny and easy to miss when looking at documents. )
