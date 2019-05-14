import { App } from '../src/app';
import { Indexer, PhpFileInfo } from '../src/index/indexer';
import { getCaseDir, getDebugDir } from "../src/testHelper";
import * as path from "path";
import { pathToUri } from '../src/util/uri';
import { RefResolver } from "../src/handler/refResolver";
import { PhpDocumentTable } from '../src/storage/table/phpDoc';
import { Reference } from '../src/symbol/reference';
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

    for (const testCase of testCases) {
        await App.clearCache();

        for (const definitionFile of testCase.definitionFiles) {
            await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(definitionFile));
        }

        const testFileUri = pathToUri(testCase.testFile);
        await indexer.open(testFileUri);

        const phpDoc = await phpDocTable.get(testFileUri);

        if (phpDoc === null) {
            continue;
        }

        let prevRef: Reference | null = null;
        let prevDefs: Symbol[] | null = null;
        for (let i = testCase.startOffset; i <= testCase.endOffset; i++) {
            const ref = await phpDoc.findRefAt(i);

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

        if (prevDefs === null) {
            return;
        }

        expect(prevDefs.map((def) => {
            return def.toObject();
        })).toMatchSnapshot();
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
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const refTestFile = path.join(caseDir, 'reference', 'references.php');
        const testFile2 = path.join(caseDir, 'class_methods.php');
        let refTestUri = pathToUri(refTestFile);

        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(testFile2));
        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(path.join(caseDir, 'class_constants.php')));
        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(path.join(caseDir, 'global_symbols.php')));
        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(path.join(caseDir, 'function_declare.php')));
        await indexer.open(refTestUri);

        let refTestDoc = await phpDocTable.get(refTestUri);

        if (refTestDoc === null) {
            return;
        }
        let refs = [
            await refTestDoc.findRefAt(21),
            await refTestDoc.findRefAt(24),
            await refTestDoc.findRefAt(30),
            await refTestDoc.findRefAt(37),
            await refTestDoc.findRefAt(51),
            await refTestDoc.findRefAt(42),
            await refTestDoc.findRefAt(226),
            await refTestDoc.findRefAt(241),
            await refTestDoc.findRefAt(243),
            await refTestDoc.findRefAt(255),
            await refTestDoc.findRefAt(289),
            await refTestDoc.findRefAt(304),
            await refTestDoc.findRefAt(331),
            await refTestDoc.findRefAt(340),
            await refTestDoc.findRefAt(351),
            await refTestDoc.findRefAt(469),
            await refTestDoc.findRefAt(481),
            await refTestDoc.findRefAt(493),
            await refTestDoc.findRefAt(505),
        ];

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
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const caseDir = getCaseDir();
        const refTestFile = path.join(caseDir, 'reference', 'references.php');
        const refTestUri = pathToUri(refTestFile);

        await indexer.open(refTestUri);
        const phpDoc = await phpDocTable.get(refTestUri);

        expect(phpDoc).not.toEqual(null);
        if (phpDoc === null) {
            return;
        }

        let variables = [
            await phpDoc.findRefAt(376),
            await phpDoc.findRefAt(418),
            await phpDoc.findRefAt(437),
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
        const refTestFile = path.join(caseDir, 'reference', 'scopedMemberBeforeVariable.php');
        const refTestUri = pathToUri(refTestFile);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        await indexer.open(refTestUri);

        const phpDoc = await phpDocTable.get(refTestUri);

        expect(phpDoc).not.toEqual(null);
        if (phpDoc === null) {
            return;
        }

        const ref = await phpDoc.findRefAt(20);

        expect(ref).toMatchSnapshot();
    });

    it('references for global variables', async () => {
        await testRefAndDef([
            {
                definitionFiles: [
                    path.join(getCaseDir(), 'global_variables.php'),
                    path.join(getCaseDir(), 'class_methods.php'),
                ],
                testFile: path.join(getCaseDir(), 'reference', 'global_variables.php'),
                startOffset: 14,
                endOffset: 23,
            },
            {
                definitionFiles: [
                    path.join(getCaseDir(), 'global_variables.php'),
                    path.join(getCaseDir(), 'class_methods.php'),
                ],
                testFile: path.join(getCaseDir(), 'reference', 'global_variables.php'),
                startOffset: 31,
                endOffset: 40,
            },
            {
                definitionFiles: [
                    path.join(getCaseDir(), 'global_variables.php'),
                    path.join(getCaseDir(), 'class_methods.php'),
                ],
                testFile: path.join(getCaseDir(), 'reference', 'global_variables.php'),
                startOffset: 52,
                endOffset: 61,
            },
            {
                definitionFiles: [
                    path.join(getCaseDir(), 'global_variables.php'),
                    path.join(getCaseDir(), 'class_methods.php'),
                ],
                testFile: path.join(getCaseDir(), 'reference', 'global_variables.php'),
                startOffset: 73,
                endOffset: 86,
            },
        ]);
    });

    it('temp ref test', async () => {
        await testRefAndDef([
            {
                definitionFiles: [],
                testFile: path.join(getCaseDir(), 'moodle_database.php'),
                startOffset: 482,
                endOffset: 482,
            }
        ]);
    });
});