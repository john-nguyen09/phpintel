import { TreeNode, nodeText } from '../util/parseTree';
import { Token } from 'php7parser';
import { PhpDocument } from './phpDocument';
import { nonenumerable } from '../util/decorator';
import { DocBlock } from './docBlock';
import { TypeName } from '../type/name';
import { TypeComposite } from '../type/composite';
import { TokenKind } from '../util/parser';
import { isFieldGetter, FieldGetter } from './fieldGetter';
import { createObject } from '../util/genericObject';
import { Location } from './meta/location';

export abstract class Symbol {
    @nonenumerable
    public node: TreeNode | null;

    @nonenumerable
    public doc: PhpDocument | null;

    constructor(node: TreeNode | null, doc: PhpDocument | null) {
        this.node = node;
        this.doc = doc;
    }

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
        super(token, doc);

        this.type = <number>token.tokenType;
        this.text = nodeText(token, doc.textDocument.text);
    }
}

export abstract class TransformSymbol extends Symbol {
    abstract realSymbol: Symbol;
}

export abstract class CollectionSymbol extends Symbol {
    abstract realSymbols: Symbol[];
}

export interface Consumer {
    consume(other: Symbol): boolean;
}

export interface DocBlockConsumer {
    consumeDocBlock(docBlock: DocBlock): void;
}

export interface Reference {
    type: TypeName | TypeComposite;
}

export interface ScopeMember {
    scope: string;
}

export function isTransform(symbol: Symbol): symbol is TransformSymbol {
    return symbol != null && 'realSymbol' in symbol;
}

export function isCollection(symbol: Symbol): symbol is CollectionSymbol {
    return symbol != null && 'realSymbols' in symbol;
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