import { App } from '../src/app';
import { Indexer } from '../src/index/indexer';
import { getCaseDir, getDebugDir, dumpAstToDebug, dumpToDebug } from "../src/testHelper";
import * as path from "path";
import { ReferenceTable } from '../src/storage/table/referenceTable';
import { pathToUri } from '../src/util/uri';
import { RefResolver } from "../src/provider/refResolver";
import { PhpDocumentTable } from '../src/storage/table/phpDoc';
import { RefKind } from '../src/symbol/reference';
import { Symbol } from '../src/symbol/symbol';

describe('Testing functions around references', () => {
    it('should return the reference at the cursor', async () => {
        App.init(path.join(getDebugDir(), 'storage'));

        const indexer = App.get<Indexer>(Indexer);
        const caseDir = getCaseDir();
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const refTestFile = path.join(caseDir, 'reference', 'references.php');
        const testFile2 = path.join(caseDir, 'class_methods.php');

        await indexer.indexFile(testFile2);
        await indexer.indexFile(path.join(caseDir, 'class_constants.php'));
        await indexer.indexFile(path.join(caseDir, 'global_symbols.php'));
        await indexer.indexFile(path.join(caseDir, 'function_declare.php'));
        await indexer.indexFile(refTestFile);

        let refTestUri = pathToUri(refTestFile);
        let refs = [
            await refTable.findAt(refTestUri, 21),
            await refTable.findAt(refTestUri, 24),
            await refTable.findAt(refTestUri, 30),
            await refTable.findAt(refTestUri, 37),
            await refTable.findAt(refTestUri, 51),
            await refTable.findAt(refTestUri, 42),
            await refTable.findAt(refTestUri, 226),
            await refTable.findAt(refTestUri, 241),
            await refTable.findAt(refTestUri, 243),
            await refTable.findAt(refTestUri, 255),
            await refTable.findAt(refTestUri, 289),
            await refTable.findAt(refTestUri, 304),
            await refTable.findAt(refTestUri, 331),
            await refTable.findAt(refTestUri, 340),
            await refTable.findAt(refTestUri, 351),
        ];

        let refTestDoc = await phpDocTable.get(refTestUri);

        let defs: Symbol[] = [];
        for (let ref of refs) {
            let def: Symbol | null = null;

            if (ref !== null) {
                switch (ref.refKind) {
                    case RefKind.Class:
                        def = (await RefResolver.getClassSymbols(refTestDoc, ref)).shift();
                        break;
                    case RefKind.Function:
                        def = (await RefResolver.getFuncSymbols(refTestDoc, ref)).shift();
                        break;
                    case RefKind.Method:
                        def = (await RefResolver.getMethodSymbols(refTestDoc, ref)).shift();
                        break;
                    case RefKind.Property:
                        def = (await RefResolver.getPropSymbols(refTestDoc, ref)).shift();
                        break;
                    case RefKind.ClassConst:
                        def = (await RefResolver.getClassConstSymbols(refTestDoc, ref)).shift();
                        break;
                    case RefKind.ClassTypeDesignator:
                        let constructors = (await RefResolver.getMethodSymbols(refTestDoc, ref));

                        if (constructors.length === 0) {
                            def = (await RefResolver.getClassSymbols(refTestDoc, ref)).shift();
                        } else {
                            def = constructors.shift();
                        }
                        break;
                    case RefKind.ConstantAccess:
                        def = (await RefResolver.getConstSymbols(refTestDoc, ref)).shift();
                        break;
                }
            }

            defs.push(def);
        }

        expect(refs).toMatchSnapshot();
        expect(defs).toMatchSnapshot();

        await App.clearCache();
    });
});