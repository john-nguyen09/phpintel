import { WordSeparator } from "../src/storage/table/index/completionIndex/wordSeparator"

interface TestCase {
    input: string;
    output: string[];
}

const TEST_CASES: TestCase[] = [
    { input: '_test_func', output: [ '_test_func', 'test_func', 'func' ] },
    { input: 'test_function1', output: [ 'test_function1', 'function1' ] },
    { input: 'this_is_a_long_function_name', output: [ 'this_is_a_long_function_name', 'is_a_long_function_name', 'a_long_function_name', 'long_function_name', 'function_name', 'name' ] },
    { input: 'test__function1', output: [ 'test__function1', 'function1' ] },
    { input: '__test_function1', output: [ '__test_function1', 'test_function1', 'function1' ] },
    { input: '__test__function1', output: [ '__test__function1', 'test__function1', 'function1' ] },
    { input: 'testFunction', output: [ 'testFunction', 'Function' ] },
    { input: 'aLongCasingFunction', output: [ 'aLongCasingFunction', 'LongCasingFunction', 'CasingFunction', 'Function' ] },
    { input: 'NguyenPhuocHoangThuan', output: [ 'NguyenPhuocHoangThuan', 'PhuocHoangThuan', 'HoangThuan', 'Thuan' ] },
    { input: 'TESTFunction', output: [ 'TESTFunction', 'Function' ] },
    { input: 'TESTMyFunction', output: [ 'TESTMyFunction', 'MyFunction', 'Function' ] },
    { input: 'thisIsFunctionThatIsTESTFunction', output: [ 'thisIsFunctionThatIsTESTFunction', 'IsFunctionThatIsTESTFunction', 'FunctionThatIsTESTFunction', 'ThatIsTESTFunction', 'IsTESTFunction', 'TESTFunction', 'Function' ] },
    { input: 'thisIsAFunctionThatIsATESTFunction', output: [ 'thisIsAFunctionThatIsATESTFunction', 'IsAFunctionThatIsATESTFunction', 'AFunctionThatIsATESTFunction', 'FunctionThatIsATESTFunction', 'ThatIsATESTFunction', 'IsATESTFunction', 'ATESTFunction', 'Function' ] },
];

beforeAll(() => {
    for (let testCase of TEST_CASES) {
        testCase.output = testCase.output.sort();
    }
});

describe('word separator', () => {
    it('tests word separation', () => {
        for (let testCase of TEST_CASES) {
            let output = WordSeparator.getTokens(testCase.input);
            output = output.sort();

            expect(output).toEqual(testCase.output);
        }
    });
});