import { nodeText } from '../util/parseTree';
import { Token } from 'php7parser';
import { PhpDocument } from './phpDocument';
import { nonenumerable } from '../util/decorator';
import { DocBlock } from './docBlock';
import { TypeName } from '../type/name';
import { TokenKind } from '../util/parser';
import { isFieldGetter, FieldGetter } from './fieldGetter';
import { createObject } from '../util/genericObject';
import { Location } from './meta/location';
import { ImportTable } from '../type/importTable';

export abstract class Symbol {
    toObject(): any {
        let instance = this;
        let object: any = createObject((<any>instance).constructor);

        if (isFieldGetter(instance)) {
            this.assignFieldGetter(object, instance);

            return object;
        }

        for (let key in this) {
            let value: any = this[key];

            if (value instanceof Array) {
                object[key] = [];

                for (let child of value) {
                    object[key].push(this.createNewObject(child));
                }
            } else {
                object[key] = this.createNewObject(value);
            }
        }

        return object;
    }

    private assignFieldGetter(object: any, fieldGetter: FieldGetter) {
        let fields = fieldGetter.getFields();

        for (let key of fields) {
            object[key] = (<any>fieldGetter)[key];
        }
    }

    private createNewObject(currObj: any): any {
        if (currObj instanceof Object) {
            let newObj: any;

            if ('toObject' in currObj && typeof currObj.toObject == 'function') {
                newObj = currObj.toObject();
            } else if (isFieldGetter(currObj)) {
                newObj = createObject(currObj.constructor);
                this.assignFieldGetter(newObj, currObj);
            } else {
                newObj = currObj;
            }

            return newObj;
        }

        return currObj;
    }
}

export class TokenSymbol extends Symbol {
    @nonenumerable
    public node: Token;

    public text: string;

    @nonenumerable
    public type: TokenKind;

    constructor(token: Token, doc: PhpDocument) {
        super();

        this.node = token;
        this.type = <number>token.tokenType;
        this.text = nodeText(token, doc.text);
    }
}

export abstract class TransformSymbol extends Symbol {
    abstract realSymbol: Symbol;
}

export abstract class CollectionSymbol extends Symbol {
    abstract realSymbols: Symbol[];
    abstract isParentIncluded: boolean;
}

export interface Consumer {
    consume(other: Symbol): boolean;
}

export interface DocBlockConsumer {
    consumeDocBlock(docBlock: DocBlock): void;
}

export interface ScopeMember {
    scope: TypeName | null;
}

export interface NamedSymbol {
    getName(): string;
}

export interface Locatable {
    location: Location;
}

export interface NameResolvable {
    resolveName(importTable: ImportTable): void;
}

export function isTransform(symbol: Symbol): symbol is TransformSymbol {
    return symbol != null && 'realSymbol' in symbol;
}

export function isCollection(symbol: Symbol): symbol is CollectionSymbol {
    return symbol instanceof CollectionSymbol;
}

export function isConsumer(symbol: Symbol): symbol is (Symbol & Consumer) {
    return 'consume' in symbol;
}

export function isDocBlockConsumer(symbol: Symbol): symbol is (Symbol & DocBlockConsumer) {
    return 'consumeDocBlock' in symbol;
}

export function isScopeMember(symbol: Symbol): symbol is (Symbol & ScopeMember) {
    return 'scope' in symbol;
}

export function isNamedSymbol(symbol: Symbol): symbol is (Symbol & NamedSymbol) {
    return 'getName' in symbol && typeof (<any>symbol).getName == 'function';
}

export function needsNameResolve(symbol: Symbol): symbol is (Symbol & NameResolvable) {
    return 'resolveName' in symbol && typeof (<any>symbol).resolveName == 'function';
}

export function isLocatable(symbol: Symbol): symbol is (Symbol & Locatable) {
    return 'location' in symbol;
}

