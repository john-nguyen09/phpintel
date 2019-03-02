import * as path from "path";
import { Position } from "./meta";
import { PhpFile } from "./phpFile";
import { FileInfo } from "./indexer";
import { Expression } from "./expression";

describe('Expression', () => {
    it('should return expressions', async () => {
        const caseDir = path.join(__dirname, '..', 'case');
        const documents: { path: string, pos: Position }[] = [
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 1, column: 19 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 3, column: 0 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 5, column: 24 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 7, column: 24 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 11, column: 14 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 11, column: 31 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 12, column: 40 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 13, column: 30 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 14, column: 19 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 24, column: 16 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 25, column: 19 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 26, column: 9 } },
            { path: path.join(caseDir, 'reference', 'references.php'), pos: { row: 27, column: 21 } },
        ];

        for (const document of documents) {
            const phpFile = await PhpFile.create(await FileInfo.create(document.path));
            const expression = new Expression(phpFile, document.pos);

            console.log({
                type: expression.type,
                name: expression.name,
                nameRange: expression.nameRange,
                scope: expression.scope,
                scopeRange: expression.scopeRange,
            });
        }
    });
});