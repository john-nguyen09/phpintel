import { App } from "../src/app";
import * as path from "path";
import { getDebugDir, getCaseDir, dumpAstToDebug } from "../src/testHelper";
import { Indexer, PhpFileInfo } from "../src/index/indexer";
import { PhpDocumentTable } from "../src/storage/table/phpDoc";
import { pathToUri } from "../src/util/uri";
import { SignatureHelpProvider } from "../src/handler/signatureHelp";

interface SignatureHelpTestCase {
    testFile: string;
    line: number;
    character: number;
}

async function testSignatureHelp(definitionFiles: string[], testCases: SignatureHelpTestCase[]) {
    const indexer = App.get<Indexer>(Indexer);
    const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

    for (let definitionFile of definitionFiles) {
        await indexer.syncFileSystem(
            await PhpFileInfo.createFileInfo(path.join(getCaseDir(), definitionFile))
        );
    }

    for (const testCase of testCases) {
        const testFileUri = pathToUri(testCase.testFile);

        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(testCase.testFile));

        const phpDoc = await phpDocTable.get(testFileUri);

        if (phpDoc === null) {
            return;
        }

        const signatureHelp = await SignatureHelpProvider.provide({
            position: { line: testCase.line, character: testCase.character },
            textDocument: {
                uri: testFileUri
            }
        });

        expect(signatureHelp).toMatchSnapshot();
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

describe('provide signature help', () => {
    it('shows list of parameters', async () => {
        await testSignatureHelp([
            'function_declare.php',
        ], [
            { testFile: path.join(getCaseDir(), 'signatureHelp', 'function.php'), line: 2, character: 16 },
            { testFile: path.join(getCaseDir(), 'signatureHelp', 'function.php'), line: 2, character: 21 },
            { testFile: path.join(getCaseDir(), 'signatureHelp', 'method.php'), line: 4, character: 23 },
            { testFile: path.join(getCaseDir(), 'signatureHelp', 'method.php'), line: 12, character: 26 },
        ]);
    });

    it('parameters for type designator', async () => {
        await testSignatureHelp([
            'class_methods.php',
        ], [
            { testFile: path.join(getCaseDir(), 'signatureHelp', 'type_designator.php'), line: 2, character: 20 },
            { testFile: path.join(getCaseDir(), 'signatureHelp', 'type_designator.php'), line: 4, character: 23 },
        ]);
    });
});