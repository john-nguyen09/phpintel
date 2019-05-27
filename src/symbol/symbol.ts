import { nodeText } from '../util/parseTree';
import { Token } from 'php7parser';
import { PhpDocument } from './phpDocument';
import { DocBlock } from './docBlock';
import { TypeName } from '../type/name';
import { TokenKind } from '../util/parser';
import { isFieldGetter, FieldGetter } from './fieldGetter';
import { createObject } from '../util/genericObject';
import { Location } from './meta/location';
import { ImportTable } from '../type/importTable';
import { toRelative } from '../util/uri';
import { Class } from './class/class';
import { Interface } from './interface/interface';

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
            } else if (key === 'uri') {
                object[key] = toRelative(value);
            } else {
                object[key] = this.createNewObject(value);
            }
        }

        return object;
    }

    private assignFieldGetter(object: any, fieldGetter: FieldGetter) {
        let fields = fieldGetter.getFields();

        for (let key of fields) {
            let value: any = (<any>fieldGetter)[key];

            if (key === 'uri') {
                object[key] = toRelative(value);
            } else {
                object[key] = this.createNewObject(value);
            }
        }
    }

    private createNewObject(currObj: any): any {
        if (currObj !== null && typeof currObj == 'object') {
            let newObj: any;

            if ('toObject' in currObj && typeof currObj.toObject == 'function') {
                newObj = currObj.toObject();
            } else if (isFieldGetter(currObj)) {
                newObj = createObject(currObj.constructor);
                this.assignFieldGetter(newObj, currObj);
            } else if (Array.isArray(currObj)) {
                return currObj.map((obj) => {
                    return this.createNewObject(obj);
                });
            } else {
                if ('uri' in currObj) {
                    currObj.uri = toRelative(currObj.uri);
                }

                newObj = currObj;
            }

            return newObj;
        }

        return currObj;
    }
}

export class TokenSymbol extends Symbol {
    public node: Token;

    public text: string;

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
    setScopeClass(scopeClass: Class | Interface): void;
}

export interface HasScope {
    scope: TypeName | null;
}

export interface NamedSymbol {
    name: TypeName;
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
    return 'setScopeClass' in symbol;
}

export function isNamedSymbol(symbol: Symbol): symbol is (Symbol & NamedSymbol) {
    return 'name' in symbol && (<any>symbol).name instanceof TypeName;
}

export function needsNameResolve(symbol: Symbol): symbol is (Symbol & NameResolvable) {
    return 'resolveName' in symbol && typeof (<any>symbol).resolveName == 'function';
}

export function isLocatable(symbol: Symbol): symbol is (Symbol & Locatable) {
    return 'location' in symbol;
}

