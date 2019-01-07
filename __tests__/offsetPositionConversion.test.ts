import { PhpDocument } from "../src/symbol/phpDocument";
import * as path from "path";
import { getCaseDir } from "../src/testHelper";
import { pathToUri } from "../src/util/uri";
import * as fs from "fs";

interface ConversionCase {
    path: string;
    offset: number;
    line: number;
    character: number;
}

const TEST_CASES: ConversionCase[] = [
    { path: path.join(getCaseDir(), 'bigFile.php'), offset: 3691, line: 91, character: 0 },
    { path: path.join(getCaseDir(), 'bigFile.php'), offset: 3692, line: 91, character: 1 },
    { path: path.join(getCaseDir(), 'bigFile.php'), offset: 2962, line: 77, character: 20 },
    { path: path.join(getCaseDir(), 'bigFile.php'), offset: 0, line: 0, character: 0 },
    { path: path.join(getCaseDir(), 'moodleTestFile1.php'), offset: 39183, line: 914, character: 1 },
];

describe('Test offset and position conversion', () => {
    it('should return correct offset', () => {
        for (let testCase of TEST_CASES) {
            let fileUri = pathToUri(testCase.path);
            let phpDoc = new PhpDocument(fileUri, fs.readFileSync(testCase.path).toString());

            expect(phpDoc.getOffset(testCase.line, testCase.character)).toEqual(testCase.offset);
        }
    });

    it('should return correct position', () => {
        for (let testCase of TEST_CASES) {
            let fileUri = pathToUri(testCase.path);
            let phpDoc = new PhpDocument(fileUri, fs.readFileSync(testCase.path).toString());
            let position = phpDoc.getPosition(testCase.offset);

            expect({
                line: position.line,
                character: position.character
            }).toEqual({
                line: testCase.line,
                character: testCase.character
            });
        }
    });
});