import { CollectionSymbol, Symbol, Consumer, TokenSymbol } from "../symbol";
import { ClassRef } from "../class/classRef";
import { MethodCall } from "../function/methodCall";
import { nonenumerable } from "../../util/decorator";
import { TokenKind } from "../../util/parser";

export class MethodCallExpression extends CollectionSymbol implements Consumer {
    public isParentIncluded = true;
    public classRef: ClassRef = new ClassRef();
    public methodCall: MethodCall = new MethodCall();

    @nonenumerable
    private hasColonColon: boolean = false;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type == TokenKind.ColonColon) {
            this.hasColonColon = true;
            this.methodCall.scope = this.classRef.type;

            return true;
        }

        if (!this.hasColonColon) {
            return this.classRef.consume(other);
        } else {
            return this.methodCall.consume(other);
        }
    }

    get realSymbols(): Symbol[] {
        return [
            this.classRef,
            this.methodCall
        ]
    }
}