import { App } from '../src/app';
import { Indexer } from '../src/index/indexer';
import { getCaseDir, getDebugDir } from "../src/testHelper";
import * as path from "path";
import { ReferenceTable } from '../src/storage/table/referenceTable';
import { pathToUri } from '../src/util/uri';

describe('Testing functions around references', () => {
    it('should return the reference at the cursor', async () => {
        App.init(path.join(getDebugDir(), 'storage'));

        const indexer = App.get<Indexer>(Indexer);
        const caseDir = getCaseDir();
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const refTestFile = path.join(caseDir, 'reference', 'references.php');

        await indexer.indexFile(path.join(caseDir, 'global_symbols.php'));
        await indexer.indexFile(path.join(caseDir, 'function_declare.php'));
        await indexer.indexFile(refTestFile);

        let refTestUri = pathToUri(refTestFile);
        let refs = [
            await refTable.findAt(refTestUri, 7),
            await refTable.findAt(refTestUri, 14),
            await refTable.findAt(refTestUri, 10),
            await refTable.findAt(refTestUri, 37),
            await refTable.findAt(refTestUri, 51),
            await refTable.findAt(refTestUri, 42),
        ];

        expect(refs).toMatchSnapshot();
    });
});