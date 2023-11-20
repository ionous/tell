# header buffering

```yaml
 mapping:
   - # padding
     # buffered padding or header
     1
 
   - 2
     # footer for 2
   
   # header for 3
   - 3
   # this becomes a footer for "mapping"
   # note, it has the same indentation as the header for 3.
   
# i am footer for document.
```