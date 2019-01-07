import { App } from "../src/app";
import * as path from "path";
import { getDebugDir, getCaseDir } from "../src/testHelper";
import { FunctionTable } from "../src/storage/table/function";
import { Indexer } from "../src/index/indexer";
import { PhpDocumentTable } from "../src/storage/table/phpDoc";
import { pathToUri } from "../src/util/uri";
import { ReferenceTable } from "../src/storage/table/referenceTable";

beforeEach(() => {
    App.init(path.join(getDebugDir(), 'storage'));
});

afterEach(async () => {
    await App.clearCache();
});

describe('completion', () => {
    it('directly query function table', async () => {
        const indexer = App.get<Indexer>(Indexer);
        const funcTable = App.get<FunctionTable>(FunctionTable);
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        const testFilePath = path.join(getCaseDir(), 'completion', 'function.php');
        const testFileUri = pathToUri(testFilePath);

        await indexer.syncFileSystem(path.join(getCaseDir(), 'global_symbols.php'));
        await indexer.syncFileSystem(path.join(getCaseDir(), 'function_declare.php'));
        await indexer.syncFileSystem(testFilePath);

        let ref = await refTable.findAt(testFileUri, 18);
        let phpDoc = await phpDocTable.get(testFileUri);

        if (phpDoc === null || ref === null) {
            return;
        }

        console.log({
            ref,
            keyword: ref.type.toString()
        });
        console.log(await funcTable.search(phpDoc, ref.type.toString()));
    });
});