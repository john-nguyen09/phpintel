import { TypeName } from "../src/type/name";

interface TypeNameTestCase {
    name: string;
    isFqn: boolean;
}

const TEST_CASES: TypeNameTestCase[] = [
    { name: '\\Namespace\\ClassName', isFqn: true },
    { name: 'function1', isFqn: false },
    { name: '\\function1', isFqn: true },
    { name: 'COMPLETION_COMPLETE', isFqn: false },
    { name: '\\COMPLETION_COMPLETE', isFqn: true },
];

describe('', () => {
    it('returns whether a name is FQN', () => {
        for (let testCase of TEST_CASES) {
            expect(TypeName.isFqn(testCase.name), `${testCase.name} is failing`)
                .toEqual(testCase.isFqn);
        }
    });
});