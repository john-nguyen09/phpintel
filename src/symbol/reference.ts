import { TypeName } from "../type/name";
import { TypeComposite } from "../type/composite";
import { Symbol } from "./symbol";
import { Location } from "./meta/location";

export enum RefKind {
    FunctionCall = 1,
    Variable = 2,
    Parameter = 3,
    TypeDeclaration = 4,
    Constant = 5,
    ConstantAccess = 6,
    DefineConstant = 7,
    ClassTypeDesignator = 8
}

export interface Reference {
    refKind: RefKind;
    type: TypeName | TypeComposite;
    location: Location;
}

export function isReference(symbol: Symbol): symbol is (Symbol & Reference) {
    return 'refKind' in symbol && 'type' in symbol && 'location' in symbol;
}

export function refKindToString(refKind: RefKind): string {
    switch (refKind) {
        case RefKind.FunctionCall:
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
        case RefKind.DefineConstant:
            return 'DefineConstant';
        case RefKind.ClassTypeDesignator:
            return 'ClassTypeDesignator';
    }

    return '';
}
