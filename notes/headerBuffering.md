# header buffering notes

```yaml
# document header
# without nesting this becomes an element header
mapping:
  - # key comment
    # buffered, key or header
    1
  - 2
    # footer for 2
  # header for 3
  - 3
  # this becomes a *footer* for "mapping"
  # note, it has the same indentation as the header for 3.
# document footer
```