import { PhpFile } from "./phpFile";
import * as path from "path";
import { TreeAnalyser } from "./treeAnalyser";
import { TreeTraverser } from "./treeTraverser";
import { Position } from "./meta";
import { FileInfo } from "./indexer";

describe('test parsing php file', () => {
    it('should return a reflection of php file', async () => {
        const phpFile = await PhpFile.create(
            await FileInfo.create(path.join(__dirname, '..', 'case', 'global_symbols.php'))
        );

        TreeAnalyser.analyse(phpFile);
    });
});