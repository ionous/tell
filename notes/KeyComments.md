
```yaml
- # key comment only
  - "sequence"

- # in fact, a fully left aligned block of comments
  # are all key comments.
  - "sequence"

- # key comment with nesting:
    # it might be neat if the alignment
    # of the nest with the sub-collection
    # caused the nest to become an element header
    # but notes doesnt have enough info for that.
    - "sequence"

- # nesting groups split between key comments and header comments.
    # ( a fully blank line would work too )
  # this is a header for the sub-sequence.
  - "sequence"

- 
  # a header because of the leading blank line
  - "sequence"

- # key comment 
  # a header line because any use of
    # any nesting causes an opt-in.
  - "sequence"

- # specific antagonism
  # no nesting 
  # no nesting
  # no nesting
      # and now? probably illegal
  - "sequence"

- # key comment
  # usually a header comment
   # ( because of nesting )
  # but with multiple blocks
    # of nesting 
  # only the*last* block
   #  becomes a header.
  - "sequence"
```