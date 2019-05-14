import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { Expression } from "./expression";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";
import { TypeComposite } from "../../type/composite";

export class AdditiveExpression extends Expression implements Consumer {
    protected valueSymbols: Symbol[] = [];

    consume(other: Symbol) {
        this.valueSymbols.push(other);

        return true;
    }

    get value() {
        let values: string[] = [];
        let stringValues: TokenSymbol[] = [];

        for (let value of this.valueSymbols) {
            if (value instanceof TokenSymbol) {
                if (
                    // This is a string
                    value.type == TokenKind.StringLiteral ||
                    // Or a string concat token
                    (
                        (value.type == TokenKind.Whitespace || value.type == TokenKind.Dot) &&
                        stringValues.length != 0
                    )
                ) {
                    stringValues.push(value);
                    continue;
                }

                values.push(this.concatStrings(stringValues));
                stringValues = [];

                values.push(value.text);
            } else {
                values.push(this.concatStrings(stringValues));
                stringValues = [];

                values.push(this.getValue(value));
            }
        }

        if (stringValues.length > 0) {
            values.push(this.concatStrings(stringValues));
            stringValues = [];
        }

        return values.join('');
    }

    get type(): TypeComposite {
        if (this.valueSymbols.length >= 1) {
            let firstValue = this.valueSymbols[0];

            const type = this.getType(firstValue);

            if (type instanceof TypeName) {
                return type;
            }
        }

        return new TypeComposite();
    }

    private concatStrings(stringValues: TokenSymbol[]): string {
        if (stringValues.length == 0) {
            return '';
        }

        let quote = stringValues[0].text.slice(0, 1);
        let result: string = '';
        let isConcatString = stringValues[stringValues.length - 1].type == TokenKind.StringLiteral;

        if (isConcatString) {
            result += quote;
        }

        for (let stringValue of stringValues) {
            if (
                isConcatString &&
                (stringValue.type == TokenKind.Dot || stringValue.type == TokenKind.Whitespace)
            ) {
                continue;
            }

            if (isConcatString && stringValue.type == TokenKind.StringLiteral) {
                result += stringValue.text.slice(1, -1);
            } else {
                result += stringValue.text;
            }
        }

        if (isConcatString) {
            result += quote;
        }

        return result;
    }
}