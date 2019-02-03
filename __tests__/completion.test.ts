import { App } from "../src/app";
import * as path from "path";
import { getDebugDir, getCaseDir, dumpAstToDebug } from "../src/testHelper";
import { Indexer, PhpFileInfo } from "../src/index/indexer";
import { PhpDocumentTable } from "../src/storage/table/phpDoc";
import { pathToUri } from "../src/util/uri";
import { ReferenceTable } from "../src/storage/table/reference";
import { RefResolver } from "../src/handler/refResolver";

interface CompletionTestCase {
    path: string;
    offset: number;
}

async function testCompletions(definitionFiles: string[], testCases: CompletionTestCase[]) {
    const indexer = App.get<Indexer>(Indexer);
    const refTable = App.get<ReferenceTable>(ReferenceTable);
    const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

    for(let definitionFile of definitionFiles) {
        await indexer.syncFileSystem(
            await PhpFileInfo.createFileInfo(path.join(getCaseDir(), definitionFile))
        );
    }

    for (let testCase of testCases) {
        await indexer.syncFileSystem(await PhpFileInfo.createFileInfo(testCase.path));

        const uri = pathToUri(testCase.path);
        const phpDoc = await phpDocTable.get(uri);
        const ref = await refTable.findAt(uri, testCase.offset);

        if (phpDoc === null || ref === null) {
            continue;
        }

        let symbols = await RefResolver.searchSymbolsForReference(phpDoc, ref, testCase.offset);
        for (let i = 0; i < symbols.length; i++) {
            symbols[i] = symbols[i].toObject();
        }

        expect(symbols).toMatchSnapshot();
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

describe('completion', () => {
    it('completion for constant access references', async () => {
        await testCompletions([
            'global_symbols.php', 'function_declare.php'
        ], [
            { path: path.join(getCaseDir(), 'completion', 'function.php'), offset: 18 },
            { path: path.join(getCaseDir(), 'completion', 'constant.php'), offset: 14 },
            { path: path.join(getCaseDir(), 'completion', 'constant.php'), offset: 29 },
            { path: path.join(getCaseDir(), 'completion', 'constant.php'), offset: 37 },
        ]);
    });

    it('completion for scoped member references', async () => {
        await testCompletions(['global_symbols.php'], [
            { path: path.join(getCaseDir(), 'completion', 'scopedMember.php'), offset: 18 },
            { path: path.join(getCaseDir(), 'completion', 'scopedMember.php'), offset: 33 },
            { path: path.join(getCaseDir(), 'completion', 'scopedMember.php'), offset: 47 },
            { path: path.join(getCaseDir(), 'completion', 'scopedMember.php'), offset: 62 },
            { path: path.join(getCaseDir(), 'completion', 'scopedMember.php'), offset: 78 },
        ]);
    });

    it('completion for current scoped member references', async () => {
        await testCompletions(['global_symbols.php'], [
            { path: path.join(getCaseDir(), 'completion', 'currentScopedMember.php'), offset: 176 },
            { path: path.join(getCaseDir(), 'completion', 'currentScopedMember.php'), offset: 244 },
            { path: path.join(getCaseDir(), 'completion', 'currentScopedMember.php'), offset: 309 },
        ]);
    });

    it('completion for variables', async() => {
        await testCompletions(['global_symbols.php'], [
            { path: path.join(getCaseDir(), 'completion', 'variables.php'), offset: 27 },
            { path: path.join(getCaseDir(), 'completion', 'variables.php'), offset: 94 },
            { path: path.join(getCaseDir(), 'completion', 'variables.php'), offset: 103 },
        ]);
    });

    it('variable arrow completion', async() => {
        await testCompletions(['global_symbols.php'], [
            { path: path.join(getCaseDir(), 'completion', 'arrow1.php'), offset: 44 },
        ]);
    });
});