import { CollectionSymbol, Consumer, Symbol, TokenSymbol } from "../symbol";
import { ClassRef } from "../class/classRef";
import { ClassConstRef } from "../constant/classConstRef";
import { nonenumerable } from "../../util/decorator";
import { TokenKind } from "../../util/parser";

export class ClassConstRefExpression extends CollectionSymbol implements Consumer {
    public classRef: ClassRef = new ClassRef();
    public classConstRef: ClassConstRef = new ClassConstRef();

    @nonenumerable
    private hasColonColon: boolean = false;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type === TokenKind.ColonColon) {
            this.hasColonColon = true;
            this.classConstRef.scope = this.classRef.type;

            return true;
        }

        if (!this.hasColonColon) {
            return this.classRef.consume(other);
        } else {
            return this.classConstRef.consume(other);
        }
    }

    get realSymbols(): Symbol[] {
        return [
            this.classRef,
            this.classConstRef
        ]
    }
}