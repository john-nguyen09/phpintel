import { App } from '../src/app';
import { Indexer, PhpFileInfo } from '../src/index/indexer';
import { getCaseDir, getDebugDir, dumpAstToDebug } from "../src/testHelper";
import * as path from "path";
import { ReferenceTable } from '../src/storage/table/reference';
import { pathToUri } from '../src/util/uri';
import { RefResolver } from "../src/handler/refResolver";
import { PhpDocumentTable } from '../src/storage/table/phpDoc';
import { RefKind } from '../src/symbol/reference';
import { Symbol } from '../src/symbol/symbol';

beforeAll(() => {
    App.init(path.join(getDebugDir(), 'storage'));
});

beforeEach(async () => {
    await App.clearCache();
});

afterAll(async () => {
    await App.shutdown();
});

describe('Testing functions around references', () => {
    it('should return the reference at the cursor', async () => {
        const indexer = App.get<Indexer>(Indexer);
        const caseDir = getCaseDir();
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const refTestFile = path.join(caseDir, 'reference', 'references.php');
        const testFile2 = path.join(caseDir, 'class_methods.php');

        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(testFile2));
        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(path.join(caseDir, 'class_constants.php')));
        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(path.join(caseDir, 'global_symbols.php')));
        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(path.join(caseDir, 'function_declare.php')));
        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(refTestFile));

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

        if (refTestDoc === null) {
            return;
        }

        let defs: Symbol[] = [];
        for (let ref of refs) {
            let def: Symbol | undefined = undefined;

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

            if (def !== undefined) {
                defs.push(def);
            }
        }

        expect(refs).toMatchSnapshot();
        expect(defs).toMatchSnapshot();
    });

    it('reference variable', async () => {
        const indexer = App.get<Indexer>(Indexer);
        const caseDir = getCaseDir();
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const refTestFile = path.join(caseDir, 'reference', 'references.php');
        let refTestUri = pathToUri(refTestFile);

        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(refTestFile));

        let variables = [
            await refTable.findAt(refTestUri, 376),
            await refTable.findAt(refTestUri, 418),
            await refTable.findAt(refTestUri, 437),
        ];

        expect(variables).toMatchSnapshot();

        // for (let variable of variables) {
        //     console.log(inspect(variable, {
        //         depth: 4,
        //         colors: true,
        //     }));
        // }
    });

    it('returns class constant ref before variable', async() => {
        const indexer = App.get<Indexer>(Indexer);
        const caseDir = getCaseDir();
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const refTestFile = path.join(caseDir, 'reference', 'scopedMemberBeforeVariable.php');
        const refTestUri = pathToUri(refTestFile);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(refTestFile));

        let phpDoc = await phpDocTable.get(refTestUri);
        let ref = await refTable.findAt(refTestUri, 20);
    });
});