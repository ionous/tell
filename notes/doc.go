// Package notes provides a method for building comment blocks.
//
// each term in a collection's comment block uses the following pattern:
// ( noting that sub-collections can't have inline comments )
//
// \n # sub headers
// \r # padding follows the key ( \n \t # nested padding... )
// \r # inline comments follow the value ( \n \t # nested inline... )
// \n # footer
// \n # extra footer; no nesting
// \f
package notes
