import { App } from '../src/app';
import { Indexer, PhpFileInfo } from '../src/index/indexer';
import { getCaseDir, getDebugDir, dumpAstToDebug } from "../src/testHelper";
import * as path from "path";
import { ReferenceTable } from '../src/storage/table/reference';
import { pathToUri } from '../src/util/uri';
import { RefResolver } from "../src/handler/refResolver";
import { PhpDocumentTable } from '../src/storage/table/phpDoc';
import { RefKind, Reference } from '../src/symbol/reference';
import { Symbol } from '../src/symbol/symbol';

interface ReferenceTestCase {
    definitionFiles: string[];
    testFile: string;
    startOffset: number;
    endOffset: number;
}

async function testRefAndDef(testCases: ReferenceTestCase[]) {
    const indexer = App.get<Indexer>(Indexer);
    const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
    const refTable = App.get<ReferenceTable>(ReferenceTable);

    for (const testCase of testCases) {
        await App.clearCache();

        for (const definitionFile of testCase.definitionFiles) {
            await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(definitionFile));
        }

        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(testCase.testFile));
        const testFileUri = pathToUri(testCase.testFile);

        const phpDoc = await phpDocTable.get(testFileUri);

        if (phpDoc === null) {
            continue;
        }

        let prevRef: Reference | null = null;
        let prevDefs: Symbol[] | null = null;
        for (let i = testCase.startOffset; i <= testCase.endOffset; i++) {
            const ref = await refTable.findAt(testFileUri, i);

            console.log({
                offset: i,
                ref
            });

            expect(ref).not.toEqual(null);
            if (ref === null) {
                break;
            }
            const thisDefs: Symbol[] = await RefResolver.getSymbolsByReference(phpDoc, ref);

            if (prevRef === null) {
                prevRef = ref;
            } else {
                expect(ref).toEqual(prevRef);
            }

            if (prevDefs === null) {
                prevDefs = thisDefs;
            } else {
                expect(thisDefs).toEqual(prevDefs);
            }
        }
    }
}

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
            await refTable.findAt(refTestUri, 469),
            await refTable.findAt(refTestUri, 481),
            await refTable.findAt(refTestUri, 493),
            await refTable.findAt(refTestUri, 505),
        ];

        let refTestDoc = await phpDocTable.get(refTestUri);

        if (refTestDoc === null) {
            return;
        }

        let defs: Symbol[] = [];
        for (let ref of refs) {
            let def: Symbol | undefined = undefined;

            if (ref !== null) {
                defs.push(...await RefResolver.getSymbolsByReference(refTestDoc, ref));
            }
        }

        expect(refs.map((ref) => {
            if (ref === null) {
                return ref;
            }

            return Reference.convertToTest(ref);
        })).toMatchSnapshot();
        expect(defs.map((def) => {
            return def.toObject();
        })).toMatchSnapshot();
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

        expect(variables.map((variable) => {
            if (variable === null) {
                return variable;
            }

            return Reference.convertToTest(variable);
        })).toMatchSnapshot();

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

    // it('temp ref test', async () => {
    //     await testRefAndDef([
    //         {
    //             definitionFiles: [path.join(getCaseDir(), 'moodleTestFile2.php')],
    //             testFile: path.join(getCaseDir(), 'moodleTestFile1.php'),
    //             startOffset: 425,
    //             endOffset: 444,
    //         }
    //     ]);
    // });
});