package util

import (
	"github.com/john-nguyen09/phpintel/indexer"
	"github.com/sourcegraph/go-lsp"
)

func WriteLocation(serialiser *indexer.Serialiser, location lsp.Location) {
	serialiser.WriteString(string(location.URI))
	WritePosition(serialiser, location.Range.Start)
	WritePosition(serialiser, location.Range.End)
}

func WritePosition(serialiser *indexer.Serialiser, position lsp.Position) {
	serialiser.WriteInt(position.Line)
	serialiser.WriteInt(position.Character)
}

func ReadLocation(serialiser *indexer.Serialiser) lsp.Location {
	return lsp.Location{
		URI: lsp.DocumentURI(serialiser.ReadString()),
		Range: lsp.Range{
			Start: ReadPosition(serialiser),
			End:   ReadPosition(serialiser),
		},
	}
}

func ReadPosition(serialiser *indexer.Serialiser) lsp.Position {
	return lsp.Position{
		Line:      serialiser.ReadInt(),
		Character: serialiser.ReadInt(),
	}
}
