import { App } from "../src/app";
import * as path from "path";
import { getDebugDir, getCaseDir, dumpAstToDebug } from "../src/testHelper";
import { Indexer, PhpFileInfo } from "../src/index/indexer";
import { ReferenceTable } from "../src/storage/table/reference";
import { PhpDocumentTable } from "../src/storage/table/phpDoc";
import { pathToUri } from "../src/util/uri";

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
        const indexer = App.get<Indexer>(Indexer);
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const definitionFiles = [
            'global_symbols.php',
        ];

        for (let definitionFile of definitionFiles) {
            await indexer.syncFileSystem(
                await PhpFileInfo.createFileInfo(path.join(getCaseDir(), definitionFile))
            );
        }

        const testFile = path.join(getCaseDir(), 'signatureHelp', 'function.php');
        const testFileUri = pathToUri(testFile);

        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(testFile));

        const ref = await refTable.findAt(testFileUri, 24);
        const phpDoc = await phpDocTable.get(testFileUri);

        if (phpDoc === null) {
            return;
        }

        dumpAstToDebug(path.basename(testFile) + '.ast.json', phpDoc.getTree());

        console.log(ref);
    });
});