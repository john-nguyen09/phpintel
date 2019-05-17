import { App } from "../src/app";
import * as path from "path";
import { getDebugDir, getCaseDir } from "../src/testHelper";
import { Indexer, PhpFileInfo } from "../src/index/indexer";
import { DocumentSymbolProvider } from "../src/handler/documentSymbol";
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

describe('documentSymbol', () => {
    it('global_symbols.php', async() => {
        const indexer = App.get<Indexer>(Indexer);
        const filePath = path.join(getCaseDir(), 'global_symbols.php');
        
        await indexer.syncFileSystem(
            await PhpFileInfo.createFileInfo(filePath)
        );

        const symbols = await DocumentSymbolProvider.provide({
            textDocument: {
                uri: pathToUri(filePath)
            }
        });

        console.log(symbols);
    });
});