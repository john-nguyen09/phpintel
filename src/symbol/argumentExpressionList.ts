import { Symbol, TokenSymbol, Consumer } from "./symbol";
import { TokenKind } from "../util/parser";
import { Location } from "./meta/location";
import { Range } from "./meta/range";
import { FieldGetter } from "./fieldGetter";
import { Reference, RefKind } from "./reference";
import { TypeName } from "../type/name";
import { TypeComposite } from "../type/composite";

export class ArgumentExpressionList extends Symbol implements Consumer, FieldGetter, Reference {
    public readonly refKind = RefKind.ArgumentList;
    public arguments: Symbol[] = [];
    public location: Location = {};
    public type: TypeName = new TypeName('');
    public scope: TypeName | TypeComposite | null = null;

    public commaOffsets: number[] = [];

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

    getFields() {
        return [
            'arguments',
            'location'
        ];
    }
}