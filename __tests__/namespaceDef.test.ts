import "reflect-metadata";
import { SymbolParser } from "../src/symbol/symbolParser";
import { PhpDocument } from "../src/symbol/phpDocument";
import { pathToUri } from "../src/util/uri";
import * as path from 'path';
import * as fs from 'fs';
import { Parser } from "php7parser";
import { Traverser } from "../src/traverser";
import { App } from "../src/app";
import { getDebugDir, getCaseDir } from "../src/testHelper";
import { Indexer, PhpFileInfo } from "../src/index/indexer";
import { PhpDocumentTable } from "../src/storage/table/phpDoc";

beforeAll(() => {
    App.init(path.join(getDebugDir(), 'storage'));
});

beforeEach(async () => {
    await App.clearCache();
});

afterAll(async () => {
    await App.shutdown();
});

describe('namespaceDef', () => {
    it('should assign namespace to phpDocument', () => {
        let workspaceDir = path.resolve(__dirname, '..', 'case', 'namespaceDef');
        let files = fs.readdirSync(workspaceDir);
        let treeTraverser = new Traverser();

        for (let file of files) {
            let filePath = path.join(workspaceDir, file);
            let fileUri = pathToUri(filePath);

            if (file.endsWith('.php')) {
                const fileContent = fs.readFileSync(filePath).toString();
                let symbolParser = new SymbolParser(new PhpDocument(
                    fileUri,
                    fileContent
                ));
                let parseTree = Parser.parse(fileContent);

                treeTraverser.traverse(parseTree, [
                    symbolParser
                ]);

                expect(symbolParser.getPhpDoc().toObject()).toMatchSnapshot();
            }
        }
    });

    it('returns import table', async () => {
        const indexer = App.get<Indexer>(Indexer);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        const filePath = path.join(getCaseDir(), 'namespaceDef', 'import_table.php');
        const fileUri = pathToUri(filePath);

        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(filePath));

        let phpDoc = await phpDocTable.get(fileUri);

        if (phpDoc === null) {
            return;
        }

        expect(phpDoc.importTable).toMatchSnapshot();
    });
});