Comment Sections
------------------

```
+------------------+
|                  |
|   [Doc Header]   |
|                  |
+-----------+------+
| Key Name: |      | ----> inline comment
|    _______|      |
|   |              |
|   |    [Keyed]   |  }---> key comments
|   |              |
+---+--------+-----+
|   "Scalar" |     |  ----> inline trailing 
|    ________+     |
|   |              | 
|   |  [Trailing]  |  }---> trailing block
|   |              |
+---+--------------+
|                  |
|   [Inter Key]    |  ( until the next key )
|                  |
+------------------+
|                  |  ( used for docs with scalar values;
|   [Doc Footer]   |    for docs with a sequence or mapping, 
|                  |    these are read by interkey )
+------------------+    
```