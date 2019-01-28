import { TypeName } from "../type/name";
import { TypeComposite } from "../type/composite";
import { Symbol } from "./symbol";
import { Location } from "./meta/location";
import { Range } from "./meta/range";

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
}

export interface Reference {
    refName?: string;
    refKind: RefKind;
    type: TypeName | TypeComposite;
    location: Location;
    scope: TypeName | null;
    scopeRange?: Range;
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
