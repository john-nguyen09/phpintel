import { App } from '../src/app';
import { Indexer } from '../src/index/indexer';
import { getCaseDir, getDebugDir, dumpAstToDebug } from "../src/testHelper";
import * as path from "path";
import { ReferenceTable } from '../src/storage/table/referenceTable';
import { pathToUri } from '../src/util/uri';
import { RefResolver } from "../src/provider/refResolver";
import { PhpDocumentTable } from '../src/storage/table/phpDoc';

describe('Testing functions around references', () => {
    it('should return the reference at the cursor', async () => {
        App.init(path.join(getDebugDir(), 'storage'));

        const indexer = App.get<Indexer>(Indexer);
        const caseDir = getCaseDir();
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const refTestFile = path.join(caseDir, 'reference', 'references.php');

        await indexer.indexFile(path.join(caseDir, 'class_methods.php'));
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
            await refTable.findAt(refTestUri, 241),
            await refTable.findAt(refTestUri, 226),
            await refTable.findAt(refTestUri, 243),
            await refTable.findAt(refTestUri, 255)
        ];

        // let refTestDoc = await phpDocTable.get(refTestUri);
        // let typeDesignator = await refTable.findAt(refTestUri, 196);
        // console.log(await RefResolver.getClassConstructorSymbols(refTestDoc, typeDesignator));

        expect(refs).toMatchSnapshot();

        await App.clearCache();
    });
});