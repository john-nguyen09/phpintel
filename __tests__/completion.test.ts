import { App } from "../src/app";
import * as path from "path";
import { getDebugDir, getCaseDir } from "../src/testHelper";
import { FunctionTable } from "../src/storage/table/function";
import { Indexer, PhpFileInfo } from "../src/index/indexer";
import { PhpDocumentTable } from "../src/storage/table/phpDoc";
import { pathToUri } from "../src/util/uri";
import { ReferenceTable } from "../src/storage/table/referenceTable";
import { RefResolver } from "../src/handler/refResolver";

interface CompletionTestCase {
    path: string;
    offset: number;
}

beforeAll(() => {
    App.init(path.join(getDebugDir(), 'storage'));
});

beforeEach(async () => {
    await App.clearCache();
});

afterAll(async() => {
    await App.shutdown();
});

describe('completion', () => {
    it('query from RefResolver', async () => {
        const indexer = App.get<Indexer>(Indexer);
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        const testCases: CompletionTestCase[] = [
            { path: path.join(getCaseDir(), 'completion', 'function.php'), offset: 18 },
            { path: path.join(getCaseDir(), 'completion', 'constant.php'), offset: 14 },
            { path: path.join(getCaseDir(), 'completion', 'constant.php'), offset: 29 },
            { path: path.join(getCaseDir(), 'completion', 'constant.php'), offset: 37 },
        ];

        await indexer.syncFileSystem(
            await PhpFileInfo.createFileInfo(path.join(getCaseDir(), 'global_symbols.php'))
        );
        await indexer.syncFileSystem(
            await PhpFileInfo.createFileInfo(path.join(getCaseDir(), 'function_declare.php'))
        );
        for (let testCase of testCases) {
            await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(testCase.path));

            let uri = pathToUri(testCase.path);
            let phpDoc = await phpDocTable.get(uri);
            let ref = await refTable.findAt(uri, testCase.offset);

            if (phpDoc === null || ref === null) {
                continue;
            }

            let symbols = await RefResolver.searchSymbolsForReference(phpDoc, ref);
            for (let i = 0; i < symbols.length; i++) {
                symbols[i] = symbols[i].toObject();
            }

            expect(symbols).toMatchSnapshot();
        }
    });
});