import { Symbol, TokenSymbol, Consumer } from "./symbol";
import { TokenKind } from "../util/parser";
import { Location } from "./meta/location";
import { Range } from "./meta/range";
import { FieldGetter } from "./fieldGetter";
import { TypeName } from "../type/name";
import { MethodCall } from "./function/methodCall";
import { FunctionCall } from "./function/functionCall";
import { MethodCallExpression } from "./type/methodCallExpression";
import { TypeComposite } from "../type/composite";
import { ObjectCreationExpression } from "./type/objectCreationExpression";

export type CallExpression = MethodCall | FunctionCall | MethodCallExpression | ObjectCreationExpression;

export class ArgumentExpressionList extends Symbol implements Consumer, FieldGetter {
    public arguments: Symbol[] = [];
    public location: Location = {};

    public commaOffsets: number[] = [];

    private _callExpression: CallExpression | null = null;
    private _type: TypeName | null = null;
    private _scope: TypeComposite | TypeName | null = null;

    constructor(callExpression?: CallExpression) {
        super();
        if (callExpression !== undefined) {
            this._callExpression = callExpression;
        }
    }

    consume(other: Symbol) {
        let isCommaOrWhitespace = false;

        if (other instanceof TokenSymbol) {
            if (other.type === TokenKind.Comma) {
                isCommaOrWhitespace = true;
                this.commaOffsets.push(other.node.offset);
            } else if (other.type === TokenKind.Whitespace) {
                isCommaOrWhitespace = true;
            }
        }

        if (!isCommaOrWhitespace) {
            this.arguments.push(other);
        }

        return true;
    }

    get ranges(): Range[] {
        if (this.location.range === undefined) {
            return [];
        }

        const ranges: Range[] = [];
        let lastStart = this.location.range.start;

        for (const offset of this.commaOffsets) {
            ranges.push({
                start: lastStart,
                end: offset
            });
            lastStart = offset + 1;
        }
        ranges.push({
            start: lastStart,
            end: this.location.range.end
        });

        return ranges;
    }

    get type(): TypeName {
        if (this._callExpression === null) {
            if (this._type === null) {
                return new TypeName('');
            }

            return this._type;
        }

        return this._callExpression.type;
    }

    set type(value: TypeName) {
        this._type = value;
    }

    get scope(): TypeComposite | TypeName | null {
        if (this._callExpression === null) {
            if (this._scope === null) {
                return null;
            }

            return this._scope;
        }

        if (this._callExpression instanceof FunctionCall) {
            return null;
        }

        return this._callExpression.scope;
    }

    set scope(value: TypeComposite | TypeName | null) {
        this._scope = value;
    }

    getFields() {
        return [
            'arguments',
            'location'
        ];
    }
}