package analysis

import "testing"

func TestAnalysisTimeRegression1(t *testing.T) {
	withTestStore("test", "TestInfiniteLoopRegression1", func(store *Store) {
		doc := indexDocumentAndGet(store, "../cases/infiniteLoop1.php", "test1")

		doc.Load()
	})
}
