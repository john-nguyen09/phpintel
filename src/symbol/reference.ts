import { TypeName } from "../type/name";
import { TypeComposite } from "../type/composite";
import { Symbol } from "./symbol";
import { Location } from "./meta/location";
import { Range } from "./meta/range";
import { toRelative } from "../util/uri";

export enum RefKind {
    Function = 1,
    Variable = 2,
    Parameter = 3,
    TypeDeclaration = 4,
    Constant = 5,
    ConstantAccess = 6,
    ClassTypeDesignator = 8,
    Class = 9,
    Method = 10,
    Property = 11,
    ClassConst = 12,
    ScopedAccess = 13,
    PropertyAccess = 14,
    MethodCall = 15,
    ArgumentList = 16,
}

export interface Reference {
    refName?: string;
    refKind: RefKind;
    type: TypeName | TypeComposite;
    location: Location;
    scope: TypeName | TypeComposite | null;
    scopeRange?: Range;
    memberLocation?: Location;
    ranges?: Range[];
}

export function isReference(symbol: Symbol): symbol is (Symbol & Reference) {
    return 'refKind' in symbol && 'type' in symbol && 'location' in symbol;
}

export function refKindToString(refKind: RefKind): string {
    switch (refKind) {
        case RefKind.Function:
            return 'FunctionCall';
        case RefKind.Variable:
            return 'Variable';
        case RefKind.Parameter:
            return 'Parameter';
        case RefKind.TypeDeclaration:
            return 'TypeDeclaration';
        case RefKind.Constant:
            return 'Constant';
        case RefKind.ConstantAccess:
            return 'ConstantAccess';
        case RefKind.ClassTypeDesignator:
            return 'ClassTypeDesignator';
    }

    return '';
}

export namespace Reference {
    export function convertToTest(ref: Reference): Reference {
        const testRef = Object.assign({}, ref);

        if (ref.location.uri !== undefined) {
            testRef.location.uri = toRelative(ref.location.uri);
        }
        if (
            ref.memberLocation !== undefined &&
            ref.memberLocation.uri !== undefined &&
            testRef.memberLocation !== undefined
        ) {
            testRef.memberLocation.uri = toRelative(ref.memberLocation.uri);
        }

        return testRef;
    }
}
