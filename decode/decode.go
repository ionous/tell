package decode

import (
	"io"
	"log"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/charmed"
	"github.com/ionous/tell/maps"
	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
)

func Decode(src io.RuneReader, maps maps.BuilderFactory, comments notes.Commentator) (ret any, err error) {
	d := decoder{mapMaker: mapMaker{maps}, memo: makeMemo(comments)}
	var x, y int
	run := charm.Parallel("parallel",
		charmed.FilterControlCodes(),
		d.decode(), // tbd: wrap with charmed.UnhandledError()? why/why not.
		charmed.DecodePos(&y, &x),
	)
	if e := charm.Read(src, run); e != nil {
		log.Println("error at", y, x)
		err = e
	} else {

		if next := charm.RunState(runes.Eof, run); next != nil {
			if es, ok := next.(charm.Terminal); ok && es != charm.Error(nil) {
				log.Println("error at", y, x)
				err = es
			}
		}
		if err == nil {
			// pop everything
			ret, err = d.out.finalizeAll()
		}

		d.memo.OnEof() // fix; can this be removed?
	}
	return
}
