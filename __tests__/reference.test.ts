import { App } from '../src/app';
import { Indexer } from '../src/index/indexer';
import { getCaseDir } from "../src/testHelper";
import * as path from "path";
import { ReferenceTable } from '../src/storage/table/referenceTable';
import { pathToUri } from '../src/util/uri';

describe('Testing functions around references', () => {
    it('should return the reference at the cursor', async () => {
        App.setUpForTest();

        const indexer = App.get<Indexer>(Indexer);
        const caseDir = getCaseDir();
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const refTestFile = path.join(caseDir, 'reference', 'references.php');

        await indexer.indexFile(path.join(caseDir, 'global_symbols.php'));
        await indexer.indexFile(path.join(caseDir, 'function_declare.php'));
        await indexer.indexFile(refTestFile);

        let ref1 = await refTable.findAt(pathToUri(refTestFile), 10);

        console.log(ref1);
    });
});