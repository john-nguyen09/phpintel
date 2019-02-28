import { PhpFile } from "./phpFile";
import * as path from "path";
import { TreeAnalyser } from "./treeAnalyser";

describe('test parsing php file', () => {
    it('should return a reflection of php file', async () => {
        const phpFile = await PhpFile.create(path.join(__dirname, '..', 'case', 'global_symbols.php'));

        TreeAnalyser.analyse(phpFile);
    });
});