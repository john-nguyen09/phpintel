import { Symbol, TokenSymbol } from "../symbol";
import { TokenType } from "php7parser";
import { Expression } from "./expression";

export class AdditiveExpression extends Expression {
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
                    value.type == TokenType.StringLiteral ||
                    // Or a string concat token
                    (
                        (value.type == TokenType.Whitespace || value.type == TokenType.Dot) &&
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

    get type(): string {
        if (this.valueSymbols.length >= 1) {
            let firstValue = this.valueSymbols[0];

            return this.getType(firstValue);
        }

        return '';
    }

    private concatStrings(stringValues: TokenSymbol[]): string {
        if (stringValues.length == 0) {
            return '';
        }

        let quote = stringValues[0].text.slice(0, 1);
        let result: string = '';
        let isConcatString = stringValues[stringValues.length - 1].type == TokenType.StringLiteral;

        if (isConcatString) {
            result += quote;
        }

        for (let stringValue of stringValues) {
            if (
                isConcatString &&
                (stringValue.type == TokenType.Dot || stringValue.type == TokenType.Whitespace)
            ) {
                continue;
            }

            if (isConcatString && stringValue.type == TokenType.StringLiteral) {
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